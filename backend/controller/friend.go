package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	dto "chinese-chess-backend/dto"
	"chinese-chess-backend/service"
)

type FriendController struct {
	friendService *service.FriendService
	frService     *service.FriendRequestService
}

func NewFriendController(friendService *service.FriendService) *FriendController {
	return &FriendController{friendService: friendService, frService: service.NewFriendRequestService()}
}

func (fc *FriendController) GetFriends(c *gin.Context) {
	uid, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	userID := uid.(int)

	resp, err := fc.friendService.ListFriends(userID)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(resp))
}

func (fc *FriendController) DeleteFriend(c *gin.Context) {
	uid, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	userID := uid.(int)

	fidStr := c.Param("friendId")
	fid, err := strconv.Atoi(fidStr)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("非法的好友ID"))
		return
	}

	if err := fc.friendService.DeleteFriend(userID, fid); err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	// 返回 200 标准响应
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}

// Get incoming friend requests
func (fc *FriendController) GetFriendRequests(c *gin.Context) {
	uid, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	userID := uid.(int)
	resp, err := fc.frService.ListIncoming(uint(userID))
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(resp))
}

// Check whether current user has already sent a friend request to given receiver
func (fc *FriendController) CheckFriendRequest(c *gin.Context) {
	uid, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	userID := uid.(int)
	recvStr := c.Query("receiverId")
	if recvStr == "" {
		dto.ErrorResponse(c, dto.WithMessage("缺少 receiverId 参数"))
		return
	}
	rid, err := strconv.Atoi(recvStr)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("receiverId 参数非法"))
		return
	}
	existsFlag, err := fc.frService.Exists(uint(userID), uint(rid))
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(gin.H{"exists": existsFlag}))
}
