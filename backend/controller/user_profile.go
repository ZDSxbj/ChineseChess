package controller

import (
	"chinese-chess-backend/dto"
	"chinese-chess-backend/dto/user"
	"chinese-chess-backend/service"

	"github.com/gin-gonic/gin"
)

func (uc *UserController) GetUserProfile(c *gin.Context) {
	// 从token中获取当前登录用户ID（需配合认证中间件）
	userID, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}

	// 调用服务层获取用户信息
	req := &user.GetUserInfoRequest{Id: userID.(int)}
	resp, err := service.NewUserService().GetUserInfo(req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(resp))
}

func (uc *UserController) UpdateUserProfile(c *gin.Context) {
	// 从token获取用户ID
	userID, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}

	// 接收前端提交的更新数据
	var req user.UpdateUserRequest
	if err := dto.BindData(c, &req); err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	// 调用服务层更新数据
	err := service.NewUserService().UpdateUserInfo(userID.(int), &req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	dto.SuccessResponse(c, dto.WithMessage("更新成功"))
}
