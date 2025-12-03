package websocket

import (
	"fmt"
	"log"

	"github.com/gin-gonic/gin"

	"context"
	"encoding/json"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"

	"chinese-chess-backend/database"
	"chinese-chess-backend/dto"
	"chinese-chess-backend/dto/room"
	dtouser "chinese-chess-backend/dto/user"
	
	modeluser "chinese-chess-backend/model/user"
	"chinese-chess-backend/utils"
	"slices"
)

const (
	HeartbeatInterval = 5 * time.Second  // 发送心跳的间隔
	HeartbeatTimeout  = 30 * time.Second // 心跳超时时间
	ReconnectGrace    = 8 * time.Second  // 断线后等待重连的宽限时间
)

var upgrader = &websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// 允许所有CORS请求，生产环境应该限制
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type ChessHub struct {
	Rooms      map[int](*ChessRoom)
	Clients    map[int]*Client
	commands   chan hubCommand
	spareRooms []room.RoomInfo // 有空位的房间id
	mu         sync.Mutex
	pool       *utils.WorkerPool
	matchPool  [](*Client)
	// 记录断开后的延迟删除定时器，以支持短时重连
	disconnectTimers map[int]*time.Timer
}

func NewChessHub() *ChessHub {
	pool := utils.NewWorkerPool()
	hub := &ChessHub{
		Rooms:            make(map[int](*ChessRoom)),
		Clients:          make(map[int]*Client),
		commands:         make(chan hubCommand),
		spareRooms:       make([]room.RoomInfo, 0),
		mu:               sync.Mutex{},
		pool:             pool,
		disconnectTimers: make(map[int]*time.Timer),
	}
	pool.Start()

	return hub
}

func (ch *ChessHub) Run() {
	go func() {
		for err := range ch.pool.ErrChan {
			log.Printf("Worker pool error: %v\n", err)
		}
	}()
	for cmd := range ch.commands {
		ch.pool.Process(context.Background(), func() error {
			switch cmd.commandType {
			case commandRegister:
				client := cmd.client
				ch.mu.Lock()
				ch.Clients[client.Id] = client
				ch.mu.Unlock()
				// 在线用户
				database.SetValue(fmt.Sprint(client.Id), "a", 0)
				// 将在线状态写入 MySQL
				if err := database.GetMysqlDb().Model(&modeluser.User{}).Where("id = ?", client.Id).Update("online", true).Error; err != nil {
					// 不阻塞主流程，记录或忽略错误
				}
			case commandUnregister:
				client := cmd.client
				roomId := client.RoomId
				ch.mu.Lock()
				room, ok := ch.Rooms[roomId]
				ch.mu.Unlock()
				if ok {
					var target *Client
					if room.Current == client {
						target = room.Next
					} else {
						target = room.Current
					}
					if target != nil {
						ch.sendMessage(target, NormalMessage{
							BaseMessage: BaseMessage{Type: messageNormal},
							Message:     "对方已断开连接",
						})
					}
					room.clear()
					ch.mu.Lock()
					delete(ch.Rooms, roomId)
					// 如果房间原本只有一个人，那么删除房间
					for i, r := range ch.spareRooms {
						if r.Id == roomId {
							ch.spareRooms = slices.Delete(ch.spareRooms, i, i+1)
							break
						}
					}
					ch.mu.Unlock()
				}
				ch.mu.Lock()
				// 从 Clients 中移除
				if _, ok := ch.Clients[client.Id]; ok {
					// 更新在线状态为 false
					if err := database.GetMysqlDb().Model(&modeluser.User{}).Where("id = ?", client.Id).Update("online", false).Error; err != nil {
						// 忽略错误
					}
					delete(ch.Clients, client.Id)
					if client.Conn != nil {
						client.Conn.Close()
					}
				}
				// 从匹配池中移除任何残留的该客户端引用，避免再次被匹配
				for i := len(ch.matchPool) - 1; i >= 0; i-- {
					if ch.matchPool[i].Id == client.Id {
						ch.matchPool = append(ch.matchPool[:i], ch.matchPool[i+1:]...)
					}
				}
				ch.mu.Unlock()
				database.DeleteValue(fmt.Sprint(client.Id))
			case commandMatch:
				client := cmd.client
				ch.mu.Lock()
				// 防止重复加入匹配池
				already := false
				for _, c := range ch.matchPool {
					if c.Id == client.Id {
						already = true
						break
					}
				}
				if !already {
					ch.matchPool = append(ch.matchPool, client)
				}
				fmt.Println(ch.matchPool)
				if len(ch.matchPool) < 2 {
					client.sendMessage(NormalMessage{
						BaseMessage: BaseMessage{Type: messageNormal},
						Message:     "正在匹配，请稍等",
					})
					ch.mu.Unlock()
					return nil
				}
				// 匹配成功，创建房间
				room := NewChessRoom()
				room.join(ch.matchPool[0])
				room.join(ch.matchPool[1])
				ch.matchPool = ch.matchPool[2:]
				ch.Rooms[room.Id] = room
				ch.mu.Unlock()
				// 发送消息给两个客户端，通知他们开始游戏
				go func() {
					ch.commands <- hubCommand{
						commandType: commandStart,
						client:      client,
					}
				}()
			case commandMove:
				req := cmd.payload.(moveRequest)
				room := ch.Rooms[req.from.RoomId]
				if room == nil {
					req.from.sendMessage(NormalMessage{
						BaseMessage: BaseMessage{Type: messageNormal},
						Message:     "房间不存在",
					})
					return nil
				}

				if !room.isFull() {
					req.from.sendMessage(NormalMessage{
						BaseMessage: BaseMessage{Type: messageNormal},
						Message:     "游戏未开始",
					})
					return nil
				}

				if room.Current != req.from {
					// 如果不是当前玩家，则不允许移动
					req.from.sendMessage(NormalMessage{
						BaseMessage: BaseMessage{Type: messageNormal},
						Message:     "请等待对方移动",
					})
					return nil
				}

				target := room.Next

				target.sendMessage(req.move)

				// 将此次走棋记录追加到房间历史（按顺序保存 from, to）
				room.mu.Lock()
				room.History = append(room.History, req.move.From)
				room.History = append(room.History, req.move.To)
				room.mu.Unlock()

				// 交换当前玩家和下一个玩家
				room.exchange()
			case commandSendMessage:
				req := cmd.payload.(sendMessageRequest)
				err := req.target.sendMessage(req.message)
				if err != nil {
					return fmt.Errorf("发送消息失败: %v", err)
				}
			case commandStart:
				room := ch.Rooms[cmd.client.RoomId]
				if room == nil {
					cmd.client.RoomId = -1
					cmd.client.Status = userOnline
					cmd.client.sendMessage(NormalMessage{
						BaseMessage: BaseMessage{Type: messageNormal},
						Message:     "请进行匹配",
					})
					return nil
				}
				if !room.isFull() {
					cmd.client.sendMessage(NormalMessage{
						BaseMessage: BaseMessage{Type: messageNormal},
						Message:     "房间未满员，无法开始游戏",
					})
					return nil
				}
				room.Current.startPlay(roleRed)
				room.Next.startPlay(roleBlack)
				cur := startMessage{BaseMessage: BaseMessage{Type: messageStart}, Role: "red"}
				next := startMessage{BaseMessage: BaseMessage{Type: messageStart}, Role: "black"}
				room.Current.sendMessage(cur)
				room.Next.sendMessage(next)
				// 记录对局开始时间
				room.StartTime = time.Now()
				// 移除空余房间
				ch.mu.Lock()
				for i, r := range ch.spareRooms {
					if r.Id == room.Id {
						ch.spareRooms = slices.Delete(ch.spareRooms, i, i+1)
						break
					}
				}
				ch.mu.Unlock()
			case commandEnd:
				var winner clientRole

				room := ch.Rooms[cmd.client.RoomId]
				if room == nil {
					// 房间不存在：仍然尝试告知当前客户端比赛结束并返回胜者信息（避免显示“房间不存在”）
					if cmd.payload == nil {
						winner = cmd.client.Role
					} else {
						// payload 已经是客户端传来的 winner（clientRole），直接使用
						winner = cmd.payload.(clientRole)
					}
					endMsg := endMessage{
						BaseMessage: BaseMessage{Type: messageEnd},
						Winner:      winner,
					}
					cmd.client.sendMessage(endMsg)
					return nil
				}
				if cmd.payload == nil {
					winner = cmd.client.Role
				} else {
					// payload 已经是客户端传来的 winner（clientRole），直接使用
					winner = cmd.payload.(clientRole)
				}
				// 发送消息给两个客户端，通知他们结束游戏
				endMsg := endMessage{
					BaseMessage: BaseMessage{Type: messageEnd},
					Winner:      winner,
				}
				room.Current.sendMessage(endMsg)
				room.Next.sendMessage(endMsg)
				// 保存对局记录到数据库（在清理房间前保存），按实际赢家记录
				saveGameRecord(room, winner)
				room.clear()
				delete(ch.Rooms, cmd.client.RoomId)
			case commandHeartbeat:
				// 更新客户端的最后一次心跳时间
				client := cmd.client
				client.LastPong = time.Now()

			case commandDisconnect:
				// 当检测到连接断开时，设置一个定时器等待客户端重连，超时后再真正注销
				client := cmd.client
				ch.mu.Lock()
				// 如果已有定时器则先停止
				if t, ok := ch.disconnectTimers[client.Id]; ok {
					t.Stop()
				}
				// 标记连接为空
				if existing, ok := ch.Clients[client.Id]; ok {
					existing.Conn = nil
					// 如果该玩家在匹配池中，立即移除，避免在短暂断线时被匹配
					for i := len(ch.matchPool) - 1; i >= 0; i-- {
						if ch.matchPool[i].Id == existing.Id {
							ch.matchPool = append(ch.matchPool[:i], ch.matchPool[i+1:]...)
						}
					}
				}
				// 通知对手：对方断线，正在等待重连
				room := ch.Rooms[client.RoomId]
				if room != nil {
					var target *Client
					if room.Current == client {
						target = room.Next
					} else {
						target = room.Current
					}
					if target != nil {
						ch.sendMessage(target, NormalMessage{BaseMessage: BaseMessage{Type: messageNormal}, Message: "对方连接中断，正在等待重连"})
					}
				}

				// 启动定时器，宽限期到达则根据当时状态决定处理方式
				t := time.AfterFunc(ReconnectGrace, func() {
					ch.mu.Lock()
					currentClient, exists := ch.Clients[client.Id]
					ch.mu.Unlock()
					if !exists {
						return // 玩家已被清理
					}

					// 若玩家仍在游戏中，则判对手胜
					if currentClient.Status == userPlaying {
						var winner clientRole
						if currentClient.Role == roleRed {
							winner = roleBlack
						} else if currentClient.Role == roleBlack {
							winner = roleRed
						} else {
							winner = roleNone
						}
						ch.commands <- hubCommand{
							commandType: commandEnd,
							client:      currentClient,
							payload:     winner,
						}
					} else {
						// 若玩家不在游戏中（已离开或空闲），则直接注销
						ch.commands <- hubCommand{
							commandType: commandUnregister,
							client:      currentClient,
						}
					}
					// 清理定时器记录
					ch.mu.Lock()
					delete(ch.disconnectTimers, client.Id)
					ch.mu.Unlock()
				})
				ch.disconnectTimers[client.Id] = t
				ch.mu.Unlock()
			case commandJoin:
				joinMsg := cmd.payload.(joinMessage)
				ch.mu.Lock()
				room := ch.Rooms[joinMsg.RoomId]
				if room == nil {
					ch.sendMessage(cmd.client, NormalMessage{
						BaseMessage: BaseMessage{Type: messageNormal},
						Message:     "房间不存在",
					})
					ch.mu.Unlock()
					return nil
				}
				err := room.join(cmd.client)
				if err != nil {
					cmd.client.sendMessage(NormalMessage{
						BaseMessage: BaseMessage{Type: messageNormal},
						Message:     err.Error(),
					})
					ch.mu.Unlock()
					return nil
				}
				ch.mu.Unlock()
				// 发送消息给两个客户端，通知他们开始游戏
				go func() {
					ch.commands <- hubCommand{
						commandType: commandStart,
						client:      cmd.client,
					}
				}()
			case commandCreate:
				// 创建房间
				client := cmd.client
				r := NewChessRoom()
				r.join(client)
				ch.Rooms[r.Id] = r
				roomInfo := room.RoomInfo{
					Id: client.RoomId,
					Current: dtouser.UserInfo{
						ID: uint(client.Id),
					},
				}
				ch.mu.Lock()
				ch.spareRooms = append(ch.spareRooms, roomInfo)
				ch.mu.Unlock()
				// 发送消息给客户端，通知他们创建房间成功
				ch.sendMessage(client, NormalMessage{
					BaseMessage: BaseMessage{Type: messageCreate},
				})
				return nil
			// 新增：处理悔棋请求命令
			case commandRegretRequest:
				payload := cmd.payload.(regretRequestPayload)
				client := payload.from
				ch.handleRegretRequest(client)

			// 新增：处理悔棋响应命令
			case commandRegretResponse:
				payload := cmd.payload.(regretResponsePayload)
				client := payload.from
				ch.handleRegretResponse(client, payload.accepted)

			case commandDrawRequest:
				payload := cmd.payload.(drawRequestPayload)
				client := payload.from
				ch.handleDrawRequest(client)

			case commandDrawResponse:
				payload := cmd.payload.(drawResponsePayload)
				client := payload.from
				ch.handleDrawResponse(client, payload.accepted)
			case commandChatMessage:
				client := cmd.client
				chatMsg := cmd.payload.(*ChatMessage)
				room := ch.Rooms[client.RoomId]
				if room == nil {
					client.sendMessage(NormalMessage{
						BaseMessage: BaseMessage{Type: messageError},
						Message:     "房间不存在",
					})
					return nil
				}
				// 获取对手
				target := room.Next
				if client == room.Next {
					target = room.Current
				}
				if target != nil {
					target.sendMessage(chatMsg)
				}
			}
			return nil
		})
	}
}

func (ch *ChessHub) HandleConnection(c *gin.Context) {
	userId, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("用户未登录"))
		return
	}

	id, ok := userId.(int)
	if !ok {
		dto.ErrorResponse(c, dto.WithMessage("用户ID转换失败"))
		return
	}

	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("websocket upgrade error"))
		return
	}
	defer conn.Close()

	// 从数据库获取用户信息
	var user modeluser.User
	if err := database.GetMysqlDb().First(&user, id).Error; err != nil {
		dto.ErrorResponse(c, dto.WithMessage("获取用户信息失败"))
		return
	}

	// 检查是否已有客户端实例（短时重连）
	var client *Client
	ch.mu.Lock()
	existing, exists := ch.Clients[id]
	if exists {
		// 若存在断线定时器，取消它（客户端正在重连）
		if t, ok := ch.disconnectTimers[id]; ok {
			t.Stop()
			delete(ch.disconnectTimers, id)
		}
		// 复用现有客户端结构，仅替换连接
		existing.Conn = conn
		existing.LastPong = time.Now()
		existing.Username = user.Name
		client = existing

		// 更新在线状态为 true（重连成功）
		if err := database.GetMysqlDb().Model(&modeluser.User{}).Where("id = ?", id).Update("online", true).Error; err != nil {
			// 忽略错误
		}

		// 关键：检查房间是否还存在
		// 若房间已被清理（定时器到期导致的结果），则重置玩家状态为"在线"
		_, roomExists := ch.Rooms[client.RoomId]
		if !roomExists || client.RoomId == -1 {
			// 房间已不存在或玩家未在房间中，重置状态以允许重新匹配
			existing.Status = userOnline
			existing.RoomId = -1
			existing.Role = roleNone
		}
		ch.mu.Unlock()

		// 通知客户端重连成功
		client.sendMessage(NormalMessage{BaseMessage: BaseMessage{Type: messageNormal}, Message: "重连成功"})

		// 如果房间已被清理（例如对局已结束），发送同步消息以清理客户端的本地对局状态
		if !roomExists || client.RoomId == -1 {
			client.sendMessage(SyncMessage{BaseMessage: BaseMessage{Type: messageSync}, History: []Position{}, Role: "", CurrentTurn: ""})
		}

		// 若房间仍存在且玩家在其中，发送房间当前状态（同步棋步、角色与当前轮次）
		if roomExists && client.RoomId != -1 {
			ch.mu.Lock()
			room := ch.Rooms[client.RoomId]
			if room != nil {
				history := make([]Position, len(room.History))
				copy(history, room.History)
				var roleStr string
				if client.Role == roleRed {
					roleStr = "red"
				} else if client.Role == roleBlack {
					roleStr = "black"
				}
				var currentTurn string
				if room.Current != nil {
					if room.Current.Role == roleRed {
						currentTurn = "red"
					} else if room.Current.Role == roleBlack {
						currentTurn = "black"
					}
				}
				ch.mu.Unlock()
				client.sendMessage(SyncMessage{BaseMessage: BaseMessage{Type: messageSync}, History: history, Role: roleStr, CurrentTurn: currentTurn})
			} else {
				ch.mu.Unlock()
			}
		}
	} else {
		ch.mu.Unlock()
		client = NewClient(conn, id, user.Name)
		ch.commands <- hubCommand{
			commandType: commandRegister,
			client:      client,
		}
	}

	conn.SetReadLimit(1024 * 1024)
	conn.SetPongHandler(func(string) error {
		conn.SetReadDeadline(time.Now().Add(HeartbeatTimeout))
		client.LastPong = time.Now()
		return nil
	})
	conn.SetCloseHandler(func(code int, text string) error {
		fmt.Printf("WebSocket connection closed with code %d: %s\n", code, text)
		return nil
	})

	conn.SetReadDeadline(time.Now().Add(HeartbeatTimeout))

	go func() {
		ticker := time.NewTicker(HeartbeatInterval)
		defer ticker.Stop()

		for range ticker.C {
			conn.SetWriteDeadline(time.Now().Add(10 * time.Second))
			if err := conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				fmt.Printf("发送 ping 失败: %v\n", err)
				return
			}
		}
	}()

	// 断开时不立即注销，而是进入断线等待流程，使用 commandDisconnect 启动定时器
	defer func() {
		ch.commands <- hubCommand{
			commandType: commandDisconnect,
			client:      client,
		}
	}()

	ch.sendMessage(client, NormalMessage{
		BaseMessage: BaseMessage{Type: messageNormal},
		Message:     "连接成功",
	})

	for {
		_, message, err := conn.ReadMessage()
		if err != nil {
			fmt.Printf("读取消息失败: %v\n", err)
			break
		}

		err = ch.handleMessage(client, message)
		if err != nil {
			fmt.Printf("处理消息失败: %v\n", err)
			return
		}
	}
	fmt.Println("客户端断开连接")
}

func (ch *ChessHub) GetSpareRooms(c *gin.Context) {
	ch.mu.Lock()
	defer ch.mu.Unlock()

	c.Set("rooms", ch.spareRooms)
	c.Next()
}

func (ch *ChessHub) handleMessage(client *Client, rawMessage []byte) error {
	var base BaseMessage
	err := json.Unmarshal(rawMessage, &base)
	if err != nil {
		return fmt.Errorf("解析消息失败: %v", err)
	}

	switch base.Type {
	case messageMatch:
		switch client.Status {
		case userOnline:
			client.Status = userMatching
			ch.commands <- hubCommand{
				commandType: commandMatch,
				client:      client,
			}
		case userMatching:
			msg := NormalMessage{
				BaseMessage: BaseMessage{Type: messageNormal},
				Message:     "您已在匹配队列中，请耐心等待",
			}
			ch.sendMessage(client, msg)
		case userPlaying:
			msg := NormalMessage{
				BaseMessage: BaseMessage{Type: messageNormal},
				Message:     "您已在游戏中",
			}
			ch.sendMessage(client, msg)
		}
	case messageMove:
		if client.Status == userPlaying {
			var moveMsg MoveMessage
			err := json.Unmarshal(rawMessage, &moveMsg)
			if err != nil {
				fmt.Printf("解析移动消息失败: %v\n", err)
				return err
			}

			ch.commands <- hubCommand{
				commandType: commandMove,
				client:      client,
				payload: moveRequest{
					from: client,
					move: moveMsg,
				},
			}
		} else {
			return fmt.Errorf("玩家不在游戏中")
		}
	case messageEnd:
		if client.Status == userPlaying {
			// 尝试解析消息体中的 winner 字段（可选）并作为 payload 传递到命令队列
			var em endMessage
			if err := json.Unmarshal(rawMessage, &em); err == nil {
				ch.commands <- hubCommand{
					commandType: commandEnd,
					client:      client,
					payload:     em.Winner,
				}
			} else {
				ch.commands <- hubCommand{
					commandType: commandEnd,
					client:      client,
				}
			}
		}
	case messageJoin:
		// 用户加入房间
		if client.Status == userPlaying {
			// 如果用户已经在游戏中，则不允许加入房间
			msg := NormalMessage{
				BaseMessage: BaseMessage{Type: messageNormal},
				Message:     "您已在游戏中",
			}
			ch.sendMessage(client, msg)
			return nil
		}
		var joinMsg joinMessage
		err := json.Unmarshal(rawMessage, &joinMsg)
		if err != nil {
			fmt.Printf("解析加入房间消息失败: %v\n", err)
			return nil
		}
		ch.commands <- hubCommand{
			commandType: commandJoin,
			client:      client,
			payload:     joinMsg,
		}
	case messageCreate:
		// 用户创建房间
		if client.Status == userPlaying {
			// 如果用户已经在游戏中，则不允许创建房间
			msg := NormalMessage{
				BaseMessage: BaseMessage{Type: messageNormal},
				Message:     "您已在游戏中",
			}
			ch.sendMessage(client, msg)
			return nil
		}
		ch.commands <- hubCommand{
			commandType: commandCreate,
			client:      client,
		}
	case messageGiveUp:
		if client.Status == userPlaying {
			// 将认输请求转换为结束命令，payload 传递为对手角色（认输方的对手为胜者）
			var winner clientRole
			if client.Role == roleRed {
				winner = roleBlack
			} else if client.Role == roleBlack {
				winner = roleRed
			} else {
				winner = roleNone
			}
			ch.commands <- hubCommand{
				commandType: commandEnd,
				client:      client,
				payload:     winner,
			}
		}
	// 新增：处理悔棋请求
	case messageRegretRequest:
		if client.Status != userPlaying || client.RoomId == -1 {
			return client.sendMessage(NormalMessage{
				BaseMessage: BaseMessage{Type: messageError},
				Message:     "不在游戏中，无法请求悔棋",
			})
		}
		// 发送内部命令到命令队列
		ch.commands <- hubCommand{
			commandType: commandRegretRequest,
			client:      client,
			payload: regretRequestPayload{
				from: client,
			},
		}

	// 新增：处理前端悔棋响应消息，转为内部命令
	case messageRegretResponse:
		if client.Status != userPlaying || client.RoomId == -1 {
			return client.sendMessage(NormalMessage{
				BaseMessage: BaseMessage{Type: messageError},
				Message:     "不在游戏中，无法响应悔棋",
			})
		}
		var resp RegretResponseMessage
		if err := json.Unmarshal(rawMessage, &resp); err != nil {
			return fmt.Errorf("解析悔棋响应失败: %v", err)
		}
		// 发送内部命令到命令队列
		ch.commands <- hubCommand{
			commandType: commandRegretResponse,
			client:      client,
			payload: regretResponsePayload{
				from:     client,
				accepted: resp.Accepted,
			},
		}
	case messageChatMessage:
		if client.Status != userPlaying || client.RoomId == -1 {
			return client.sendMessage(NormalMessage{
				BaseMessage: BaseMessage{Type: messageError},
				Message:     "不在游戏中，无法发送消息",
			})
		}
		var chatMsg ChatMessage
		if err := json.Unmarshal(rawMessage, &chatMsg); err != nil {
			return fmt.Errorf("解析聊天消息失败: %v", err)
		}

		ch.commands <- hubCommand{
			commandType: commandChatMessage,
			client:      client,
			payload: &ChatMessage{
				BaseMessage: BaseMessage{Type: messageChatMessage},
				Content:     chatMsg.Content,
				Sender:      client.Username,
			},
		}

	case messageDrawRequest:
		if client.Status != userPlaying || client.RoomId == -1 {
			return client.sendMessage(NormalMessage{
				BaseMessage: BaseMessage{Type: messageError},
				Message:     "不在游戏中，无法请求和棋",
			})
		}
		ch.commands <- hubCommand{
			commandType: commandDrawRequest,
			client:      client,
			payload:     drawRequestPayload{from: client},
		}

	case messageDrawResponse:
		if client.Status != userPlaying || client.RoomId == -1 {
			return client.sendMessage(NormalMessage{
				BaseMessage: BaseMessage{Type: messageError},
				Message:     "不在游戏中，无法响应和棋",
			})
		}
		var resp DrawResponseMessage
		if err := json.Unmarshal(rawMessage, &resp); err != nil {
			return fmt.Errorf("解析和棋响应失败: %v", err)
		}
		ch.commands <- hubCommand{
			commandType: commandDrawResponse,
			client:      client,
			payload: drawResponsePayload{
				from:     client,
				accepted: resp.Accepted,
			},
		}
	}
	return nil
}

