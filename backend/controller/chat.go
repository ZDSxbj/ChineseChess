package controller

import (
	"strconv"
	"time"

	"github.com/gin-gonic/gin"

	"chinese-chess-backend/database"
	dto "chinese-chess-backend/dto"
	friendModel "chinese-chess-backend/model/friend"
	chatService "chinese-chess-backend/service"
	"chinese-chess-backend/websocket"
)

type ChatController struct {
	chatService *chatService.ChatService
}

func NewChatController(chatService *chatService.ChatService) *ChatController {
	return &ChatController{chatService: chatService}
}

// GET /api/user/friends/:relationId/messages
func (cc *ChatController) GetMessages(c *gin.Context) {
	uid, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	userID := uid.(int)

	relStr := c.Param("relationId")
	relID64, err := strconv.ParseUint(relStr, 10, 64)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("非法的关系ID"))
		return
	}
	relID := uint(relID64)

	// 验证当前用户属于该好友关系
	db := database.GetMysqlDb()
	var rel friendModel.Friend
	if err := db.First(&rel, relID).Error; err != nil {
		dto.ErrorResponse(c, dto.WithMessage("好友关系不存在"))
		return
	}
	if int(rel.UserID) != userID && int(rel.FriendID) != userID {
		dto.ErrorResponse(c, dto.WithMessage("无权限查看该聊天记录"))
		return
	}

	// 简单分页参数
	limit := 100
	offset := 0

	msgs, err := cc.chatService.GetMessagesByRelation(relID, limit, offset)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	// 将后端的字段名转换为前端期望的小驼峰格式（createdAt 等）
	type msgResp struct {
		ID         uint   `json:"id"`
		FriendId   uint   `json:"friendId"`
		SenderId   uint   `json:"senderId"`
		ReceiverId uint   `json:"receiverId"`
		IsRead     bool   `json:"isRead"`
		Content    string `json:"content"`
		CreatedAt  string `json:"createdAt"`
	}
	out := make([]msgResp, 0, len(msgs))
	for _, m := range msgs {
		out = append(out, msgResp{
			ID:         m.ID,
			FriendId:   m.FriendID,
			SenderId:   m.SenderID,
			ReceiverId: m.ReceiverID,
			IsRead:     m.IsRead,
			Content:    m.Content,
			CreatedAt:  m.CreatedAt.Format(time.RFC3339),
		})
	}

	dto.SuccessResponse(c, dto.WithData(out))
}

type sendReq struct {
	Content string `json:"content" binding:"required"`
}

// POST /api/user/friends/:relationId/messages
func (cc *ChatController) SendMessage(c *gin.Context) {
	uid, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	userID := uid.(int)

	relStr := c.Param("relationId")
	relID64, err := strconv.ParseUint(relStr, 10, 64)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("非法的关系ID"))
		return
	}
	relID := uint(relID64)

	var req sendReq
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, dto.WithMessage("消息内容不能为空"))
		return
	}

	// 验证关系并确定接收方
	db := database.GetMysqlDb()
	var rel friendModel.Friend
	if err := db.First(&rel, relID).Error; err != nil {
		dto.ErrorResponse(c, dto.WithMessage("好友关系不存在"))
		return
	}
	var receiver uint
	if int(rel.UserID) == userID {
		receiver = rel.FriendID
	} else if int(rel.FriendID) == userID {
		receiver = rel.UserID
	} else {
		dto.ErrorResponse(c, dto.WithMessage("无权限发送消息"))
		return
	}

	msg, err := cc.chatService.SaveMessage(relID, uint(userID), receiver, req.Content)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	// 尝试通过 WebSocket 推送给接收者（若在线）
	if websocket.DefaultHub != nil {
		// 查询发送者名字
		var senderName string
		var u struct{ Name string }
		if err := db.Table("user").Select("name").Where("id = ?", userID).Take(&u).Error; err == nil {
			senderName = u.Name
		}
		// 使用 websocket 的 ChatMessage 结构，type 使用数值 15
		_ = websocket.DefaultHub.SendToUser(int(receiver), &websocket.ChatMessage{
			BaseMessage: websocket.BaseMessage{Type: websocket.MessageType(15)},
			Content:     msg.Content,
			Sender:      senderName,
			RelationId:  relID,
			SenderId:    uint(userID),
			MessageId:   msg.ID,
			CreatedAt:   msg.CreatedAt.Unix(),
		})
	}

	// 返回小驼峰格式，便于前端直接使用
	resp := struct {
		ID         uint   `json:"id"`
		FriendId   uint   `json:"friendId"`
		SenderId   uint   `json:"senderId"`
		ReceiverId uint   `json:"receiverId"`
		IsRead     bool   `json:"isRead"`
		Content    string `json:"content"`
		CreatedAt  string `json:"createdAt"`
	}{
		ID:         msg.ID,
		FriendId:   msg.FriendID,
		SenderId:   msg.SenderID,
		ReceiverId: msg.ReceiverID,
		IsRead:     msg.IsRead,
		Content:    msg.Content,
		CreatedAt:  msg.CreatedAt.Format(time.RFC3339),
	}

	dto.SuccessResponse(c, dto.WithData(resp))
}

// POST /api/user/friends/:relationId/mark-read
func (cc *ChatController) MarkRead(c *gin.Context) {
	uid, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	userID := uid.(int)

	relStr := c.Param("relationId")
	relID64, err := strconv.ParseUint(relStr, 10, 64)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("非法的关系ID"))
		return
	}
	relID := uint(relID64)

	// 验证权限
	db := database.GetMysqlDb()
	var rel friendModel.Friend
	if err := db.First(&rel, relID).Error; err != nil {
		dto.ErrorResponse(c, dto.WithMessage("好友关系不存在"))
		return
	}
	if int(rel.UserID) != userID && int(rel.FriendID) != userID {
		dto.ErrorResponse(c, dto.WithMessage("无权限操作"))
		return
	}

	if err := cc.chatService.MarkRead(relID, uint(userID)); err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(nil))
}
