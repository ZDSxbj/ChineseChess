package websocket

import (
	"fmt"
	"sync"
	"time"
)

var (
	nextId int = 0
	idLock sync.Mutex
)

type ChessRoom struct {
	Id              int
	Nums            int     // 已有人数
	Current         *Client // 先进入房间的作为先手，默认为当前玩家
	Next            *Client // 后进入房间的作为后手，默认为下一个玩家
	History         []Position
	StartTime       time.Time  // 记录对局开始时间
	RegretRequester *Client    // 新增：记录悔棋请求发起方
	mu              sync.Mutex // 保护History等共享资源
}

func NewChessRoom() *ChessRoom {
	idLock.Lock()
	defer idLock.Unlock()
	nextId++
	return &ChessRoom{
		Id:        nextId,
		Nums:      0,
		Current:   nil,
		Next:      nil,
		History:   make([]Position, 0),
		StartTime: time.Time{},
	}
}

// func (cr * ChessRoom) isEmpty() bool {
// 	return cr.Nums == 0
// }

func (cr *ChessRoom) isFull() bool {
	return cr.Nums >= 2
}

func (cr *ChessRoom) exchange() {
	if cr.Current == nil || cr.Next == nil {
		return
	}
	cr.Current, cr.Next = cr.Next, cr.Current
}

func (cr *ChessRoom) clear() {
	if cr.Current != nil {
		cr.Current.RoomId = -1
		cr.Current.Status = userOnline
		cr.Current = nil
	}
	if cr.Next != nil {
		cr.Next.RoomId = -1
		cr.Next.Status = userOnline
		cr.Next = nil
	}
	cr.Nums = 0
}

func (cr *ChessRoom) join(c *Client) error {
	if cr.isFull() {
		return fmt.Errorf("房间满了")
	}
	c.RoomId = cr.Id
	if cr.Current == nil {
		cr.Current = c
	} else {
		cr.Next = c
	}

	cr.Nums++

	return nil
}

// func (cr *ChessRoom) leave(c *Client) error {
// 	if cr.isEmpty() {
// 		return fmt.Errorf("不在该房间")
// 	}
// 	if c == cr.Current {
// 		cr.Current = nil
// 		cr.Nums--
// 		return nil
// 	}
// 	if c == cr.Next {
// 		cr.Next = nil
// 		cr.Nums--
// 		return nil
// 	}
// 	return fmt.Errorf("不在该房间")
// }

// 处理聊天消息
func (cr *ChessRoom) handleChat(sender *Client, content string) {
	// 创建聊天消息
	chatMsg := &ChatMessage{
		BaseMessage: BaseMessage{Type: messageChatMessage},
		Content:     content,
		Sender:      sender.Username,
	}

	// 发送给对手
	target := cr.Current
	if sender == cr.Current {
		target = cr.Next
	}

	if target != nil {
		target.Send <- chatMsg
	}
}
