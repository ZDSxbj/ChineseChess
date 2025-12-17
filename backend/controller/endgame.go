package controller

import (
	"chinese-chess-backend/dto"
	endgameDto "chinese-chess-backend/dto/endgame"
	"chinese-chess-backend/service"

	"github.com/gin-gonic/gin"
)

type EndgameController struct {
	endgameService *service.EndgameService
}

func NewEndgameController(s *service.EndgameService) *EndgameController {
	return &EndgameController{
		endgameService: s,
	}
}

// GetProgress 获取当前登录用户的残局挑战进度（所有关卡）
func (ec *EndgameController) GetProgress(c *gin.Context) {
	userID := c.GetInt("userId")
	if userID == 0 {
		dto.ErrorResponse(c, dto.WithMessage("未登录"))
		return
	}

	req := &endgameDto.GetEndgameProgressRequest{UserID: userID}
	resp, err := ec.endgameService.GetProgress(req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(resp))
}

// SaveProgress 保存当前登录用户在某个残局关卡的本次挑战结果
func (ec *EndgameController) SaveProgress(c *gin.Context) {
	userID := c.GetInt("userId")
	if userID == 0 {
		dto.ErrorResponse(c, dto.WithMessage("未登录"))
		return
	}

	var req endgameDto.SaveEndgameProgressRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		dto.ErrorResponse(c, dto.WithMessage("参数错误"))
		return
	}

	if err := ec.endgameService.SaveProgress(userID, &req); err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithMessage("保存成功"))
}


