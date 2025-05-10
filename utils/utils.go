package utils

import (
	"context"
	"fmt"
	"log"
	"strconv"
	"sync"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/websocket"
)

var Upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
}

type Client struct {
	Conn   *websocket.Conn
	RoomID string
	UserID uint
}

var (
	Clients     = make(map[*Client]bool)         // 导出本地缓存
	Connections = make(map[uint]*websocket.Conn) // 导出本地连接池
	ClientsMu   sync.Mutex                       // 导出锁
	redisClient = redis.NewClient(&redis.Options{Addr: "redis:6379"})
)

func RegisterClient(client *Client) {
	ClientsMu.Lock()
	Clients[client] = true
	Connections[client.UserID] = client.Conn
	ClientsMu.Unlock()
	err := redisClient.HSet(context.Background(), "clients:"+client.RoomID, fmt.Sprintf("%d", client.UserID), "active").Err()
	if err != nil {
		log.Printf("Failed to register client in Redis for room %s, user %d: %v", client.RoomID, client.UserID, err)
	} else {
		log.Printf("Client registered in Redis for room %s, user %d", client.RoomID, client.UserID)
	}
}
func UnregisterClient(client *Client) {
	ClientsMu.Lock()
	delete(Clients, client)
	delete(Connections, client.UserID)
	ClientsMu.Unlock()
	redisClient.SRem(context.Background(), "clients:"+client.RoomID, client.UserID)
}

func BroadcastToRoom(roomID, msg string) {
	users, err := redisClient.HGetAll(context.Background(), "clients:"+roomID).Result()
	if err != nil {
		log.Printf("Failed to get clients from Redis for room %s: %v", roomID, err)
		return
	}
	if len(users) == 0 {
		log.Printf("No clients found in Redis for room %s", roomID)
	}

	ClientsMu.Lock()
	defer ClientsMu.Unlock()
	for userIDStr := range users {
		uid, err := strconv.ParseUint(userIDStr, 10, 32)
		if err != nil {
			log.Printf("Invalid user ID %s in room %s: %v", userIDStr, roomID, err)
			continue
		}
		if conn, ok := Connections[uint(uid)]; ok {
			err := conn.WriteMessage(websocket.TextMessage, []byte(msg))
			if err != nil {
				log.Printf("Failed to send message to user %d in room %s: %v", uid, roomID, err)
				conn.Close()
				delete(Connections, uint(uid))
			} else {
				log.Printf("Message sent to user %d in room %s: %s", uid, roomID, msg)
			}
		} else {
			log.Printf("No connection found for user %d in room %s", uid, roomID)
		}
	}
}
