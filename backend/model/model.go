package model

import (
	"gorm.io/gorm"

	"chinese-chess-backend/model/room"
	"chinese-chess-backend/model/user"
)

func InitTable(db *gorm.DB) error {
	// 自动迁移数据库表结构
	err := db.AutoMigrate(
		&user.User{},
		&room.Room{},
	)
	if err != nil {
		return err
	}
	return nil
}
