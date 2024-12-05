package handlers

import (
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool { return true },
}

var clients = make(map[*websocket.Conn]*Player)
var broadcast = make(chan Message)

type Player struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Action   string  `json:"action"`
}

type Message struct {
	Type     string  `json:"type"`      // Type of message (e.g., "update", "join", "leave")
	PlayerID int     `json:"player_id"` // ID of the player sending the message
	Username string  `json:"username"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Action   string  `json:"action"`
}

var players sync.Map // Thread-safe map to store active players

// HandleWebSocketConnections handles new WebSocket connections and manages player state
func HandleWebSocketConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	// Register a new player
	player := &Player{
		ID:       len(clients) + 1, // Simplistic ID assignment; replace with proper ID logic
		Username: r.URL.Query().Get("username"),
		X:        0,
		Y:        0,
		Action:   "join",
	}
	clients[ws] = player
	players.Store(player.ID, player)

	log.Printf("New client connected: %v", player.Username)

	// Notify others about the new player
	broadcast <- Message{
		Type:     "join",
		PlayerID: player.ID,
		Username: player.Username,
		X:        player.X,
		Y:        player.Y,
		Action:   player.Action,
	}

	// Handle incoming messages
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message: %v", err)
			delete(clients, ws)
			players.Delete(player.ID)
			broadcast <- Message{
				Type:     "leave",
				PlayerID: player.ID,
				Username: player.Username,
			}
			break
		}

		// Update player state
		player.X = msg.X
		player.Y = msg.Y
		player.Action = msg.Action

		// Broadcast the update to others
		broadcast <- Message{
			Type:     "update",
			PlayerID: player.ID,
			Username: player.Username,
			X:        player.X,
			Y:        player.Y,
			Action:   player.Action,
		}
	}
}

// BroadcastMessages handles broadcasting messages to all connected clients
func BroadcastMessages() {
	for {
		msg := <-broadcast

		// Send the message to all clients
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Error writing message %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
