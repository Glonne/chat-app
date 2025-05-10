package controllers

import (
	"chat-app/models"
	"chat-app/utils"
	"context"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/streadway/amqp"
)

var (
	broadcasters   = make(map[string]bool) // 跟踪运行中的广播器
	broadcastersMu sync.Mutex
)

func StartMessageBroadcaster(roomID string) {
	fmt.Println("启动消息广播器")
	broadcastersMu.Lock()
	if broadcasters[roomID] {
		broadcastersMu.Unlock()
		return // 已存在广播器
	}
	broadcasters[roomID] = true
	broadcastersMu.Unlock()
	conn, err := amqp.Dial("amqp://guest:guest@rabbitmq:5672/")
	//conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
	if err != nil {
		log.Println("Failed to connect to RabbitMQ:", err)
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		log.Println("Failed to open channel:", err)
		return
	}
	defer ch.Close()

	exchangeName := fmt.Sprintf("room_%s_exchange", roomID)
	err = ch.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	if err != nil {
		log.Println("Failed to declare exchange:", err)
		return
	}
	fmt.Println("exchange成功")
	// 动态绑定队列
	q, err := ch.QueueDeclare("", false, true, false, false, nil)
	if err != nil {
		log.Println("Failed to declare queue:", err)
		return
	}
	err = ch.QueueBind(q.Name, "", exchangeName, false, nil)
	if err != nil {
		log.Println("Failed to bind queue:", err)
		return
	}
	fmt.Println("绑定队列成功")
	msgs, err := ch.Consume(q.Name, "", true, false, false, false, nil)
	if err != nil {
		log.Println("Failed to consume:", err)
		return
	}
	fmt.Println("有消息传来了")
	log.Printf("Started broadcaster for room %s, queue %s", roomID, q.Name)
	// for msg := range msgs {
	// 	log.Printf("Broadcasting message in room %s: %s", roomID, string(msg.Body))
	// 	utils.BroadcastToRoom(roomID, string(msg.Body))
	// }
	for msg := range msgs {
		var message models.Message
		if err := json.Unmarshal(msg.Body, &message); err != nil {
			log.Printf("Failed to unmarshal message: %v", err)
			continue
		}
		log.Printf("Broadcasting message in room %s: %s", roomID, message.Content)
		utils.BroadcastToRoom(roomID, message.Content)
	}
	broadcastersMu.Lock()
	delete(broadcasters, roomID) // 广播器退出时清理
	broadcastersMu.Unlock()
}

func StartAllBroadcasters() {
	acquired, err := redisClient.SetNX(context.Background(), "broadcasters_lock", "1", 10*time.Minute).Result()
	if err != nil || !acquired {
		log.Println("Not master node, skipping broadcasters")
		return
	}
	var rooms []models.Room
	fmt.Println("启动全部消息广播器")
	models.DB.Find(&rooms)
	for _, room := range rooms {
		go StartMessageBroadcaster(fmt.Sprintf("%d", room.ID))
	}
}

// 在用户进入房间时确保广播器运行
func EnsureBroadcaster(roomID string) {
	go StartMessageBroadcaster(roomID)
}
