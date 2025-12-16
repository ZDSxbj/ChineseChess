package user

type EndgameGetProgressRequest struct {
	ScenarioID string `form:"scenarioId" json:"scenarioId" binding:"required"`
}

type EndgameProgressData struct {
	Attempts  int  `json:"attempts"`
	BestSteps *int `json:"bestSteps"`
}

type EndgameGetProgressResponse struct {
	EndgameProgressData
}

type EndgameRecordProgressRequest struct {
	ScenarioID string `json:"scenarioId" binding:"required"`
	Result     string `json:"result" binding:"required"` // win/lose
	Steps      *int   `json:"steps"`                     // win 时传最小步数候选
}

type EndgameRecordProgressResponse struct {
	EndgameProgressData
}
