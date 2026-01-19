package auth

import (
	"context"
	"testing"
	"time"

	"wunderlist-backend/config"
	"wunderlist-backend/db"
	"wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

/*
	TEST SETUP
*/

func setupSessionTestDB(t *testing.T) context.Context {
	t.Helper()

	client := db.ConnectMongo()
	col := client.Database("wunderlist_test").Collection("sessions")

	InitSessionStore(col)
	InitJWT(
		config.AppConfig.JWTSecret,
		15*time.Minute,
		24*time.Hour,
	)

	if err := col.Drop(context.Background()); err != nil {
		t.Fatalf("failed to drop test collection: %v", err)
	}

	return context.Background()
}

/*
	TEST: Strict refresh – invalid token
*/

func TestValidateRefreshTokenStrict_Invalid(t *testing.T) {
	ctx := setupSessionTestDB(t)

	_, err := ValidateRefreshTokenStrict(
		ctx,
		"invalid-token",
		"UA",
		"127.0.0.1",
	)

	if err == nil {
		t.Fatal("expected error for invalid refresh token")
	}
}

/*
	TEST: Strict refresh – valid token
*/

func TestValidateRefreshTokenStrict_Valid(t *testing.T) {
	ctx := setupSessionTestDB(t)

	userID := primitive.NewObjectID()
	rawToken := "refresh-token-123"

	hash, _ := HashPassword(rawToken)

	session := models.Session{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		TokenHash: hash,
		TokenSHA:  refreshTokenSHA(rawToken),
		UserAgent: "UA",
		IPAddress: "127.0.0.1",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	_, err := sessionCol.InsertOne(ctx, session)
	if err != nil {
		t.Fatalf("failed to insert session: %v", err)
	}

	s, err := ValidateRefreshTokenStrict(
		ctx,
		rawToken,
		"UA",
		"127.0.0.1",
	)
	if err != nil {
		t.Fatalf("expected valid token, got error: %v", err)
	}

	if s.UserID != userID {
		t.Fatal("user mismatch")
	}
}

/*
	TEST: Strict refresh – hijack revokes all sessions
*/

func TestValidateRefreshTokenStrict_HijackRevokesAll(t *testing.T) {
	ctx := setupSessionTestDB(t)

	userID := primitive.NewObjectID()

	hash, _ := HashPassword("valid-token")

	session := models.Session{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		TokenHash: hash,
		TokenSHA:  refreshTokenSHA("valid-token"),
		UserAgent: "UA",
		IPAddress: "127.0.0.1",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	_, _ = sessionCol.InsertOne(ctx, session)

	_, err := ValidateRefreshTokenStrict(
		ctx,
		"valid-token",
		"OTHER-UA",
		"127.0.0.1",
	)

	if err == nil {
		t.Fatal("expected hijack error")
	}

	count, _ := sessionCol.CountDocuments(ctx, bson.M{"user_id": userID})
	if count != 0 {
		t.Fatalf("expected sessions revoked, found %d", count)
	}
}

/*
	TEST: Rotate refresh token
*/

func TestRotateRefreshToken(t *testing.T) {
	ctx := setupSessionTestDB(t)

	userID := primitive.NewObjectID()
	rawToken := "old-refresh"

	hash, _ := HashPassword(rawToken)

	session := models.Session{
		ID:        primitive.NewObjectID(),
		UserID:    userID,
		TokenHash: hash,
		TokenSHA:  refreshTokenSHA(rawToken),
		UserAgent: "UA",
		IPAddress: "127.0.0.1",
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().Add(time.Hour),
	}

	_, _ = sessionCol.InsertOne(ctx, session)

	access, refresh, err := RotateRefreshToken(ctx, &session)
	if err != nil {
		t.Fatalf("rotation failed: %v", err)
	}

	if access == "" || refresh == "" {
		t.Fatal("tokens should not be empty")
	}

	count, _ := sessionCol.CountDocuments(ctx, bson.M{"user_id": userID})
	if count != 1 {
		t.Fatalf("expected 1 active session, got %d", count)
	}
}
