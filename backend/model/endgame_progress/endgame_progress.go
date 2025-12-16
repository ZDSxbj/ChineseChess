package endgame_progress

import "time"

type EndgameProgress struct {
	ID uint `gorm:"primaryKey"`
	// 复合唯一索引：同一用户在同一关卡仅一条记录
	UserID     uint   `gorm:"not null;index:idx_user_scenario,unique"`
	ScenarioID string `gorm:"type:varchar(100);not null;index:idx_user_scenario,unique"`
	Attempts   int    `gorm:"default:0"`
	BestSteps  *int   `gorm:"default:null"`
	CreatedAt  time.Time
	UpdatedAt  time.Time
}
