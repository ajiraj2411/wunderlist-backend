package integration

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func extractJSONField(jsonStr string, key string) string {
	var m map[string]interface{}
	_ = json.Unmarshal([]byte(jsonStr), &m)

	v, ok := m[key]
	if !ok || v == nil {
		return ""
	}

	// return string as-is
	if s, ok := v.(string); ok {
		return s
	}

	// convert numbers to string
	return strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(
		fmt.Sprintf("%v", v), "\n", "",
	), "\t", ""))
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
