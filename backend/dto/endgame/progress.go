package endgame

// GetEndgameProgressRequest 获取当前用户所有残局进度（无需额外字段）
type GetEndgameProgressRequest struct {
	UserID int `json:"user_id"`
}

// 单个关卡的进度
type ScenarioProgress struct {
	ScenarioID string `json:"scenario_id"`
	Attempts   int    `json:"attempts"`
	BestSteps  *int   `json:"best_steps,omitempty"`
	LastResult string `json:"last_result,omitempty"` // "win"/"lose"
}

// GetEndgameProgressResponse 返回当前用户所有关卡进度
type GetEndgameProgressResponse struct {
	Progress []ScenarioProgress `json:"progress"`
}

// SaveEndgameProgressRequest 用于保存某一局残局挑战结果
type SaveEndgameProgressRequest struct {
	ScenarioID string `json:"scenario_id" binding:"required"` // 残局 ID（前端 endgameData.ts 中的 id）
	Result     string `json:"result" binding:"required"`      // "win" 或 "lose"
	Steps      *int   `json:"steps,omitempty"`                // 若为 win，可选地传入完成步数
}


