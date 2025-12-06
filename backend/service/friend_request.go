package service

import (
	"errors"
	"time"

	"chinese-chess-backend/database"
	dtoReq "chinese-chess-backend/dto/user"
	friendModel "chinese-chess-backend/model/friend"
	frModel "chinese-chess-backend/model/friend_request"
	userModel "chinese-chess-backend/model/user"
)

type FriendRequestService struct{}

func NewFriendRequestService() *FriendRequestService {
	return &FriendRequestService{}
}

// Create a friend request and return its model
func (frs *FriendRequestService) Create(senderID uint, receiverID uint, content string) (*frModel.FriendRequest, error) {
	db := database.GetMysqlDb()
	req := &frModel.FriendRequest{
		SenderID:   senderID,
		ReceiverID: receiverID,
		Content:    content,
		CreatedAt:  time.Now(),
	}
	if err := db.Create(req).Error; err != nil {
		return nil, errors.New("保存好友申请失败")
	}
	return req, nil
}

// List incoming friend requests for a user
func (frs *FriendRequestService) ListIncoming(userID uint) (*dtoReq.GetFriendRequestsResponse, error) {
	db := database.GetMysqlDb()
	var rows []frModel.FriendRequest
	if err := db.Where("receiver_id = ?", userID).Order("created_at desc").Find(&rows).Error; err != nil {
		return nil, errors.New("查询好友申请失败")
	}
	// build dto with sender info
	var resp dtoReq.GetFriendRequestsResponse
	resp.Requests = make([]dtoReq.FriendRequestItem, 0, len(rows))
	for _, r := range rows {
		var sender userModel.User
		_ = db.First(&sender, r.SenderID)
		resp.Requests = append(resp.Requests, dtoReq.FriendRequestItem{
			ID:           r.ID,
			SenderID:     r.SenderID,
			ReceiverID:   r.ReceiverID,
			Content:      r.Content,
			CreatedAt:    r.CreatedAt.Unix(),
			SenderName:   sender.Name,
			SenderAvatar: sender.Avatar,
		})
	}
	return &resp, nil
}

// Delete a friend request by id
func (frs *FriendRequestService) DeleteByID(id uint) error {
	db := database.GetMysqlDb()
	if err := db.Delete(&frModel.FriendRequest{}, id).Error; err != nil {
		return errors.New("删除好友申请失败")
	}
	return nil
}

// Accept a friend request: create friend relation, remove request, and return relationID
func (frs *FriendRequestService) AcceptRequest(reqID uint) (uint, error) {
	db := database.GetMysqlDb()
	var req frModel.FriendRequest
	if err := db.First(&req, reqID).Error; err != nil {
		return 0, errors.New("好友申请不存在")
	}
	// create friend relation (single row)
	relation := &friendModel.Friend{
		UserID:   req.SenderID,
		FriendID: req.ReceiverID,
	}
	if err := db.Create(relation).Error; err != nil {
		return 0, errors.New("创建好友关系失败")
	}
	// delete request
	if err := db.Delete(&frModel.FriendRequest{}, reqID).Error; err != nil {
		// non-fatal
	}
	return relation.ID, nil
}

// Exists checks whether a friend request from sender to receiver already exists
func (frs *FriendRequestService) Exists(senderID uint, receiverID uint) (bool, error) {
	db := database.GetMysqlDb()
	var cnt int64
	if err := db.Model(&frModel.FriendRequest{}).Where("sender_id = ? AND receiver_id = ?", senderID, receiverID).Count(&cnt).Error; err != nil {
		return false, errors.New("查询失败")
	}
	return cnt > 0, nil
}
