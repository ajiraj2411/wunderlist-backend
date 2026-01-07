package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type SessionResponse struct {
	ID        primitive.ObjectID `json:"id"`
	UserAgent string             `json:"user_agent"`
	IPAddress string             `json:"ip_address"`
	CreatedAt time.Time          `json:"created_at"`
	ExpiresAt time.Time          `json:"expires_at"`
	IsCurrent bool               `json:"is_current"`
}
