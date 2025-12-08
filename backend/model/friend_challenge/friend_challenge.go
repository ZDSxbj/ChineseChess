package friend_challenge

import "time"

type FriendChallenge struct {
	ID         uint `gorm:"primaryKey"`
	FriendID   uint `gorm:"not null;index;constraint:OnUpdate:CASCADE,OnDelete:CASCADE;"`
	SenderID   uint `gorm:"not null;index"`
	ReceiverID uint `gorm:"not null;index"`
	RoomID     int  `gorm:"not null"`
	CreatedAt  time.Time
}
