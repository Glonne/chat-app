package main

import (
	"chat-app/controllers"
	"chat-app/middleware"
	"chat-app/models"

	"github.com/gin-gonic/gin"
)

func main() {
	models.InitDB() // 初始化 PostgreSQL
	r := gin.Default()
	r.GET("/", func(c *gin.Context) {
		c.File("./static/chat.html")
	})
	// 公开路由
	r.POST("/register", controllers.Register)
	r.POST("/login", controllers.Login)
	r.GET("/rooms", controllers.GetRooms)

	// 认证路由
	auth := r.Group("/api", middleware.AuthMiddleware())
	{
		auth.POST("/rooms", controllers.CreateRoom)
		auth.GET("/ws/:room_id", controllers.HandleWebSocket)
		auth.GET("/rooms/:room_id/messages", controllers.GetRoomMessages)
	}

	// 启动消息广播
	go controllers.StartMessageBroadcaster()

	r.Run(":8080")
}
