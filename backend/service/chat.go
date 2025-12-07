package service

import (
	"errors"
	"fmt"
	"log"

	"chinese-chess-backend/database"
	chatModel "chinese-chess-backend/model/chat"
)

type ChatService struct{}

func NewChatService() *ChatService {
	return &ChatService{}
}

// Save a chat message to DB
func (cs *ChatService) SaveMessage(friendRelationID uint, senderID uint, receiverID uint, content string) (*chatModel.ChatMessage, error) {
	db := database.GetMysqlDb()
	msg := &chatModel.ChatMessage{
		FriendID:   friendRelationID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		IsRead:     false,
	}
	if err := db.Create(msg).Error; err != nil {
		// 记录并返回底层错误，便于定位问题
		log.Printf("SaveMessage: failed to create chat message: %v", err)
		return nil, fmt.Errorf("保存消息失败: %w", err)
	}
	return msg, nil
}

// Get messages by friend relation id, ordered by creation asc. limit/offset for pagination
func (cs *ChatService) GetMessagesByRelation(relationID uint, limit int, offset int) ([]chatModel.ChatMessage, error) {
	db := database.GetMysqlDb()
	var msgs []chatModel.ChatMessage
	q := db.Where("friend_id = ?", relationID).Order("created_at asc")
	if limit > 0 {
		q = q.Limit(limit).Offset(offset)
	}
	if err := q.Find(&msgs).Error; err != nil {
		return nil, errors.New("查询聊天记录失败")
	}
	return msgs, nil
}

// Mark all messages in a relation as read for the given receiver
func (cs *ChatService) MarkRead(relationID uint, receiverID uint) error {
	db := database.GetMysqlDb()
	if err := db.Model(&chatModel.ChatMessage{}).Where("friend_id = ? AND receiver_id = ? AND is_read = ?", relationID, receiverID, false).Update("is_read", true).Error; err != nil {
		return errors.New("标记已读失败")
	}
	return nil
}

// Get unread counts grouped by friend relation id for a user (as receiver)
func (cs *ChatService) GetUnreadCounts(userID uint) (map[uint]int64, error) {
	db := database.GetMysqlDb()
	type result struct {
		FriendID uint
		Count    int64
	}
	var rows []result
	if err := db.Model(&chatModel.ChatMessage{}).Select("friend_id as friend_id, count(*) as count").Where("receiver_id = ? AND is_read = ?", userID, false).Group("friend_id").Scan(&rows).Error; err != nil {
		return nil, errors.New("统计未读消息失败")
	}
	m := make(map[uint]int64)
	for _, r := range rows {
		m[r.FriendID] = r.Count
	}
	// convert to int64 map (caller may expect ints)
	res := make(map[uint]int64)
	for k, v := range m {
		res[k] = v
	}
	return res, nil
}
