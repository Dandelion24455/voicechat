package model

import "time"

type User struct {
	ID           string    `json:"id"`
	PlayerID     string    `json:"player_id"`
	Username     string    `json:"username"`
	PasswordHash string    `json:"-"`
	CreatedAt    time.Time `json:"created_at"`
}
