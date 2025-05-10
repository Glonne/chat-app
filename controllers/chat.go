package controllers

import (
	"chat-app/models"
	"chat-app/queue"
	"chat-app/utils"
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var redisClient = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

//	func DeleteRoom(c *gin.Context) {
//		roomID := c.Param("room_id")
//		if err := models.DB.Delete(&models.Room{}, roomID).Error; err != nil {
//			c.JSON(http.StatusNotFound, gin.H{"error": "Room not found"})
//			return
//		}
//		c.JSON(http.StatusOK, gin.H{"message": "Room deleted"})
//	}
func CreateRoom(c *gin.Context) {
	var room models.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		log.Printf("Failed to bind JSON: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	userID := c.GetFloat64("user_id")
	room.UserID = uint(userID)
	log.Printf("Creating room: %+v", room)
	if err := models.DB.Create(&room).Error; err != nil {
		log.Printf("Failed to create room: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create room"})
		return
	}
	go StartMessageBroadcaster(fmt.Sprintf("%d", room.ID))
	log.Printf("Room created: ID=%d, Name=%s", room.ID, room.Name)
	c.JSON(http.StatusOK, gin.H{"room_id": room.ID, "name": room.Name})
}

func GetRooms(c *gin.Context) {
	var rooms []models.Room
	if err := models.DB.Find(&rooms).Error; err != nil {
		log.Printf("Failed to fetch rooms: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch rooms"})
		return
	}
	log.Printf("Fetched %d rooms: %+v", len(rooms), rooms)
	if len(rooms) == 0 {
		c.JSON(http.StatusOK, []models.Room{}) // 明确返回空数组
		return
	}
	c.JSON(http.StatusOK, rooms)
}

func HandleWebSocket(c *gin.Context) {
	roomID := c.Param("room_id")
	userID := c.GetFloat64("user_id")

	conn, err := utils.Upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		return
	}
	defer conn.Close()

	client := &utils.Client{Conn: conn, RoomID: roomID, UserID: uint(userID)}
	utils.RegisterClient(client)

	redisClient.SAdd(context.Background(), "room:"+roomID, userID)

	defer func() {
		utils.UnregisterClient(client)
		redisClient.SRem(context.Background(), "room:"+roomID, userID)
	}()

	for {
		_, msg, err := conn.ReadMessage()
		if err != nil {
			break
		}

		message := models.Message{
			Content: string(msg),
			RoomID:  mustParseUint(roomID),
			UserID:  uint(userID),
		}
		models.DB.Create(&message)
		queue.BroadcastMessage(roomID, message)
	}
}

func GetRoomMessages(c *gin.Context) {
	roomID := c.Param("room_id")
	var messages []models.Message
	models.DB.Where("room_id = ?", roomID).Order("created_at desc").Limit(50).Find(&messages)
	c.JSON(http.StatusOK, messages)
}
