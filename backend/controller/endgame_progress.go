package controller

import (
	"github.com/gin-gonic/gin"

	"chinese-chess-backend/dto"
	userdto "chinese-chess-backend/dto/user"
	"chinese-chess-backend/service"
)

// EndgameGetProgress 获取当前用户某关卡的尝试次数与最小步数（包级处理函数，便于路由绑定）
func EndgameGetProgress(c *gin.Context) {
	userID := c.GetInt("userId")
	if userID == 0 {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	var req userdto.EndgameGetProgressRequest
	if err := c.ShouldBindQuery(&req); err != nil {
		dto.ErrorResponse(c, dto.WithMessage("参数错误"))
		return
	}
	svc := service.NewEndgameService()
	attempts, best, err := svc.GetProgress(userID, req.ScenarioID)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(userdto.EndgameGetProgressResponse{EndgameProgressData: userdto.EndgameProgressData{Attempts: attempts, BestSteps: best}}))
}

// EndgameRecordProgress 记录一局完成（胜/负）并更新尝试次数与最小步数（包级处理函数）
func EndgameRecordProgress(c *gin.Context) {
	userID := c.GetInt("userId")
	if userID == 0 {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	var req userdto.EndgameRecordProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, dto.WithMessage("参数错误"))
		return
	}
	svc := service.NewEndgameService()
	attempts, best, err := svc.Record(userID, req.ScenarioID, req.Result, req.Steps)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(userdto.EndgameRecordProgressResponse{EndgameProgressData: userdto.EndgameProgressData{Attempts: attempts, BestSteps: best}}))
}
