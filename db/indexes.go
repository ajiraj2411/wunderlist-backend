package db

import (
	"context"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// EnsureIndexes creates all required MongoDB indexes
// Safe to call multiple times (idempotent)
func EnsureIndexes(db *mongo.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// =========================
	// USERS
	// =========================
	_, err := db.Collection("users").Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys: bson.D{{Key: "email", Value: 1}},
			Options: options.Index().
				SetUnique(true).
				SetName("email_unique_idx"),
		},
	)
	logIndexResult("users.email", err)

	// =========================
	// SESSIONS (refresh tokens)
	// =========================
	_, err = db.Collection("sessions").Indexes().CreateOne(ctx,
		mongo.IndexModel{
			Keys: bson.D{{Key: "expires_at", Value: 1}},
			Options: options.Index().
				SetExpireAfterSeconds(0). // TTL index
				SetName("session_ttl_idx"),
		},
	)
	logIndexResult("sessions.expires_at (TTL)", err)

	// =========================
	// LISTS
	// =========================
	lists := db.Collection("lists")

	listIndexes := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "user_id", Value: 1}},
			Options: options.Index().SetName("lists_user_idx"),
		},
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "title", Value: 1},
			},
			Options: options.Index().
				SetUnique(true).
				SetName("lists_user_title_unique_idx"),
		},
	}

	_, err = lists.Indexes().CreateMany(ctx, listIndexes)
	logIndexResult("lists indexes", err)

	// =========================
	// TASKS
	// =========================
	tasks := db.Collection("tasks")

	taskIndexes := []mongo.IndexModel{
		// user → list lookup
		{
			Keys: bson.D{
				{Key: "user_id", Value: 1},
				{Key: "list_id", Value: 1},
			},
			Options: options.Index().SetName("tasks_user_list_idx"),
		},

		// text search (SearchTasks)
		{
			Keys: bson.D{
				{Key: "title", Value: "text"},
				{Key: "description", Value: "text"},
			},
			Options: options.Index().
				SetName("tasks_text_search_idx").
				SetDefaultLanguage("english"),
		},

		// active tasks filter
		{
			Keys: bson.D{{Key: "completed", Value: 1}},
			Options: options.Index().
				SetPartialFilterExpression(bson.M{"completed": false}).
				SetName("tasks_active_idx"),
		},

		// sorting by due date
		{
			Keys:    bson.D{{Key: "due_date", Value: 1}},
			Options: options.Index().SetName("tasks_due_date_idx"),
		},
	}

	_, err = tasks.Indexes().CreateMany(ctx, taskIndexes)
	logIndexResult("tasks indexes", err)
}

// Centralized index logging
func logIndexResult(name string, err error) {
	if err != nil {
		log.Printf("❌ Failed to ensure index (%s): %v\n", name, err)
		return
	}
	log.Printf("✅ Index ensured: %s\n", name)
}
