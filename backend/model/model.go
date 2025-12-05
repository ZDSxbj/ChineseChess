package model

import (
	"gorm.io/gorm"

	"chinese-chess-backend/model/chat"
	"chinese-chess-backend/model/friend"
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
	)
	if err != nil {
		return err
	}
	return nil
}
