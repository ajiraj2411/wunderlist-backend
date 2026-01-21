package auth

import (
	"context"
	"strings"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/models"

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

	// ✅ Password account: must validate
	if user.Password != "" {
		if ComparePassword(user.Password, password) != nil {
			return "", "", ErrInvalidCredentials
		}
	} else {
		// ✅ Provider account (Google): must be passwordless login
		if password != "" {
			return "", "", ErrInvalidCredentials
		}
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

	// ✅ session record first (so we can bind access token jti == session.ID)
	session := models.Session{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		TokenHash: hash,
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

	// ✅ access JWT jti == sessionID (Mongo session _id hex)
	access, err := GenerateAccessTokenForSession(user.ID.Hex(), user.Role, session.ID.Hex())
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

// ========================
// LOGOUT
// ========================
func DeleteSessionByToken(ctx context.Context, raw string) error {
	sha := refreshTokenSHA(raw)

	var s models.Session
	err := sessionCol.FindOne(ctx, bson.M{
		"token_sha":  sha,
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	}).Decode(&s)

	if err != nil {
		return nil // ✅ idempotent logout
	}

	// verify with bcrypt
	if bcrypt.CompareHashAndPassword([]byte(s.TokenHash), []byte(raw)) != nil {
		return nil // ✅ don't delete others
	}

	_, err = sessionCol.DeleteOne(ctx, bson.M{"_id": s.ID})
	return err
}

func DeleteAllSessions(ctx context.Context, userID primitive.ObjectID) error {
	_, err := sessionCol.DeleteMany(ctx, bson.M{"user_id": userID})
	return err
}
