package main

import (
	"chinese-chess-backend/config"
	"chinese-chess-backend/database"
	userModel "chinese-chess-backend/model/user"
	"chinese-chess-backend/route"
	"log"
	"time"
)

func main() {
	config.InitConfig()
	// 应用启动时，将所有用户在线状态重置为离线，避免历史脏数据导致无法登录
	func() {
		defer func() { recover() }()
		db := database.GetMysqlDb()
		if err := db.Model(&userModel.User{}).Where("online = ?", true).Update("online", false).Error; err != nil {
			log.Printf("reset online status failed: %v", err)
		}
	}()
	r := route.SetupRouter()

	// 心跳过期检测：每60秒扫描，超过2分钟未心跳的用户自动置为离线
	go func() {
		ticker := time.NewTicker(60 * time.Second)
		defer ticker.Stop()
		timeout := 2 * time.Minute
		for range ticker.C {
			db := database.GetMysqlDb()
			expireBefore := time.Now().Add(-timeout)
			if err := db.Model(&userModel.User{}).
				Where("online = ? AND (last_active_at IS NULL OR last_active_at < ?)", true, expireBefore).
				Update("online", false).Error; err != nil {
				log.Printf("heartbeat expiry scan failed: %v", err)
			}
		}
	}()

	r.Run(":8080")
}
