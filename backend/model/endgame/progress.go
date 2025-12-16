package endgame

import "time"

// EndgameProgress 记录某个用户在某个残局关卡下的挑战进度
// 一条记录对应 (user_id, scenario_id) 的组合
type EndgameProgress struct {
	ID         uint      `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     uint      `gorm:"column:user_id;index:idx_user_scenario,priority:1" json:"user_id"`
	ScenarioID string    `gorm:"column:scenario_id;type:varchar(64);index:idx_user_scenario,priority:2" json:"scenario_id"`
	Attempts   int       `gorm:"column:attempts" json:"attempts"`        // 尝试次数
	BestSteps  *int      `gorm:"column:best_steps" json:"best_steps"`    // 最少步数（NULL 表示尚未通关）
	LastResult string    `gorm:"column:last_result;type:varchar(16)" json:"last_result"` // 上一次结果: "win" / "lose"
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}


