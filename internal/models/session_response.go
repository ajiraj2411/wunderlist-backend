package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SessionResponse struct {
	ID        string    `json:"id"`
	UserAgent string    `json:"user_agent"`
	IPAddress string    `json:"ip_address"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at"`
	IsCurrent bool      `json:"is_current"`
}

func NewSessionResponse(
	s Session,
	currentSessionID primitive.ObjectID,
) SessionResponse {

	return SessionResponse{
		ID:        s.ID.Hex(),
		CreatedAt: s.CreatedAt,
		ExpiresAt: s.ExpiresAt,

		UserAgent: s.UserAgent,
		IPAddress: s.IPAddress,

		IsCurrent: s.ID == currentSessionID,
	}
}
