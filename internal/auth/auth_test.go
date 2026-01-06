package auth_test

import (
	"context"
	"testing"
	"time"

	"wunderlist-backend/internal/auth"
	"wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestLoginSuccess(t *testing.T) {

	auth.InitJWT("test-secret", 15*time.Minute, 7*24*time.Hour)

	user := &models.User{
		ID:       primitive.NewObjectID(),
		Email:    "test@example.com",
		Password: mustHash(t, "password123"),
	}

	ctx := context.Background()

	access, refresh, err := auth.Login(
		ctx,
		user,
		"password123",
		"test-agent",
		"127.0.0.1",
	)

	if err != nil {
		t.Fatalf("login failed: %v", err)
	}

	if access == "" || refresh == "" {
		t.Fatalf("tokens should not be empty")
	}
}

func TestLoginWrongPassword(t *testing.T) {

	user := &models.User{
		ID:       primitive.NewObjectID(),
		Password: mustHash(t, "correct-password"),
	}

	_, _, err := auth.Login(
		context.Background(),
		user,
		"wrong-password",
		"agent",
		"ip",
	)

	if err == nil {
		t.Fatalf("expected login failure")
	}
}

func mustHash(t *testing.T, password string) string {
	t.Helper()
	hash, err := auth.HashPassword(password)
	if err != nil {
		t.Fatal(err)
	}
	return hash
}
