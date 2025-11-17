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

func (ch *ChessHub) sendMessage(client *Client, message any) {
	ch.commands <- hubCommand{
		commandType: commandSendMessage,
		payload: sendMessageRequest{
			target:  client,
			message: message,
		},
	}
}