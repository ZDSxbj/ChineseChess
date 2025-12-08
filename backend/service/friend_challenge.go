package service

import (
	"errors"

	"chinese-chess-backend/database"
	challengeModel "chinese-chess-backend/model/friend_challenge"
)

type FriendChallengeService struct{}

func NewFriendChallengeService() *FriendChallengeService { return &FriendChallengeService{} }

// Create a challenge record
func (s *FriendChallengeService) Create(friendID, senderID, receiverID uint, roomID int) (*challengeModel.FriendChallenge, error) {
	db := database.GetMysqlDb()
	fc := &challengeModel.FriendChallenge{
		FriendID:   friendID,
		SenderID:   senderID,
		ReceiverID: receiverID,
		RoomID:     roomID,
	}
	if err := db.Create(fc).Error; err != nil {
		return nil, errors.New("创建挑战记录失败")
	}
	return fc, nil
}

// Delete by id
func (s *FriendChallengeService) DeleteByID(id uint) error {
	db := database.GetMysqlDb()
	if err := db.Delete(&challengeModel.FriendChallenge{}, id).Error; err != nil {
		return errors.New("删除挑战记录失败")
	}
	return nil
}

// Delete all pending challenges involving a user (as sender or receiver)
func (s *FriendChallengeService) DeleteAllByUser(userID uint) error {
	db := database.GetMysqlDb()
	if err := db.Where("sender_id = ? OR receiver_id = ?", userID, userID).Delete(&challengeModel.FriendChallenge{}).Error; err != nil {
		return err
	}
	return nil
}

// Get incoming challenges for receiver
func (s *FriendChallengeService) ListIncoming(receiverID uint) ([]challengeModel.FriendChallenge, error) {
	db := database.GetMysqlDb()
	var list []challengeModel.FriendChallenge
	if err := db.Where("receiver_id = ?", receiverID).Order("created_at desc").Find(&list).Error; err != nil {
		return nil, errors.New("获取挑战记录失败")
	}
	return list, nil
}

// Exists verifies a challenge exists by id and receiver
func (s *FriendChallengeService) Exists(id uint) (bool, error) {
	db := database.GetMysqlDb()
	var c challengeModel.FriendChallenge
	if err := db.First(&c, id).Error; err != nil {
		return false, nil
	}
	return true, nil
}
