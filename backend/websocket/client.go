package websocket

import (
	"fmt"
	"time"

	"github.com/gorilla/websocket"
)

type clientStatus int

const (
	userOnline clientStatus = iota + 1
	userPlaying
	userMatching
)

type clientRole int

const (
	// 0表示没有角色，1表示红方，2表示黑方
	roleNone clientRole = iota
	roleRed
	roleBlack
)

type Client struct {
	Conn     *websocket.Conn
	Id       int
	Status   clientStatus
	RoomId   int
	Role     clientRole // 角色
	LastPong time.Time  // 上次收到PONG的时间
	Username string     // 用户名
	Send     chan any   // 发送消息的通道
}

func NewClient(conn *websocket.Conn, id int, username string) *Client {
	return &Client{
		Conn:     conn,
		Id:       id,
		Status:   userOnline,
		RoomId:   -1,
		Role:     roleNone,
		LastPong: time.Now(),
		Username: username,
		Send:     make(chan any, 256),
	}
}

func (c *Client) sendMessage(message any) error {
	if c.Conn == nil {
		return fmt.Errorf("client connection is nil")
	}
	err := c.Conn.WriteJSON(message)
	if err != nil {
		return err
	}
	return nil
}

func (c *Client) startPlay(role clientRole) {
	c.Role = role
	c.Status = userPlaying
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
		}
		room.mu.Unlock()

		respMsg := RegretResponseMessage{
			BaseMessage: BaseMessage{Type: messageRegretResponse},
			Accepted:    true,
		}
		requester.sendMessage(respMsg)
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

	// 向对手发送和棋请求（使用 NormalMessage 携带类型）
	opponent.sendMessage(NormalMessage{
		BaseMessage: BaseMessage{Type: messageDrawRequest},
		Message:     "对方请求和棋",
	})
}

// 新增：处理和棋响应（若同意则结束为和棋）
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

	// 确定请求方
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

	// 通知请求方和棋响应
	respMsg := DrawResponseMessage{
		BaseMessage: BaseMessage{Type: messageDrawResponse},
		Accepted:    accepted,
	}
	requester.sendMessage(respMsg)

	if accepted {
		// 若同意，发送结束消息（和棋, winner = roleNone）给双方并清理房间
		endMsg := endMessage{
			BaseMessage: BaseMessage{Type: messageEnd},
			Winner:      roleNone,
		}
		room.Current.sendMessage(endMsg)
		room.Next.sendMessage(endMsg)
		room.clear()
		ch.mu.Lock()
		delete(ch.Rooms, requester.RoomId)
		ch.mu.Unlock()
	}
}
