package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"

	"chinese-chess-backend/dto"
	"chinese-chess-backend/utils"
)

func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 放行公共接口与静态资源
		if strings.HasPrefix(c.Request.URL.Path, "/api/public") ||
			strings.HasPrefix(c.Request.URL.Path, "/api/uploads") {
			c.Next()
			return
		}
		authHeader := c.GetHeader("Authorization")
		userId := utils.ParseToken(authHeader)
		if userId <= 0 {
			// 从路径中获取token
			token := c.Query("token")
			if token == "" {
				dto.ErrorResponse(c, dto.WithMessage("未登录或token错误"))
				c.Abort()
				return
			}
			userId = utils.ParseToken(token)
			if userId <= 0 {
				dto.ErrorResponse(c, dto.WithMessage("未登录或token错误"))
				c.Abort()
				return
			}
		}
		c.Set("userId", userId)
		c.Next()
	}
}
