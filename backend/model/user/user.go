package user

import (
	// "gorm.io/gorm"
	"time"
)

type User struct {
	ID         uint    `gorm:"primaryKey"`
	Name       string  `gorm:"type:varchar(100);not null"`             // 姓名
	Avatar     string  `gorm:"type:varchar(255);default:''"`           // 头像URL
	Gender     string  `gorm:"type:varchar(10);default:''"`            // 性别（男/女/其他）
	Email      string  `gorm:"type:varchar(100);uniqueIndex;not null"` // 邮箱
	Password   string  `gorm:"type:varchar(100);not null"`             // 密码（哈希存储）
	Online     bool    `gorm:"default:false;not null"`                 // 在线状态
	Exp        int     `gorm:"default:0"`                              // 经验值
	TotalGames int     `gorm:"default:0"`                              // 总场次
	WinRate    float64 `gorm:"default:0"`                              // 胜率（百分比）
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
