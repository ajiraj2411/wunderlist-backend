package auth

import (
	"context"
	"testing"
	"time"

	"wunderlist-backend/db"
	"wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func setupLoginTest(t *testing.T) context.Context {
	t.Helper()

	client := db.ConnectMongo()
	col := client.Database("wunderlist_test").Collection("sessions_login_test")

	// isolate from other tests
	_ = col.Drop(context.Background())

	InitSessionStore(col)
	InitJWT("test-secret", 15*time.Minute, 7*24*time.Hour)

	return context.Background()
}

func TestLogin_PasswordUser_Success(t *testing.T) {
	ctx := setupLoginTest(t)

	user := &models.User{
		ID:           primitive.NewObjectID(),
		Email:        "pw@test.com",
		Password:     mustHash2(t, "password123"),
		Role:         "user",
		AuthProvider: "password",
		CreatedAt:    time.Now().UTC(),
	}

	access, refresh, err := Login(ctx, user, "password123", "ua", "127.0.0.1")
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("expected tokens, got empty")
	}

	// ensure session inserted
	count, _ := sessionCol.CountDocuments(ctx, bson.M{"user_id": user.ID})
	if count != 1 {
		t.Fatalf("expected 1 session, got %d", count)
	}
}

func TestLogin_PasswordUser_WrongPassword(t *testing.T) {
	ctx := setupLoginTest(t)

	user := &models.User{
		ID:           primitive.NewObjectID(),
		Email:        "pw@test.com",
		Password:     mustHash2(t, "password123"),
		Role:         "user",
		AuthProvider: "password",
		CreatedAt:    time.Now().UTC(),
	}

	_, _, err := Login(ctx, user, "wrong", "ua", "127.0.0.1")
	if err == nil {
		t.Fatalf("expected error for wrong password")
	}
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func TestLogin_GoogleUser_PasswordlessAllowed(t *testing.T) {
	ctx := setupLoginTest(t)

	user := &models.User{
		ID:           primitive.NewObjectID(),
		Email:        "google@test.com",
		Password:     "", // ✅ passwordless
		Role:         "user",
		AuthProvider: "google",
		CreatedAt:    time.Now().UTC(),
	}

	access, refresh, err := Login(ctx, user, "", "ua", "127.0.0.1")
	if err != nil {
		t.Fatalf("expected success, got err: %v", err)
	}
	if access == "" || refresh == "" {
		t.Fatal("expected tokens, got empty")
	}

	// ensure session inserted
	count, _ := sessionCol.CountDocuments(ctx, bson.M{"user_id": user.ID})
	if count != 1 {
		t.Fatalf("expected 1 session, got %d", count)
	}
}

func TestLogin_GoogleUser_RejectsNonEmptyPassword(t *testing.T) {
	ctx := setupLoginTest(t)

	user := &models.User{
		ID:           primitive.NewObjectID(),
		Email:        "google@test.com",
		Password:     "", // passwordless
		Role:         "user",
		AuthProvider: "google",
		CreatedAt:    time.Now().UTC(),
	}

	_, _, err := Login(ctx, user, "some-password", "ua", "127.0.0.1")
	if err == nil {
		t.Fatalf("expected error for non-empty password on passwordless account")
	}
	if err != ErrInvalidCredentials {
		t.Fatalf("expected ErrInvalidCredentials, got: %v", err)
	}
}

func mustHash2(t *testing.T, raw string) string {
	t.Helper()
	h, err := HashPassword(raw)
	if err != nil {
		t.Fatal(err)
	}
	return h
}
