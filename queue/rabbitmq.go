package queue

import (
	"chat-app/models"
	"fmt"
	"log"

	"github.com/streadway/amqp"
)

func BroadcastMessage(roomID string, message models.Message) {
	conn, err := amqp.Dial("amqp://guest:guest@localhost:5672/")
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

	q, _ := ch.QueueDeclare("chat_messages", false, false, false, false, nil)
	msg := fmt.Sprintf("%s:%s", roomID, message.Content)
	ch.Publish("", q.Name, false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        []byte(msg),
	})
}
