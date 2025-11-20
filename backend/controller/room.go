// package controller

// import (
// 	"fmt"

// 	"github.com/gin-gonic/gin"

// 	"chinese-chess-backend/dto"
// 	"chinese-chess-backend/dto/room"
// 	"chinese-chess-backend/service"
// )

// type RoomController struct {
// 	roomService *service.RoomService
// }

// func NewRoomController(roomService *service.RoomService) *RoomController {
// 	return &RoomController{
// 		roomService: roomService,
// 	}
// }

// func (rc *RoomController) GetSpareRooms(c *gin.Context) {
// 	info, exists := c.Get("rooms")
// 	if !exists {
// 		dto.ErrorResponse(c, dto.WithMessage("room not found"))
// 		return
// 	}
// 	infos, ok := info.([]room.RoomInfo)
// 	if !ok {
// 		dto.ErrorResponse(c, dto.WithMessage("room not found"))
// 		return
// 	}
// 	fmt.Println(infos)
// 	resp, err := rc.roomService.GetSpareRooms(room.GetSpareRoomsRequest{
// 		Infos: infos})
// 	if err != nil {
// 		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
// 		return
// 	}
// 	dto.SuccessResponse(c, dto.WithData(resp))
// }

package controller

import (
	"github.com/gin-gonic/gin"

	"chinese-chess-backend/dto"
	"chinese-chess-backend/dto/room"
	"chinese-chess-backend/service"
)

type RoomController struct {
	roomService *service.RoomService
}

func NewRoomController(roomService *service.RoomService) *RoomController {
	return &RoomController{
		roomService: roomService,
	}
}

// GetSpareRooms 获取空闲房间列表
func (rc *RoomController) GetSpareRooms(c *gin.Context) {
	var req room.GetSpareRoomsRequest

	// 绑定查询参数
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, dto.WithMessage("参数格式错误"))
		return
	}

	// 调用服务层获取房间列表
	resp, err := rc.roomService.GetSpareRooms(req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("获取房间列表失败: "+err.Error()))
		return
	}

	dto.SuccessResponse(c, dto.WithData(resp))
}

// CreateRoom 创建房间
func (rc *RoomController) CreateRoom(c *gin.Context) {
	// 从认证中间件中获取用户ID
	userID, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("用户未登录"))
		return
	}

	var req room.CreateRoomRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, dto.WithMessage("房间名称不能为空"))
		return
	}

	// 调用服务层创建房间
	roomInfo, err := rc.roomService.CreateRoom(uint(userID.(int)), req.Name)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("创建房间失败: "+err.Error()))
		return
	}

	dto.SuccessResponse(c,
		dto.WithMessage("房间创建成功"),
		dto.WithData(roomInfo),
	)
}

// JoinRoom 加入房间
func (rc *RoomController) JoinRoom(c *gin.Context) {
	// 从认证中间件中获取用户ID
	userID, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("用户未登录"))
		return
	}

	var req struct {
		RoomID uint `json:"room_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, dto.WithMessage("房间ID不能为空"))
		return
	}

	// 调用服务层加入房间
	err := rc.roomService.JoinRoom(req.RoomID, uint(userID.(int)))
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("加入房间失败: "+err.Error()))
		return
	}

	dto.SuccessResponse(c, dto.WithMessage("成功加入房间"))
}

// GetRoomInfo 获取房间详细信息
func (rc *RoomController) GetRoomInfo(c *gin.Context) {
	var req struct {
		RoomID uint `form:"room_id" binding:"required"`
	}

	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, dto.WithMessage("房间ID不能为空"))
		return
	}

	// 调用服务层获取房间信息
	roomInfo, err := rc.roomService.GetRoomInfo(req.RoomID)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("获取房间信息失败: "+err.Error()))
		return
	}

	dto.SuccessResponse(c, dto.WithData(roomInfo))
}

// LeaveRoom 离开房间
func (rc *RoomController) LeaveRoom(c *gin.Context) {
	// 从认证中间件中获取用户ID
	userID, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("用户未登录"))
		return
	}

	var req struct {
		RoomID uint `json:"room_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, dto.WithMessage("房间ID不能为空"))
		return
	}

	// 调用服务层离开房间
	err := rc.roomService.LeaveRoom(req.RoomID, uint(userID.(int)))
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("离开房间失败: "+err.Error()))
		return
	}

	dto.SuccessResponse(c, dto.WithMessage("成功离开房间"))
}
