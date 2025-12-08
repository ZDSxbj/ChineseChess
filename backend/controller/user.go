package controller

import (
	"fmt"
	"os"
	"path"
	"strings"
	"time"

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

func (uc *UserController) GetGameRecords(c *gin.Context) {
	// 从 token 中获取当前登录用户ID
	userID, exists := c.Get("userId")
	if !exists {
		dto.ErrorResponse(c, dto.WithMessage("未获取到用户信息"))
		return
	}

	req := &user.GetGameRecordsRequest{UserID: userID.(int)}
	resp, err := uc.userService.GetGameRecords(req)
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

func (uc *UserController) UpdatePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}
	userID := c.GetInt("userId")
	if userID == 0 {
		c.JSON(401, gin.H{"msg": "未登录"})
		return
	}
	if err := uc.userService.UpdatePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"msg": "修改成功"})
}

// CheckPassword 校验当前用户的原密码是否正确
func (uc *UserController) CheckPassword(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(400, gin.H{"msg": "参数错误"})
		return
	}
	userID := c.GetInt("userId")
	if userID == 0 {
		c.JSON(401, gin.H{"msg": "未登录"})
		return
	}
	if err := uc.userService.CheckPassword(userID, req.Password); err != nil {
		c.JSON(400, gin.H{"msg": err.Error()})
		return
	}
	c.JSON(200, gin.H{"msg": "ok"})
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

// UploadAvatar 处理用户头像上传，返回可在前端拼接 API 基址的相对路径
// 返回 data: { path: "/uploads/avatars/<filename>" }
func (uc *UserController) UploadAvatar(c *gin.Context) {
	userID := c.GetInt("userId")
	if userID == 0 {
		dto.ErrorResponse(c, dto.WithMessage("未登录"))
		return
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		dto.ErrorResponse(c, dto.WithMessage("缺少文件或字段名应为 avatar"))
		return
	}

	// 限制大小（例如 5MB）
	const maxSize int64 = 5 * 1024 * 1024
	if file.Size > maxSize {
		dto.ErrorResponse(c, dto.WithMessage("文件过大，最大 5MB"))
		return
	}

	// 允许的扩展名
	filename := file.Filename
	ext := strings.ToLower(path.Ext(filename))
	switch ext {
	case ".jpg", ".jpeg", ".png", ".gif", ".webp":
	default:
		dto.ErrorResponse(c, dto.WithMessage("不支持的图片格式"))
		return
	}

	// 确保保存目录存在
	saveDir := path.Join("uploads", "avatars")
	if err := os.MkdirAll(saveDir, 0755); err != nil {
		dto.ErrorResponse(c, dto.WithMessage("服务器创建目录失败"))
		return
	}

	// 用 用户ID + 时间戳 生成文件名
	newName := fmt.Sprintf("u%d_%d%s", userID, time.Now().UnixNano(), ext)
	dst := path.Join(saveDir, newName)

	if err := c.SaveUploadedFile(file, dst); err != nil {
		dto.ErrorResponse(c, dto.WithMessage("保存文件失败"))
		return
	}

	// 暴露路径
	rel := "/uploads/avatars/" + newName
	// 计算完整URL（使用环境变量 PUBLIC_API_PREFIX，默认 http://localhost:8080/api）
	apiBase := os.Getenv("PUBLIC_API_PREFIX")
	if apiBase == "" {
		apiBase = "http://localhost:8080/api"
	}
	fullURL := apiBase + rel

	// 写入用户头像为完整URL，并删除旧头像文件（若存在且非默认头像）
	db := database.GetMysqlDb()
	var user userModel.User
	if err := db.Where("id = ?", userID).First(&user).Error; err == nil {
		oldAvatar := user.Avatar
		user.Avatar = fullURL
		_ = db.Save(&user).Error

		// 删除旧文件（仅当旧头像是 /uploads/avatars/ 下的自定义文件且不是默认头像）
		if oldAvatar != "" {
			defaultAvatar := apiBase + "/uploads/avatars/default.png"
			if oldAvatar != defaultAvatar {
				if idx := strings.Index(oldAvatar, "/uploads/avatars/"); idx != -1 {
					oldRel := oldAvatar[idx:]
					// 文件系统路径
					oldFsPath := path.Join(".", strings.TrimPrefix(oldRel, "/"))
					// 尝试删除，忽略错误
					_ = os.Remove(oldFsPath)
				}
			}
		}
	}

	// 返回完整URL和相对路径，前端直接使用 url
	dto.SuccessResponse(c, dto.WithData(gin.H{"path": rel, "url": fullURL}))
}
