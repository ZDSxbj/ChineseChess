package service

import (
	"chinese-chess-backend/database"
	endgameDto "chinese-chess-backend/dto/endgame"
	endgameModel "chinese-chess-backend/model/endgame"

	"errors"
)

type EndgameService struct {
}

func NewEndgameService() *EndgameService {
	return &EndgameService{}
}

// GetProgress 获取指定用户的全部残局进度
func (es *EndgameService) GetProgress(req *endgameDto.GetEndgameProgressRequest) (*endgameDto.GetEndgameProgressResponse, error) {
	if req.UserID <= 0 {
		return nil, errors.New("用户ID无效")
	}
	db := database.GetMysqlDb()

	var records []endgameModel.EndgameProgress
	if err := db.Where("user_id = ?", req.UserID).Find(&records).Error; err != nil {
		return nil, err
	}

	resp := &endgameDto.GetEndgameProgressResponse{
		Progress: make([]endgameDto.ScenarioProgress, 0, len(records)),
	}
	for _, r := range records {
		item := endgameDto.ScenarioProgress{
			ScenarioID: r.ScenarioID,
			Attempts:   r.Attempts,
			BestSteps:  r.BestSteps,
			LastResult: r.LastResult,
		}
		resp.Progress = append(resp.Progress, item)
	}
	return resp, nil
}

// SaveProgress 保存某个残局的本次挑战结果（会累加 attempts，并按需刷新 bestSteps）
func (es *EndgameService) SaveProgress(userID int, req *endgameDto.SaveEndgameProgressRequest) error {
	if userID <= 0 {
		return errors.New("未登录")
	}
	if req.ScenarioID == "" {
		return errors.New("缺少残局ID")
	}
	if req.Result != "win" && req.Result != "lose" {
		return errors.New("结果必须为 win 或 lose")
	}

	db := database.GetMysqlDb()

	var record endgameModel.EndgameProgress
	err := db.Where("user_id = ? AND scenario_id = ?", userID, req.ScenarioID).First(&record).Error
	if err != nil {
		// 若不存在则创建新纪录
		record = endgameModel.EndgameProgress{
			UserID:     uint(userID),
			ScenarioID: req.ScenarioID,
			Attempts:   1,
			LastResult: req.Result,
		}
		if req.Result == "win" && req.Steps != nil {
			val := *req.Steps
			record.BestSteps = &val
		}
		return db.Create(&record).Error
	}

	// 已存在记录则更新
	record.Attempts++
	record.LastResult = req.Result
	if req.Result == "win" && req.Steps != nil {
		steps := *req.Steps
		if record.BestSteps == nil || steps < *record.BestSteps {
			record.BestSteps = &steps
		}
	}
	return db.Save(&record).Error
}


