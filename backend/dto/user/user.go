package user

type UserInfo struct {
	ID         uint    `json:"id"`
	Token      string  `json:"token"`
	Name       string  `json:"name"`
	Email      string  `json:"email"`
	Avatar     string  `json:"avatar"`
	Gender     string  `json:"gender"`
	Exp        int     `json:"exp"`
	TotalGames int     `json:"totalGames"`
	WinRate    float64 `json:"winRate"`
}
