package service

import (
	"chinese-chess-backend/database"
	egp "chinese-chess-backend/model/endgame_progress"
	"errors"

	"gorm.io/gorm"
)

type EndgameService struct{}

func NewEndgameService() *EndgameService { return &EndgameService{} }

func (s *EndgameService) GetProgress(userID int, scenarioID string) (attempts int, best *int, err error) {
	db := database.GetMysqlDb()
	var rec egp.EndgameProgress
	if err = db.Where("user_id = ? AND scenario_id = ?", userID, scenarioID).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return 0, nil, nil
		}
		return 0, nil, err
	}
	return rec.Attempts, rec.BestSteps, nil
}

func (s *EndgameService) Record(userID int, scenarioID string, result string, steps *int) (attempts int, best *int, err error) {
	db := database.GetMysqlDb()
	var rec egp.EndgameProgress
	tx := db.Begin()
	if err = tx.Where("user_id = ? AND scenario_id = ?", userID, scenarioID).First(&rec).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			rec = egp.EndgameProgress{UserID: uint(userID), ScenarioID: scenarioID, Attempts: 0, BestSteps: nil}
			if err = tx.Create(&rec).Error; err != nil {
				tx.Rollback()
				return
			}
		} else {
			tx.Rollback()
			return
		}
	}
	rec.Attempts += 1
	// 只有胜利且提供 steps 时才尝试刷新最优
	if result == "win" && steps != nil {
		if rec.BestSteps == nil || (steps != nil && *steps < *rec.BestSteps) {
			rec.BestSteps = steps
		}
	}
	if err = tx.Save(&rec).Error; err != nil {
		tx.Rollback()
		return
	}
	if err = tx.Commit().Error; err != nil {
		return
	}
	return rec.Attempts, rec.BestSteps, nil
}
