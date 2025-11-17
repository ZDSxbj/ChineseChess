// package room

// import (
// 	"chinese-chess-backend/dto/user"
// )

// type RoomInfo struct {
// 	Id      int           `json:"id"`
// 	Current user.UserInfo `json:"current"`
// 	Next    user.UserInfo `json:"next"`
// }

package room

import (
	"chinese-chess-backend/dto/user"
)

// RoomDTO API传输用的房间信息
type RoomDTO struct {
	ID      uint          `json:"id"`
	Name    string        `json:"name"`
	Current user.UserInfo `json:"current"`
	Next    user.UserInfo `json:"next"`
	Status  string        `json:"status"`
	Board   string        `json:"board,omitempty"`
}

// CreateRoomRequest 创建房间请求
type CreateRoomRequest struct {
	Name string `json:"name" binding:"required"`
}

// CreateRoomResponse 创建房间响应
type CreateRoomResponse struct {
	Room RoomDTO `json:"room"`
}

// JoinRoomRequest 加入房间请求
type JoinRoomRequest struct {
	RoomID uint `json:"room_id" binding:"required"`
}

// LeaveRoomRequest 离开房间请求
type LeaveRoomRequest struct {
	RoomID uint `json:"room_id" binding:"required"`
}

// StartGameRequest 开始游戏请求
type StartGameRequest struct {
	RoomID uint `json:"room_id" binding:"required"`
}
