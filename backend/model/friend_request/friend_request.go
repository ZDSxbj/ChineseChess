package friend_request

import (
	"time"
)

type FriendRequest struct {
	ID         uint   `gorm:"primaryKey"`
	SenderID   uint   `gorm:"not null;index"`
	ReceiverID uint   `gorm:"not null;index"`
	Content    string `gorm:"type:text"`
	CreatedAt  time.Time
}
