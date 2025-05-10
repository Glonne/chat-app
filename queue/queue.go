package queue

import (
	"chat-app/models"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/streadway/amqp"
)

var rabbitConn *amqp.Connection

func init() {
	url := os.Getenv("RABBITMQ_URL")
	if url == "" {
		// url = "amqp://guest:guest@rabbitmq:5672/"
		url = "amqp://guest:guest@localhost:5672/" // 这里改成 localhost 或 127.0.0.1
	}
	log.Println("Attempting to connect to RabbitMQ with URL:", url)

	var err error
	for i := 0; i < 60; i++ {
		rabbitConn, err = amqp.Dial(url)
		if err == nil {
			break
		}
		log.Println("Failed to connect to RabbitMQ, retrying:", err)
		time.Sleep(2 * time.Second)
	}
	if err != nil {
		log.Println("Failed to connect to RabbitMQ after retries:", err)
		return
	}
	log.Println("Successfully connected to RabbitMQ")
}

func BroadcastMessage(roomID string, message models.Message) {
	if rabbitConn == nil {
		log.Println("RabbitMQ connection is nil, cannot broadcast")
		return
	}
	ch, err := rabbitConn.Channel()
	if err != nil {
		log.Println("Failed to open channel:", err)
		return
	}
	defer ch.Close()

	exchangeName := fmt.Sprintf("room_%s_exchange", roomID)
	err = ch.ExchangeDeclare(
		exchangeName,
		"fanout", // 使用 Fanout 类型，与 StartMessageBroadcaster 一致
		true,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Println("Failed to declare exchange:", err)
		return
	}

	body, err := json.Marshal(message)
	if err != nil {
		log.Println("Failed to marshal message:", err)
		return
	}

	err = ch.Publish(
		exchangeName,
		"", // Fanout 交换机忽略路由键
		false,
		false,
		amqp.Publishing{
			ContentType: "application/json",
			Body:        body,
		},
	)
	if err != nil {
		log.Println("Failed to publish message:", err)
	} else {
		log.Println("Message broadcasted to room", roomID)
	}
}
