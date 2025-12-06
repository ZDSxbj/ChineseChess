package model

import (
	"gorm.io/gorm"

	"chinese-chess-backend/model/chat"
	"chinese-chess-backend/model/friend"
	friendrequest "chinese-chess-backend/model/friend_request"
	"chinese-chess-backend/model/record"
	"chinese-chess-backend/model/user"
)

func InitTable(db *gorm.DB) error {
	// 自动迁移数据库表结构
	err := db.AutoMigrate(
		&user.User{},
		&record.GameRecord{},
		&friend.Friend{},
		&chat.ChatMessage{},
		&friendrequest.FriendRequest{},
	)
	if err != nil {
		return err
	}
	return nil
}
