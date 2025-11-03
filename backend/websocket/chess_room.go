package websocket

import (
	"fmt"
	"sync"
	"encoding/json"  // 新增这一行，导入JSON处理包
)

var (
	nextId int = 0
	idLock sync.Mutex
)

type ChessRoom struct {
	Id      int
	Nums    int     // 已有人数
	Current *Client // 先进入房间的作为先手，默认为当前玩家
	Next    *Client // 后进入房间的作为后手，默认为下一个玩家
	History []Position
	mu      sync.Mutex    // 保护History等共享资源
}

func NewChessRoom() *ChessRoom {
	idLock.Lock()
	defer idLock.Unlock()
	nextId++
	return &ChessRoom{
		Id:      nextId,
		Nums:    0,
		Current: nil,
		Next:    nil,
		History: make([]Position, 0),
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
	} else  {
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

// 在chess_room.go中添加
func (cr *ChessRoom) handleRegretRequest(client *Client) error {
    // // 判断是否是当前玩家的回合
    // if cr.Current != client {
    //     return fmt.Errorf("不是你的回合，无法请求悔棋")
    // }
    
    // 向前手玩家发送悔棋请求
    if cr.Next != nil {
        msg := BaseMessage{Type: messageRegretRequest}
        data, _ := json.Marshal(msg)
        cr.Next.sendMessage(data)
    }
    return nil
}

func (cr *ChessRoom) handleRegretResponse(client *Client, accepted bool) error {
    // 判断是否是后手玩家的响应
    // if cr.Next != client {
    //     return fmt.Errorf("无权响应悔棋请求")
    // }
    
    // 向请求方发送响应
    if cr.Current != nil {
        msg := RegretResponseMessage{
            BaseMessage: BaseMessage{Type: messageRegretResponse},
            Accepted: accepted,
        }
        data, _ := json.Marshal(msg)
        cr.Current.sendMessage(data)
    }
    
    return nil
}