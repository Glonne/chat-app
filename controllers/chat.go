package controllers

import (
	"bytes"
	"chat-app/models"
	"chat-app/queue"
	"chat-app/utils"
	"context"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	"github.com/streadway/amqp"
)

var redisClient = redis.NewClient(&redis.Options{Addr: "localhost:6379"})

func CreateRoom(c *gin.Context) {
	var room models.Room
	if err := c.ShouldBindJSON(&room); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	userID := c.GetFloat64("user_id")
	room.UserID = uint(userID)
	models.DB.Create(&room)

	c.JSON(http.StatusOK, gin.H{"room_id": room.ID, "name": room.Name})
}

func GetRooms(c *gin.Context) {
	var rooms []models.Room
	models.DB.Find(&rooms)
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

func StartMessageBroadcaster() {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	ch, _ := conn.Channel()
	defer ch.Close()

	q, _ := ch.QueueDeclare("chat_messages", false, false, false, false, nil)
	msgs, _ := ch.Consume(q.Name, "", true, false, false, false, nil)

	for msg := range msgs {
		// 从消息中解析 roomID 和内容
		roomID := string(msg.Body[:bytes.IndexByte(msg.Body, ':')])
		content := string(msg.Body[bytes.IndexByte(msg.Body, ':')+1:])
		utils.BroadcastToRoom(roomID, content)
	}
}

func mustParseUint(s string) uint {
	i, _ := strconv.ParseUint(s, 10, 32)
	return uint(i)
}
