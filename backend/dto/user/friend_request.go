package user

type FriendRequestItem struct {
	ID           uint   `json:"id"`
	SenderID     uint   `json:"senderId"`
	ReceiverID   uint   `json:"receiverId"`
	Content      string `json:"content"`
	CreatedAt    int64  `json:"createdAt"`
	SenderName   string `json:"senderName"`
	SenderAvatar string `json:"senderAvatar"`
}

type GetFriendRequestsResponse struct {
	Requests []FriendRequestItem `json:"requests"`
}
