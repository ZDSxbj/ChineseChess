package controller

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"chinese-chess-backend/database"
	dto "chinese-chess-backend/dto"
	"chinese-chess-backend/service"
	"chinese-chess-backend/websocket"
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

// AcceptFriendRequest accepts an incoming friend request (create relation, delete request, send greeting)
func (fc *FriendController) AcceptFriendRequest(c *gin.Context) {
	uid, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	userID := uid.(int)

	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("非法的请求ID"))
		return
	}
	reqID := uint(id64)

	// Accept request (create relation) and get relation id and sender id
	relationID, senderID, err := fc.frService.AcceptRequest(reqID)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	// Send a greeting message from accepter to sender and persist it
	chatSvc := service.NewChatService()
	greeting := "你好，我是...，我们一起聊天和下棋吧！"
	// try to derive current user's name
	var uname struct{ Name string }
	_ = database.GetMysqlDb().Table("user").Select("name").Where("id = ?", userID).Take(&uname).Error
	if uname.Name != "" {
		greeting = "你好，我是 " + uname.Name + "，我们一起聊天和下棋吧！"
	}
	// save message (non-fatal if fails)
	if msg, err := chatSvc.SaveMessage(relationID, uint(userID), senderID, greeting); err == nil {
		// push via websocket if possible
		if websocket.DefaultHub != nil {
			_ = websocket.DefaultHub.SendToUser(int(senderID), &websocket.ChatMessage{
				BaseMessage: websocket.BaseMessage{Type: websocket.MessageType(15)},
				Content:     msg.Content,
				Sender:      uname.Name,
				RelationId:  relationID,
				SenderId:    uint(userID),
				MessageId:   msg.ID,
				CreatedAt:   msg.CreatedAt.Unix(),
			})
		}
	}

	dto.SuccessResponse(c, dto.WithData(gin.H{"relationId": relationID}))
}

// DeleteFriendRequest rejects a friend request (delete DB record)
func (fc *FriendController) DeleteFriendRequest(c *gin.Context) {
	uid, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	_ = uid.(int)

	idStr := c.Param("id")
	id64, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("非法的请求ID"))
		return
	}
	reqID := uint(id64)

	if err := fc.frService.DeleteByID(reqID); err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(nil))
}
