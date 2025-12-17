package user

type EndgameCompleteRequest struct {
	ScenarioID string `json:"scenarioId"`
	Difficulty string `json:"difficulty"` // 初级/中级/高级
	Success    bool   `json:"success"`
}

type EndgameCompleteResponse struct {
	Awarded        int  `json:"awarded"`        // 本次发放的经验
	AlreadyAwarded bool `json:"alreadyAwarded"` // 是否此前已发放过
}
