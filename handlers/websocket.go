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

var clients = make(map[*websocket.Conn]int) // Maps WebSocket connection to Player ID
var players sync.Map                        // Thread-safe map to store active players
var broadcast = make(chan Message)          // Channel for broadcasting messages

type Player struct {
	ID       int     `json:"id"`
	Username string  `json:"username"`
	X        float64 `json:"x"`
	Y        float64 `json:"y"`
	Action   string  `json:"action"`
}

type Message struct {
	Type     string  `json:"type"`      // Type of message: "update", "join", "leave"
	PlayerID int     `json:"player_id"` // ID of the player sending the message
	Username string  `json:"username"`  // Username of the player
	X        float64 `json:"x"`         // X coordinate
	Y        float64 `json:"y"`         // Y coordinate
	Action   string  `json:"action"`    // Player action (e.g., "move", "attack")
}

// HandleWebSocketConnections manages new WebSocket connections
func HandleWebSocketConnections(w http.ResponseWriter, r *http.Request) {
	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("WebSocket upgrade failed: %v", err)
		return
	}
	defer ws.Close()

	// Get username from the query parameters
	username := r.URL.Query().Get("username")
	if username == "" {
		log.Println("Username is required for WebSocket connection")
		ws.Close()
		return
	}

	playerID := len(clients) + 1 // Simplistic ID generation (can be improved)
	player := &Player{
		ID:       playerID,
		Username: username,
		X:        0,
		Y:        0,
		Action:   "idle",
	}

	// Register the player
	clients[ws] = playerID
	players.Store(playerID, player)
	log.Printf("New player connected: %v (ID: %d)", username, playerID)

	// Send the welcome message to the new player
	welcomeMessage := Message{
		Type:     "welcome",
		PlayerID: playerID,
		Username: username,
		X:        player.X,
		Y:        player.Y,
		Action:   player.Action,
	}
	log.Printf("Welcome message to player %d: ", playerID)
	err = ws.WriteJSON(welcomeMessage)
	if err != nil {
		log.Printf("Error sending welcome message to player %d: %v", playerID, err)
		return
	}

	// Send the current state of all players to the new client (retroactive join)
	players.Range(func(_, value interface{}) bool {
		if existingPlayer, ok := value.(*Player); ok {
			err := ws.WriteJSON(Message{
				Type:     "update", // Update message type for existing players
				PlayerID: existingPlayer.ID,
				Username: existingPlayer.Username,
				X:        existingPlayer.X,
				Y:        existingPlayer.Y,
				Action:   existingPlayer.Action,
			})
			if err != nil {
				log.Printf("Error sending player state: %v", err)
			}
		}
		return true
	})

	// Notify others about the new player
	broadcast <- Message{
		Type:     "join",
		PlayerID: player.ID,
		Username: player.Username,
		X:        player.X,
		Y:        player.Y,
		Action:   player.Action,
	}

	// Handle incoming messages from the player
	handlePlayerMessages(ws, player)
}

// handlePlayerMessages handles incoming messages from a specific player
func handlePlayerMessages(ws *websocket.Conn, player *Player) {
	playerID := player.ID
	for {
		var msg Message
		err := ws.ReadJSON(&msg)
		if err != nil {
			log.Printf("Error reading message from %v: %v", player.Username, err)
			removePlayer(ws, playerID, player.Username)
			break
		}

		// Update the player's state
		if msg.Type == "update" {
			if p, ok := players.Load(playerID); ok {
				updatedPlayer := p.(*Player)
				updatedPlayer.X = msg.X
				updatedPlayer.Y = msg.Y
				updatedPlayer.Action = msg.Action
			}
		}

		// Broadcast the message to other players
		broadcast <- Message{
			Type:     "update",
			PlayerID: playerID,
			Username: player.Username,
			X:        msg.X,
			Y:        msg.Y,
			Action:   msg.Action,
		}
	}
}

// removePlayer handles player disconnection
func removePlayer(ws *websocket.Conn, playerID int, username string) {
	delete(clients, ws)
	players.Delete(playerID)

	// Notify others about the player leaving
	broadcast <- Message{
		Type:     "leave",
		PlayerID: playerID,
		Username: username,
	}
	log.Printf("Player disconnected: %v (ID: %d)", username, playerID)
}

// BroadcastMessages handles broadcasting messages to all connected clients
func BroadcastMessages() {
	for {
		msg := <-broadcast

		log.Printf("Broadcasting update for player %d: %s, position: (%f, %f), action: %s", msg.PlayerID, msg.Username, msg.X, msg.Y, msg.Action)

		// Send the message to all clients
		for client := range clients {
			err := client.WriteJSON(msg)
			if err != nil {
				log.Printf("Error writing message: %v", err)
				client.Close()
				delete(clients, client)
			}
		}
	}
}
