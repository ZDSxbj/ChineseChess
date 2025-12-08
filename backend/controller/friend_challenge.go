package controller

import (
	"github.com/gin-gonic/gin"

	dto "chinese-chess-backend/dto"
	"chinese-chess-backend/service"
)

type FriendChallengeController struct {
	fcService *service.FriendChallengeService
}

func NewFriendChallengeController() *FriendChallengeController {
	return &FriendChallengeController{fcService: service.NewFriendChallengeService()}
}

// GET /api/user/friend-challenges (incoming for current user)
func (cc *FriendChallengeController) ListIncoming(c *gin.Context) {
	uid, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}
	userID := uint(uid.(int))
	list, err := cc.fcService.ListIncoming(userID)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(gin.H{"challenges": list}))
}
