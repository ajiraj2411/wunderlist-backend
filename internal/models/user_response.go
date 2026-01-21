package models

import "time"

type MeResponse struct {
	ID                  string    `json:"id"`
	Email               string    `json:"email"`
	Role                string    `json:"role"`
	CreatedAt           time.Time `json:"created_at"`
	ActiveSessionsCount int64     `json:"active_sessions_count"`
}
