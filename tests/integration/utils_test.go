package integration

import (
	"context"
	"encoding/json"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// extractJSONField is a tiny helper to avoid repeating json structs
func extractJSONField(jsonStr string, key string) string {
	var m map[string]string
	_ = json.Unmarshal([]byte(jsonStr), &m)
	return m[key]
}

func getUserIDByEmail(t *testing.T, email string) string {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	err := TestDB.Collection("users").
		FindOne(ctx, bson.M{"email": strings.ToLower(email)}).
		Decode(&result)

	if err != nil {
		t.Fatalf("failed to get user by email %s: %v", email, err)
	}

	return result.ID.Hex()
}
