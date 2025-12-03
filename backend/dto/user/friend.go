package user

type FriendItem struct {
	ID         uint    `json:"id"`
	Name       string  `json:"name"`
	Avatar     string  `json:"avatar"`
	Online     bool    `json:"online"`
	Gender     string  `json:"gender"`
	Exp        int     `json:"exp"`
	TotalGames int     `json:"totalGames"`
	WinRate    float64 `json:"winRate"`
}

type GetFriendsResponse struct {
	Friends []FriendItem `json:"friends"`
}
