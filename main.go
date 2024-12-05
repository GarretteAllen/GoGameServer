package main

import (
	"game_server/db"
	"game_server/handlers"
	"log"
	"net/http"
)

func main() {

	db.Init()

	http.HandleFunc("/player/create", handlers.CreatePlayer)
	http.HandleFunc("/player/update", handlers.UpdatePlayer)
	http.HandleFunc("/player/read", handlers.ReadPlayer)
	http.HandleFunc("/ws", handlers.HandleWebSocketConnections)
	http.HandleFunc("/player/add-skill", handlers.AddSkill)

	//Start WebSocket Broadcaster in a goroutine

	go handlers.BroadcastMessages()

	// Start the Server
	port := ":31333"
	log.Println("Server is running on port", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal("Server Failed", err)
	}
}
