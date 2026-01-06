package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Session represents a refresh-token backed login session
// TTL index must exist on `expires_at`
type Session struct {
	ID primitive.ObjectID `bson:"_id,omitempty" json:"id"`

	// owner of the session
	UserID primitive.ObjectID `bson:"user_id" json:"user_id"`

	// bcrypt hash of refresh token (never store raw token)
	TokenHash string `bson:"token_hash" json:"-"`

	// request metadata
	UserAgent string `bson:"user_agent,omitempty" json:"user_agent,omitempty"`
	IPAddress string `bson:"ip_address,omitempty" json:"ip_address,omitempty"`

	// lifecycle
	CreatedAt time.Time `bson:"created_at" json:"created_at"`
	ExpiresAt time.Time `bson:"expires_at" json:"expires_at"`

	// refresh-token rotation tracking
	RotatedFrom primitive.ObjectID `bson:"rotated_from,omitempty" json:"rotated_from,omitempty"`
}
