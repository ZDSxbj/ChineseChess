package friend

import (
	"time"
)

type Friend struct {
	ID        uint `gorm:"primaryKey"`
	UserID    uint `gorm:"not null;index"`
	FriendID  uint `gorm:"not null;index"`
	CreatedAt time.Time
}
