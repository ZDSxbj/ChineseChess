package record

import "time"

// GameRecord 表示一局对局的持久化记录
type GameRecord struct {
	ID        uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	RedID     uint      `gorm:"column:red_id" json:"red_id"`
	BlackID   uint      `gorm:"column:black_id" json:"black_id"`
	StartTime time.Time `gorm:"column:start_time" json:"start_time"`
	// Result: 0 = red win, 1 = black win, 2 = draw
	Result int `gorm:"column:result" json:"result"`
	// 历史记录以 JSON 或长文本形式存储（例如前端的 Position 列表序列化）
	History   string `gorm:"type:longtext;column:history" json:"history"`
	RedFlag   bool   `gorm:"column:red_flag" json:"red_flag"`
	BlackFlag bool   `gorm:"column:black_flag" json:"black_flag"`
	// 对局类型: 0=随机匹配,1=人机,2=好友
	GameType int `gorm:"column:game_type" json:"game_type"`
	// AI难度: 1-6，仅在 game_type=1 时有效
	AILevel   int `gorm:"column:ai_level;default:3" json:"ai_level"`
	CreatedAt time.Time
	UpdatedAt time.Time
}
