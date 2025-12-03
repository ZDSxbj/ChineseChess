package websocket

import (
	"chinese-chess-backend/database"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"
	recordModel "chinese-chess-backend/model/record"
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
	RecordSaved     bool       // 标记对局记录是否已保存，防止重复保存
	mu              sync.Mutex // 保护History等共享资源
}

func NewChessRoom() *ChessRoom {
	idLock.Lock()
	defer idLock.Unlock()
	nextId++
	return &ChessRoom{
		Id:          nextId,
		Nums:        0,
		Current:     nil,
		Next:        nil,
		History:     make([]Position, 0),
		StartTime:   time.Time{},
		RecordSaved: false,
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

// saveGameRecord 将房间数据持久化到数据库
func saveGameRecord(room *ChessRoom, winner clientRole) {
	if room == nil {
		return
	}

	// 避免重复保存：如果已经保存过则直接返回
	room.mu.Lock()
	if room.RecordSaved {
		room.mu.Unlock()
		return
	}
	// 标记为已保存，防止并发或重复调用导致多次写入
	room.RecordSaved = true
	room.mu.Unlock()

	var redID uint
	var blackID uint
	// 根据 Client.Role 来确定红黑双方的用户ID
	if room.Current != nil {
		if room.Current.Role == roleRed {
			redID = uint(room.Current.Id)
		} else if room.Current.Role == roleBlack {
			blackID = uint(room.Current.Id)
		}
	}
	if room.Next != nil {
		if room.Next.Role == roleRed {
			redID = uint(room.Next.Id)
		} else if room.Next.Role == roleBlack {
			blackID = uint(room.Next.Id)
		}
	}

	// 以紧凑数字串保存历史，例如: [{x:6,y:6},{x:6,y:5}] -> "6665"
	// 注意：room.History 中保存的位置是按移动者视角记录的（客户端未统一坐标），
	// 因此这里将历史规范化为统一的“红方视角”再保存：如果某一步的移动者为黑方，则翻转坐标 (x->8-x, y->9-y)
	room.mu.Lock()
	historyCopy := make([]Position, len(room.History))
	copy(historyCopy, room.History)
	room.mu.Unlock()

	// 将按移动对对（from,to）处理，假设红方先手
	var sb strings.Builder
	for i := 0; i+1 < len(historyCopy); i += 2 {
		// moveIdx: 0 表示第一手（红方），1 表示第二手（黑方），以此类推
		moveIdx := i / 2
		moverIsBlack := (moveIdx%2 == 1)
		from := historyCopy[i]
		to := historyCopy[i+1]
		if moverIsBlack {
			from = Position{X: 8 - from.X, Y: 9 - from.Y}
			to = Position{X: 8 - to.X, Y: 9 - to.Y}
		}
		sb.WriteString(strconv.Itoa(from.X))
		sb.WriteString(strconv.Itoa(from.Y))
		sb.WriteString(strconv.Itoa(to.X))
		sb.WriteString(strconv.Itoa(to.Y))
	}
	historyStr := sb.String()

	// 映射结果：0 = red win, 1 = black win, 2 = draw
	result := 2
	if winner == roleRed {
		result = 0
	} else if winner == roleBlack {
		result = 1
	} else {
		result = 2
	}

	rec := recordModel.GameRecord{
		RedID:     redID,
		BlackID:   blackID,
		StartTime: room.StartTime,
		Result:    result,
		History:   historyStr,
		RedFlag:   false,
		BlackFlag: false,
		GameType:  0,
	}

	if err := database.GetMysqlDb().Create(&rec).Error; err != nil {
		log.Printf("failed to save game record: %v", err)
	}
}
