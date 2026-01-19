package auth

import (
	"context"
	"strings"
	"time"

	"wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// ========================
// LOGIN
// ========================
//
// Supports:
// - password login (bcrypt validation)
// - provider login (passwordless accounts like Google)
//
// Rules:
// - if user.Password exists => must validate password
// - if user.Password empty  => allow only if password is also empty (provider login)
func Login(
	ctx context.Context,
	user *models.User,
	password string,
	userAgent string,
	ip string,
) (string, string, error) {

	password = strings.TrimSpace(password)

	// ✅ If password hash exists => enforce password check
	if user.Password != "" {
		if ComparePassword(user.Password, password) != nil {
			return "", "", ErrInvalidCredentials
		}
	} else {
		// ✅ Passwordless user (Google)
		// Do NOT attempt bcrypt compare, it will fail.
		// Only allow if caller isn't trying to supply a password.
		if password != "" {
			return "", "", ErrInvalidCredentials
		}
	}

	// ✅ access JWT
	access, err := GenerateAccessToken(user.ID.Hex(), user.Role)
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

	sha := refreshTokenSHA(refresh)

	// ✅ session record
	session := models.Session{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		TokenHash: hash, // 🔴 NEVER EMPTY
		TokenSHA:  sha,
		Role:      user.Role,
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
	sha := refreshTokenSHA(raw)
	cursor, err := sessionCol.Find(ctx, bson.M{
		"token_sha": sha,
	})
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
