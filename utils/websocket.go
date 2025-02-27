package utils

import (
	"sync"

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
	clients   = make(map[*Client]bool)
	clientsMu sync.Mutex
	broadcast = make(chan struct{ RoomID, Msg string })
)

func RegisterClient(client *Client) {
	clientsMu.Lock()
	clients[client] = true
	clientsMu.Unlock()
}

func UnregisterClient(client *Client) {
	clientsMu.Lock()
	delete(clients, client)
	clientsMu.Unlock()
}

func BroadcastToRoom(roomID, msg string) {
	clientsMu.Lock()
	defer clientsMu.Unlock()
	for client := range clients {
		if client.RoomID == roomID {
			client.Conn.WriteMessage(websocket.TextMessage, []byte(msg))
		}
	}
}

func init() {
	go func() {
		for msg := range broadcast {
			BroadcastToRoom(msg.RoomID, msg.Msg)
		}
	}()
}
