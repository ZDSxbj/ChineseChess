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

	// dtouser "chinese-chess-backend/dto/user"
	modeluser "chinese-chess-backend/model/user"
	"chinese-chess-backend/utils"
	"slices"
)

type RoomInfo struct {
	Id      int      `json:"id"`
	Current UserInfo `json:"current"`
	Next    UserInfo `json:"next"`
}

type UserInfo struct {
	ID   uint   `json:"id"`
	Name string `json:"name"`
	Exp  int    `json:"exp"`
}

const (
	HeartbeatInterval = 5 * time.Second  // 发送心跳的间隔
	HeartbeatTimeout  = 30 * time.Second // 心跳超时时间
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
	spareRooms []RoomInfo // 有空位的房间id
	mu         sync.Mutex
	pool       *utils.WorkerPool
	matchPool  [](*Client)
}

func NewChessHub() *ChessHub {
	pool := utils.NewWorkerPool()
	hub := &ChessHub{
		Rooms:      make(map[int](*ChessRoom)),
		Clients:    make(map[int]*Client),
		commands:   make(chan hubCommand),
		spareRooms: make([]RoomInfo, 0),
		mu:         sync.Mutex{},
		pool:       pool,
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
				if _, ok := ch.Clients[client.Id]; ok {
					delete(ch.Clients, client.Id)
					client.Conn.Close()
				}
				ch.mu.Unlock()
				database.DeleteValue(fmt.Sprint(client.Id))
			case commandMatch:
				client := cmd.client
				ch.mu.Lock()
				ch.matchPool = append(ch.matchPool, client)
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
					cmd.client.sendMessage(NormalMessage{
						BaseMessage: BaseMessage{Type: messageNormal},
						Message:     "房间不存在",
					})
					return nil
				}
				if cmd.payload == nil {
					winner = cmd.client.Role
				} else {
					r := cmd.payload.(clientRole)
					if r != roleNone {
						if r == roleRed {
							r = roleBlack
						} else {
							r = roleRed
						}
					}
					winner = r
				}
				// 发送消息给两个客户端，通知他们结束游戏
				endMsg := endMessage{
					BaseMessage: BaseMessage{Type: messageEnd},
					Winner:      winner,
				}
				room.Current.sendMessage(endMsg)
				room.Next.sendMessage(endMsg)
				room.clear()
				delete(ch.Rooms, cmd.client.RoomId)
			case commandHeartbeat:
				// 更新客户端的最后一次心跳时间
				client := cmd.client
				client.LastPong = time.Now()
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
				roomInfo := RoomInfo{
					Id: client.RoomId,
					Current: UserInfo{
						ID:   uint(client.Id),
						Name: client.Username,
						Exp:  0,
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

			// 新增：处理和棋请求命令
			case commandDrawRequest:
				payload := cmd.payload.(drawRequestPayload)
				client := payload.from
				ch.handleDrawRequest(client)

			// 新增：处理和棋响应命令
			case commandDrawResponse:
				payload := cmd.payload.(drawResponsePayload)
				client := payload.from
				ch.handleDrawResponse(client, payload.accepted)
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

	// 创建一个新的客户端
	client := NewClient(conn, id, user.Name)

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

	ch.commands <- hubCommand{
		commandType: commandRegister,
		client:      client,
	}
	defer func() {
		ch.commands <- hubCommand{
			commandType: commandUnregister,
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

	// 添加详细日志
	fmt.Printf("🔍 收到消息 - 类型: %d, 用户ID: %d, 房间ID: %s, 状态: %d\n",
		base.Type, client.Id, client.RoomId, client.Status)

	switch base.Type {
	case messageMatch:
		fmt.Printf("🎯 处理匹配消息，用户状态: %d\n", client.Status)
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
			ch.commands <- hubCommand{
				commandType: commandEnd,
				client:      client,
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
			ch.commands <- hubCommand{
				commandType: commandEnd,
				client:      client,
				payload:     client.Role,
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

	// 新增：处理和棋请求
	case messageDrawRequest:
		if client.Status != userPlaying || client.RoomId == -1 {
			return client.sendMessage(NormalMessage{
				BaseMessage: BaseMessage{Type: messageError},
				Message:     "不在游戏中，无法请求和棋",
			})
		}
		// 发送内部命令到命令队列
		ch.commands <- hubCommand{
			commandType: commandDrawRequest,
			client:      client,
			payload: drawRequestPayload{
				from: client,
			},
		}

	// 新增：处理前端和棋响应消息，转为内部命令
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
		// 发送内部命令到命令队列
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

func (ch *ChessHub) sendMessage(client *Client, message any) {
	ch.commands <- hubCommand{
		commandType: commandSendMessage,
		payload: sendMessageRequest{
			target:  client,
			message: message,
		},
	}
}

// 新增：处理悔棋请求（转发给对手）
func (ch *ChessHub) handleRegretRequest(requester *Client) {
	ch.mu.Lock()
	room, ok := ch.Rooms[requester.RoomId]
	ch.mu.Unlock()
	if !ok {
		requester.sendMessage(NormalMessage{
			BaseMessage: BaseMessage{Type: messageError},
			Message:     "房间不存在",
		})
		return
	}

	// 确定对手
	var opponent *Client
	if room.Current == requester {
		opponent = room.Next
	} else {
		opponent = room.Current
	}
	if opponent == nil {
		requester.sendMessage(NormalMessage{
			BaseMessage: BaseMessage{Type: messageError},
			Message:     "对手不存在",
		})
		return
	}

	// 向对手发送悔棋请求
	opponent.sendMessage(NormalMessage{
		BaseMessage: BaseMessage{Type: messageRegretRequest},
		Message:     "对方请求悔棋",
	})
}

// 新增：处理悔棋响应（同步双方状态）
func (ch *ChessHub) handleRegretResponse(responder *Client, accepted bool) {
	ch.mu.Lock()
	room, ok := ch.Rooms[responder.RoomId]
	ch.mu.Unlock()
	if !ok {
		responder.sendMessage(NormalMessage{
			BaseMessage: BaseMessage{Type: messageError},
			Message:     "房间不存在",
		})
		return
	}

	// 确定悔棋请求发起方
	var requester *Client
	if room.Current == responder {
		requester = room.Next
	} else {
		requester = room.Current
	}
	if requester == nil {
		responder.sendMessage(NormalMessage{
			BaseMessage: BaseMessage{Type: messageError},
			Message:     "请求方不存在",
		})
		return
	}

	if accepted {
		// 同意悔棋：同步双方执行悔棋，更新房间历史记录
		room.mu.Lock()
		if len(room.History) > 0 {
			room.History = room.History[:len(room.History)-1] // 移除最后一步,这个有争议，需要后续修改
		}
		room.mu.Unlock()

		// 通知请求方执行悔棋
		respMsg := RegretResponseMessage{
			BaseMessage: BaseMessage{Type: messageRegretResponse},
			Accepted:    true,
		}
		requester.sendMessage(respMsg)
		if room.Current == responder {
			room.Current = requester
			room.Next = responder
		}
	} else {
		// 拒绝悔棋：仅通知请求方
		requester.sendMessage(RegretResponseMessage{
			BaseMessage: BaseMessage{Type: messageRegretResponse},
			Accepted:    false,
		})
	}
}

// 新增：处理和棋请求（转发给对手）
func (ch *ChessHub) handleDrawRequest(requester *Client) {
	fmt.Printf("🚀 进入 handleDrawRequest，用户: %d, 房间: %s\n", requester.Id, requester.RoomId)
	ch.mu.Lock()
	room, ok := ch.Rooms[requester.RoomId]
	ch.mu.Unlock()
	if !ok {
		requester.sendMessage(NormalMessage{
			BaseMessage: BaseMessage{Type: messageError},
			Message:     "房间不存在",
		})
		return
	}

	// 确定对手
	var opponent *Client
	if room.Current == requester {
		opponent = room.Next
	} else {
		opponent = room.Current
	}
	if opponent == nil {
		requester.sendMessage(NormalMessage{
			BaseMessage: BaseMessage{Type: messageError},
			Message:     "对手不存在",
		})
		return
	}
	fmt.Printf("📤 准备向对手发送和棋请求，对手ID: %d\n", opponent.Id)
	// 向对手发送和棋请求
	opponent.sendMessage(NormalMessage{
		BaseMessage: BaseMessage{Type: messageDrawRequest},
		Message:     "对方请求和棋",
	})
	fmt.Printf("✅ 和棋请求发送完成\n")
}

// 新增：处理和棋响应（同步双方状态）
func (ch *ChessHub) handleDrawResponse(responder *Client, accepted bool) {
	ch.mu.Lock()
	room, ok := ch.Rooms[responder.RoomId]
	ch.mu.Unlock()
	if !ok {
		responder.sendMessage(NormalMessage{
			BaseMessage: BaseMessage{Type: messageError},
			Message:     "房间不存在",
		})
		return
	}

	// 确定和棋请求发起方
	var requester *Client
	if room.Current == responder {
		requester = room.Next
	} else {
		requester = room.Current
	}
	if requester == nil {
		responder.sendMessage(NormalMessage{
			BaseMessage: BaseMessage{Type: messageError},
			Message:     "请求方不存在",
		})
		return
	}

	// if accepted {
	// 	// 同意和棋：通知双方游戏结束（和局）
	// 	drawMsg := NormalMessage{
	// 		BaseMessage: BaseMessage{Type: messageEnd},
	// 		Message:     "游戏结束，和棋",
	// 	}

	// 	requester.sendMessage(drawMsg)
	// 	responder.sendMessage(drawMsg)

	// 	// 可选：重置房间状态或标记游戏结束
	// 	room.mu.Lock()
	// 	// 这里可以添加清理房间状态的逻辑，比如：
	// 	// room.Status = "finished"
	// 	// room.Winner = "" // 和棋没有胜者
	// 	room.mu.Unlock()
	// } else {
	// 	// 拒绝和棋：仅通知请求方
	// 	requester.sendMessage(DrawResponseMessage{
	// 		BaseMessage: BaseMessage{Type: messageDrawResponse},
	// 		Accepted:    false,
	// 	})
	// }

	if accepted {
		// 同意和棋：通知双方和棋成功
		drawMsg := DrawResponseMessage{
			BaseMessage: BaseMessage{Type: messageDrawResponse},
			Accepted:    true,
			// Message:     "对方同意和棋，游戏结束",
		}

		requester.sendMessage(drawMsg)

		// 同时发送游戏结束命令
		ch.commands <- hubCommand{
			commandType: commandEnd,
			client:      responder,
			payload:     roleNone, // 和棋没有胜者
		}
	} else {
		// 拒绝和棋：仅通知请求方
		requester.sendMessage(DrawResponseMessage{
			BaseMessage: BaseMessage{Type: messageDrawResponse},
			Accepted:    false,
			// Message:     "对方拒绝和棋",
		})
	}
}
