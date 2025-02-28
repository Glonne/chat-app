package main

import (
    "github.com/gin-gonic/gin"
    "chat-app/controllers"
    "chat-app/middleware"
    "chat-app/models"
)

func main() {
    models.InitDB()
    r := gin.Default()

    r.GET("/", func(c *gin.Context) {
        c.File("./static/chat.html")
    })

    r.POST("/register", controllers.Register)
    r.POST("/login", controllers.Login)
    r.GET("/rooms", controllers.GetRooms)
    auth := r.Group("/api", middleware.AuthMiddleware())
    {
        auth.POST("/rooms", controllers.CreateRoom)
        auth.GET("/rooms/:room_id/messages", controllers.GetRoomMessages)
    }
    // WebSocket 不加中间件
    r.GET("/api/ws/:room_id", controllers.HandleWebSocket)

    go controllers.StartMessageBroadcaster()
    r.Run(":8080")
}
