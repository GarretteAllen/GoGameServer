package handlers

import (
	"encoding/json"
	"game_server/db"
	"log"
	"net/http"
)

type AddSkillRequest struct {
	PlayerID   int    `json:"player_id"`
	SkillName  string `json:"skill_name"`
	SkillLevel int    `json:"skill_level"`
}

func AddSkill(w http.ResponseWriter, r *http.Request) {
	var req AddSkillRequest
	err := json.NewDecoder(r.Body).Decode(&req)
	if err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate request data
	if req.PlayerID <= 0 || req.SkillName == "" {
		http.Error(w, "PlayerID and SkillName are required and must be valid", http.StatusBadRequest)
		return
	}

	// Default skill level to 1 if not provided
	if req.SkillLevel <= 0 {
		req.SkillLevel = 1
	}

	// Insert skill into database
	_, err = db.GetDB().Exec(`
		INSERT INTO skills (player_id, skill_name, skill_level)
		VALUES (?, ?, ?)`,
		req.PlayerID, req.SkillName, req.SkillLevel)
	if err != nil {
		log.Printf("Error adding skill: %v", err)
		http.Error(w, "Failed to add skill", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Write([]byte("Skill added successfully"))
}
