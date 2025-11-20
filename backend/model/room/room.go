// package room

// import (
// 	userModel "chinese-chess-backend/model/user"
// )

// type RoomInfo struct {
// 	Id      int  `json:"id"`
// 	Current userModel.User `json:"current"`
// 	Next    userModel.User `json:"next"`
// }

package room

import (
	"time"

	userModel "chinese-chess-backend/model/user"
)

// Room 房间数据表模型
type Room struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	Name      string    `gorm:"size:100;not null" json:"name"`                 // 房间名称
	Player1ID uint      `gorm:"not null" json:"player1_id"`                    // 玩家1 ID
	Player2ID uint      `gorm:"default:0" json:"player2_id"`                   // 玩家2 ID（0表示空位）
	Status    string    `gorm:"size:20;default:'waiting';index" json:"status"` // 房间状态：waiting, playing, finished
	Board     string    `gorm:"type:text" json:"board"`                        // 棋盘状态（JSON格式）
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// 关联关系
	Player1 userModel.User `gorm:"foreignKey:Player1ID" json:"current"` // 保持原有的json标签
	Player2 userModel.User `gorm:"foreignKey:Player2ID" json:"next"`    // 保持原有的json标签
}

// TableName 设置表名
func (Room) TableName() string {
	return "rooms"
}

// RoomInfo 保持向后兼容的结构体（用于业务逻辑）
type RoomInfo struct {
	ID      uint           `json:"id"`
	Name    string         `json:"name"`
	Current userModel.User `json:"current"`
	Next    userModel.User `json:"next"`
	Status  string         `json:"status"`
	Board   string         `json:"board,omitempty"`
}

// ToRoomInfo 将Room转换为RoomInfo
func (r *Room) ToRoomInfo() RoomInfo {
	roomInfo := RoomInfo{
		ID:      r.ID,
		Name:    r.Name,
		Status:  r.Status,
		Board:   r.Board,
		Current: r.Player1,
	}

	// 如果有第二个玩家
	if r.Player2ID != 0 {
		roomInfo.Next = r.Player2
	} else {
		// 设置空的用户对象
		roomInfo.Next = userModel.User{}
	}

	return roomInfo
}

// IsFull 检查房间是否已满
func (r *Room) IsFull() bool {
	return r.Player2ID != 0
}

// CanStart 检查房间是否可以开始游戏
func (r *Room) CanStart() bool {
	return r.Status == "waiting" && r.IsFull()
}
