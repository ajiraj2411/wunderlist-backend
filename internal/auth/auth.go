package auth

import (
	"context"
	"time"

	"wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// ========================
// LOGIN
// ========================

func Login(
	ctx context.Context,
	user *models.User,
	password string,
	userAgent string,
	ip string,
) (string, string, error) {

	// ✅ password check (bcrypt hash vs raw password)
	if ComparePassword(user.Password, password) != nil {
		return "", "", ErrInvalidCredentials
	}

	// ✅ access JWT
	access, err := GenerateAccessToken(user.ID.Hex())
	if err != nil {
		return "", "", err
	}

	// ✅ opaque refresh token
	refresh, err := GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	// ✅ bcrypt hash refresh token
	hash, err := HashPassword(refresh)
	if err != nil {
		return "", "", err
	}

	// ✅ session record
	session := models.Session{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		TokenHash: hash, // 🔴 NEVER EMPTY
		UserAgent: userAgent,
		IPAddress: ip,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(refreshTTL),
	}

	if err := createSession(ctx, session); err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

// ========================
// LOGOUT
// ========================

func DeleteSessionByToken(ctx context.Context, raw string) error {
	cursor, err := sessionCol.Find(ctx, bson.M{})
	if err != nil {
		return err
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var s models.Session
		if err := cursor.Decode(&s); err != nil {
			continue
		}

		if bcrypt.CompareHashAndPassword(
			[]byte(s.TokenHash),
			[]byte(raw),
		) == nil {
			_, err := sessionCol.DeleteOne(ctx, bson.M{"_id": s.ID})
			return err
		}
	}

	return nil
}

func DeleteAllSessions(ctx context.Context, userID primitive.ObjectID) error {
	_, err := sessionCol.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}
