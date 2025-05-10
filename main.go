package main

import (
	"chat-app/controllers"
	"chat-app/middleware"
	"chat-app/models"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func main() {
	models.InitDB()
	r := gin.Default()

	// 自定义 CORS 配置
	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://127.0.0.1:30080", "http://localhost:30080"}, // 指定前端地址
		AllowMethods:     []string{"GET", "POST", "OPTIONS"},                           // 允许的方法
		AllowHeaders:     []string{"Origin", "Content-Type", "Authorization"},          // 允许的头部
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,         // 支持凭证（如 Cookie）
		MaxAge:           12 * 60 * 60, // 预检请求缓存时间（12小时）
	}))

	// 根路径返回 chat.html，禁用缓存
	r.GET("/", func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache, no-store, must-revalidate")
		c.File("./static/chat.html")
	})

	// 静态文件服务
	r.Static("/static", "./static")

	// API 路由组
	api := r.Group("/api")
	{
		api.POST("/register", controllers.Register)
		api.POST("/login", controllers.Login)
		auth := api.Group("", middleware.AuthMiddleware())
		{
			auth.GET("/rooms", controllers.GetRooms)
			auth.POST("/rooms", controllers.CreateRoom)
			auth.GET("/rooms/:room_id/messages", controllers.GetRoomMessages)
			auth.GET("/ws/:room_id", controllers.HandleWebSocketWithAuth)
		}
	}

	go controllers.StartAllBroadcasters()
	r.Run(":8080")
}
