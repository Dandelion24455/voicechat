package model

import "time"

type Room struct {
	ID         string    `json:"id"`
	Name       string    `json:"name"`
	CreatorID   string    `json:"creator_id"`
	InviteCode string    `json:"invite_code"`
	CreatedAt  time.Time `json:"created_at"`
}
