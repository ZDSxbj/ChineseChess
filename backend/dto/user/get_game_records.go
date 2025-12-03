package user

import (
	"fmt"
	"time"
)

type GetGameRecordsRequest struct {
	UserID int `json:"user_id"`
}

func (r *GetGameRecordsRequest) Examine() error {
	if r.UserID <= 0 {
		return fmt.Errorf("用户ID无效")
	}
	return nil
}

type GameRecordItem struct {
	ID           uint   `json:"id"`
	OpponentID   uint   `json:"opponent_id"`
	OpponentName string `json:"opponent_name"`
	// Result: 0=win, 1=lose, 2=draw
	Result int `json:"result"`
	// GameType: 0=随机匹配, 1=人机对战, 2=好友对战
	GameType int `json:"game_type"`
	// IsRed: true=红方, false=黑方
	IsRed      bool      `json:"is_red"`
	TotalSteps int       `json:"total_steps"`
	History    string    `json:"history"`
	StartTime  time.Time `json:"start_time"`
}

type GetGameRecordsResponse struct {
	Records []GameRecordItem `json:"records"`
}
