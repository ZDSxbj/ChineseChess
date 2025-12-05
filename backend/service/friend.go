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

	// map otherUserID -> relationID
	otherToRelation := make(map[uint]uint)
	for _, r := range relations {
		if int(r.UserID) == userID {
			otherToRelation[r.FriendID] = r.ID
		} else {
			otherToRelation[r.UserID] = r.ID
		}
	}

	var friendIDs []uint
	for id := range otherToRelation {
		friendIDs = append(friendIDs, id)
	}

	var users []userModel.User
	if len(friendIDs) > 0 {
		if err := db.Where("id IN ?", friendIDs).Find(&users).Error; err != nil {
			return nil, errors.New("查询好友信息失败")
		}
	}

	// 获取未读统计
	chatSvc := NewChatService()
	unreadMap, _ := chatSvc.GetUnreadCounts(uint(userID))

	resp := &dto.GetFriendsResponse{Friends: make([]dto.FriendItem, 0, len(users))}
	for _, u := range users {
		rel := otherToRelation[u.ID]
		uc := int64(0)
		if v, ok := unreadMap[rel]; ok {
			uc = v
		}
		resp.Friends = append(resp.Friends, dto.FriendItem{
			ID:          u.ID,
			RelationID:  rel,
			Name:        u.Name,
			Avatar:      u.Avatar,
			Online:      u.Online,
			Gender:      u.Gender,
			Exp:         u.Exp,
			TotalGames:  u.TotalGames,
			WinRate:     u.WinRate,
			UnreadCount: uc,
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
