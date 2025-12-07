package route

import (
	"os"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"chinese-chess-backend/controller"
	"chinese-chess-backend/service"

	"chinese-chess-backend/middleware"
	"chinese-chess-backend/websocket"
)

func SetupRouter() *gin.Engine {
	r := gin.Default()

	// 支持多个前端地址，通过环境变量 FRONTEND_URL 指定，多个地址用逗号分隔
	// 开发默认支持 Vite 的 5173 和另一个常见的 5174
	originsEnv := os.Getenv("FRONTEND_URL")
	var origins []string
	if originsEnv == "" {
		origins = []string{"http://localhost:5173", "http://localhost:5174"}
	} else {
		// 支持逗号分隔的多个地址
		for _, p := range strings.Split(originsEnv, ",") {
			if s := strings.TrimSpace(p); s != "" {
				origins = append(origins, s)
			}
		}
	}

	// 设置跨域请求
	r.Use(cors.New(cors.Config{
		AllowOrigins:     origins,
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))
	// r.Use(middleware.CorsMiddleware())
	r.Use(middleware.AuthMiddleware())

	user := controller.NewUserController(service.NewUserService())
	friend := controller.NewFriendController(service.NewFriendService())
	chat := controller.NewChatController(service.NewChatService())
	room := controller.NewRoomController(service.NewRoomService())
	// 设置路由组
	api := r.Group("/api")
	// 静态资源：通过 /api/uploads 访问后端本地的 ./uploads 目录
	// 这样配合 nginx 的 /chess/api/ 反向代理即可在生产环境访问
	api.Static("/uploads", "./uploads")
	api.POST("/info", user.GetUserInfo)
	// userRoute := api.Group("/user")

	publicRoute := api.Group("/public")
	publicRoute.POST("/register", user.Register)
	publicRoute.POST("/login", user.Login)
	publicRoute.POST("/send-code", user.SendVCode)

	userRoute := api.Group("/user")
	userRoute.GET("/profile", user.GetUserProfile)
	userRoute.POST("/profile", user.UpdateUserProfile)
	userRoute.POST("/avatar", user.UploadAvatar)
	userRoute.POST("/update_email", user.UpdateEmail)
	userRoute.POST("/update_password", user.UpdatePassword)
	userRoute.POST("/check_password", user.CheckPassword)
	userRoute.POST("/delete_account", user.DeleteAccount)
	userRoute.PUT("/profile", user.UpdateUserProfile)
	userRoute.GET("/friends", friend.GetFriends)
	userRoute.GET("/friend-requests", friend.GetFriendRequests)
	// 接受或拒绝好友申请
	userRoute.POST("/friend-requests/:id/accept", friend.AcceptFriendRequest)
	userRoute.DELETE("/friend-requests/:id", friend.DeleteFriendRequest)
	userRoute.GET("/friend-requests/check", friend.CheckFriendRequest)
	userRoute.DELETE("/friends/:friendId", friend.DeleteFriend)

	// 聊天相关路由
	userRoute.GET("/friends/:relationId/messages", chat.GetMessages)
	userRoute.POST("/friends/:relationId/messages", chat.SendMessage)
	userRoute.POST("/friends/:relationId/mark-read", chat.MarkRead)

	hub := websocket.NewChessHub()
	userRoute.POST("/rooms", hub.GetSpareRooms, room.GetSpareRooms)
	userRoute.GET("/game-records", user.GetGameRecords)
	r.GET("/ws", hub.HandleConnection)
	go hub.Run()

	return r
}
