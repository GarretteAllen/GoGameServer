package handlers

import (
	"encoding/json"
	"game_server/db"
	"game_server/models"
	"log"
	"net/http"
	"time"
)

type CreatePlayerRequest struct {
	Username    string `json:"username"`     // Player's username
	CombatLevel int    `json:"combat_level"` // Player's combat level
}

func CreatePlayer(w http.ResponseWriter, r *http.Request) {
	var req CreatePlayerRequest

	// Decode the request body into CreatePlayerRequest struct
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		log.Printf("Error decoding request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate that username is provided
	if req.Username == "" {
		http.Error(w, "Username is required", http.StatusBadRequest)
		return
	}

	// Default combat level to 1 if not provided
	if req.CombatLevel <= 0 {
		req.CombatLevel = 1
	}

	db := db.GetDB()
	tx, err := db.Begin()
	if err != nil {
		log.Printf("Error starting transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Insert player into the players table
	result, err := tx.Exec("INSERT INTO players (username, combat_level) VALUES (?, ?)", req.Username, req.CombatLevel)
	if err != nil {
		log.Printf("Error creating player: %v", err)
		tx.Rollback()
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	playerID, err := result.LastInsertId()
	if err != nil {
		log.Printf("Error fetching player ID: %v", err)
		tx.Rollback()
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Default skills
	defaultSkills := []string{"woodcutting", "firemaking", "attack", "strength", "defense"}
	for _, skill := range defaultSkills {
		_, err := tx.Exec("INSERT INTO skills (player_id, skill_name, skill_level) VALUES (?, ?, ?)", playerID, skill, 1)
		if err != nil {
			log.Printf("Error adding skill: %v", err)
			tx.Rollback()
			http.Error(w, "Database error", http.StatusInternalServerError)
			return
		}
	}

	// Commit the transaction
	err = tx.Commit()
	if err != nil {
		log.Printf("Error committing transaction: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	// Return response with player info
	response := map[string]interface{}{
		"id":           playerID,
		"username":     req.Username,
		"combat_level": req.CombatLevel,
		"skills":       defaultSkills,
		"created_at":   time.Now().Format(time.RFC3339),
		"updated_at":   time.Now().Format(time.RFC3339),
	}

	// Send a JSON response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

func UpdatePlayer(w http.ResponseWriter, r *http.Request) {
	playerID := r.URL.Query().Get("player_id")
	skillName := r.URL.Query().Get("skill_name")
	skillLevel := r.URL.Query().Get("skill_level")

	if playerID == "" || skillName == "" || skillLevel == "" {
		http.Error(w, "Player ID, skill name, and skill level are required", http.StatusBadRequest)
		return
	}

	db := db.GetDB()
	_, err := db.Exec(
		"UPDATE skills SET skill_level = ? WHERE player_id = ? AND skill_name = ?",
		skillLevel, playerID, skillName,
	)
	if err != nil {
		log.Printf("Error updating skill: %v", err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}

	w.Write([]byte("Skill updated successfully"))
}

func ReadPlayer(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing player ID", http.StatusBadRequest)
		return
	}

	var player models.Player
	err := db.GetDB().QueryRow(`
		SELECT id, username, combat_level, created_at, updated_at
		FROM players
		WHERE id=?`, id).Scan(
		&player.ID,
		&player.Username,
		&player.CombatLevel,
		&player.CreatedAt,
		&player.UpdatedAt,
	)
	if err != nil {
		http.Error(w, "Player not found", http.StatusNotFound)
		return
	}

	// Fetch skills
	rows, err := db.GetDB().Query(`
		SELECT skill_name, skill_level
		FROM skills
		WHERE player_id=?`, player.ID)
	if err != nil {
		http.Error(w, "Failed to fetch skills", http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	player.Skills = make(map[string]int)
	for rows.Next() {
		var skillName string
		var skillLevel int
		err := rows.Scan(&skillName, &skillLevel)
		if err != nil {
			log.Printf("Error scanning skill: %v", err)
			continue
		}
		player.Skills[skillName] = skillLevel
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(player)
}
