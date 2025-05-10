package controllers

import (
	"chat-app/models"
	"chat-app/queue"
	"chat-app/utils"
	"fmt"
	"log"
	"strconv"

	"github.com/gin-gonic/gin"
)

func HandleWebSocketWithAuth(c *gin.Context) {
	roomID := c.Param("room_id")
	userID := c.GetFloat64("user_id")

	conn, err := utils.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Printf("Failed to upgrade WebSocket: %v", err)
		return
	}
	defer conn.Close()

	client := &utils.Client{Conn: conn, RoomID: roomID, UserID: uint(userID)}
	utils.RegisterClient(client)
	defer utils.UnregisterClient(client)

	EnsureBroadcaster(roomID) // 确保广播器运行
	log.Printf("WebSocket connected for room %s, user %d", roomID, userID)

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			log.Printf("WebSocket read error for room %s: %v", roomID, err)
			break
		}
		log.Printf("Received message in room %s: %s", roomID, string(msg))

		message := models.Message{
			Content: string(msg),
			RoomID:  mustParseUint(roomID),
			UserID:  uint(userID),
		}
		if err := models.DB.Create(&message).Error; err != nil {
			log.Printf("Failed to save message: %v", err)
		}
		fmt.Println("websocket.go 准备queue广播")
		queue.BroadcastMessage(roomID, message)
	}
}

func mustParseUint(s string) uint {
	i, _ := strconv.ParseUint(s, 10, 32)
	return uint(i)
}
