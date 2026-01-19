package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// User represents a registered user
type User struct {
	ID       primitive.ObjectID `bson:"_id,omitempty" json:"id"`
	Email    string             `bson:"email" json:"email"`
	Password string             `bson:"password" json:"-"`
	Role     string             `bson:"role" json:"role"`
	// ✅ Auth identity
	AuthProvider  string `bson:"auth_provider" json:"auth_provider"` // "password" or "google"
	GoogleSub     string `bson:"google_sub,omitempty" json:"-"`
	EmailVerified bool   `bson:"email_verified" json:"email_verified"`
	Name          string `bson:"name,omitempty" json:"name,omitempty"`

	CreatedAt time.Time `bson:"created_at" json:"created_at"`
}
