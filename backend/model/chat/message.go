package chat

import (
	"time"
)

type ChatMessage struct {
	ID       uint `gorm:"primaryKey"`
	FriendID uint `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SenderID   uint   `gorm:"not null;index"`
	ReceiverID uint   `gorm:"not null;index"`
	IsRead     bool   `gorm:"default:false"`
	Content    string `gorm:"type:text;"`
	CreatedAt  time.Time
}
