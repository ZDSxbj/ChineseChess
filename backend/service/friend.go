package service

import (
	"chinese-chess-backend/database"
	dto "chinese-chess-backend/dto/user"
	friendModel "chinese-chess-backend/model/friend"
	userModel "chinese-chess-backend/model/user"
	"errors"
)

type FriendService struct{}

func NewFriendService() *FriendService {
	return &FriendService{}
}

// List friends for given user id
func (fs *FriendService) ListFriends(userID int) (*dto.GetFriendsResponse, error) {
	db := database.GetMysqlDb()
	var relations []friendModel.Friend
	if err := db.Where("user_id = ? OR friend_id = ?", userID, userID).Find(&relations).Error; err != nil {
		return nil, errors.New("查询好友失败")
	}

	// 收集好友 id
	friendIDsMap := make(map[uint]struct{})
	for _, r := range relations {
		if int(r.UserID) == userID {
			friendIDsMap[r.FriendID] = struct{}{}
		} else {
			friendIDsMap[r.UserID] = struct{}{}
		}
	}

	var friendIDs []uint
	for id := range friendIDsMap {
		friendIDs = append(friendIDs, id)
	}

	var users []userModel.User
	if len(friendIDs) > 0 {
		if err := db.Where("id IN ?", friendIDs).Find(&users).Error; err != nil {
			return nil, errors.New("查询好友信息失败")
		}
	}

	resp := &dto.GetFriendsResponse{Friends: make([]dto.FriendItem, 0, len(users))}
	for _, u := range users {
		resp.Friends = append(resp.Friends, dto.FriendItem{
			ID:         u.ID,
			Name:       u.Name,
			Avatar:     u.Avatar,
			Online:     u.Online,
			Gender:     u.Gender,
			Exp:        u.Exp,
			TotalGames: u.TotalGames,
			WinRate:    u.WinRate,
		})
	}

	return resp, nil
}

// Delete friend relation (either direction)
func (fs *FriendService) DeleteFriend(userID int, friendID int) error {
	db := database.GetMysqlDb()

	// 删除任意方向的好友关系条目
	if err := db.Where("(user_id = ? AND friend_id = ?) OR (user_id = ? AND friend_id = ?)", userID, friendID, friendID, userID).Delete(&friendModel.Friend{}).Error; err != nil {
		return errors.New("删除好友失败")
	}
	return nil
}
