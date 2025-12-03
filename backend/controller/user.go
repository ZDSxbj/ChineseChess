package controller

import (
	"github.com/gin-gonic/gin"

	"chinese-chess-backend/database"
	"chinese-chess-backend/dto"
	"chinese-chess-backend/dto/user"
	userModel "chinese-chess-backend/model/user"
	"chinese-chess-backend/service"
)

type UserController struct {
	userService *service.UserService
}

func NewUserController(userService *service.UserService) *UserController {
	return &UserController{
		userService: userService,
	}
}

func (uc *UserController) Register(c *gin.Context) {
	var req user.RegisterRequest
	err := dto.BindData(c, &req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	resp, err := uc.userService.Register(&req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(resp))

}

func (uc *UserController) Login(c *gin.Context) {
	var req user.LoginRequest
	err := dto.BindData(c, &req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	resp, err := uc.userService.Login(&req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(resp))
}

func (uc *UserController) SendVCode(c *gin.Context) {
	var req user.SendVCodeRequest
	err := dto.BindData(c, &req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	err = uc.userService.SendVCode(&req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithMessage("验证码发送成功"))
}

func (uc *UserController) GetUserInfo(c *gin.Context) {
	var req user.GetUserInfoRequest
	err := dto.BindData(c, &req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}

	resp, err := uc.userService.GetUserInfo(&req)
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage(err.Error()))
		return
	}
	dto.SuccessResponse(c, dto.WithData(resp))
}

func (uc *UserController) UpdateEmail(c *gin.Context) {
	var req struct {
		Email string `json:"email"`
		Code  string `json:"code"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}
	// 假设已通过中间件获取 userID
	userID := c.GetInt("userId")
	err := uc.userService.UpdateEmailWithCode(userID, req.Email, req.Code)
	if err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"msg": "修改成功"})
}

func (uc *UserController) DeleteAccount(c *gin.Context) {
	userID := c.GetInt("userId")
	if userID == 0 {
		c.JSON(401, gin.H{"msg": "未登录"})
		return
	}
	db := database.GetMysqlDb()
	if err := db.Where("id = ?", userID).Delete(&userModel.User{}).Error; err != nil {
		c.JSON(500, gin.H{"msg": "注销失败"})
		return
	}
	c.JSON(200, gin.H{"msg": "注销成功"})
}
