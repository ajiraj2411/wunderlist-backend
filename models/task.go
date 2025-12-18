package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// Task represents a single task in a list
type Task struct {
	ID          primitive.ObjectID `bson:"_id,omitempty" json:"id,omitempty"`
	Title       string             `bson:"title" json:"title"`
	Description string             `bson:"description,omitempty" json:"description,omitempty"`
	Completed   bool               `bson:"completed" json:"completed"`
	Priority    string             `bson:"priority,omitempty" json:"priority,omitempty"` // e.g., low, medium, high
	DueDate     time.Time          `bson:"due_date,omitempty" json:"due_date,omitempty"`
	ListID      primitive.ObjectID `bson:"list_id" json:"list_id"`
	UserID      primitive.ObjectID `bson:"user_id" json:"user_id"`
	CreatedAt   time.Time          `bson:"created_at,omitempty" json:"created_at,omitempty"`
	UpdatedAt   time.Time          `bson:"updated_at,omitempty" json:"updated_at,omitempty"`
}
