package models

type Player struct {
	ID          int            `json:"id"`           // Unique identifier
	Username    string         `json:"username"`     // Player name
	CombatLevel int            `json:"combat_level"` // Calculated combat level
	Skills      map[string]int `json:"skills"`       // Skills and their levels
	CreatedAt   string         `json:"created_at"`   // Creation timestamp
	UpdatedAt   string         `json:"updated_at"`   // Last update timestamp
}
