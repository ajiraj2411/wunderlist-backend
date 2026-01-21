package auth

import (
	"context"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// ========================
// VALIDATE REFRESH (STRICT)
// ========================
//
// Rules:
// 1. bcrypt match
// 2. not expired
// 3. UA + IP match
// 4. hijack → revoke ALL sessions
func ValidateRefreshTokenStrict(
	ctx context.Context,
	rawToken string,
	userAgent string,
	ip string,
) (*models.Session, error) {

	sha := refreshTokenSHA(rawToken)

	var s models.Session
	err := sessionCol.FindOne(ctx, bson.M{
		"token_sha":  sha,
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	}).Decode(&s)

	if err != nil {
		return nil, ErrInvalidRefreshToken
	}

	// 🔑 bcrypt compare (protect against SHA collision / DB tampering)
	if bcrypt.CompareHashAndPassword([]byte(s.TokenHash), []byte(rawToken)) != nil {
		return nil, ErrInvalidRefreshToken
	}

	// 🔐 UA binding
	if s.UserAgent != "" && s.UserAgent != userAgent {
		revokeUserSessions(ctx, s.UserID)
		return nil, ErrSessionHijacked
	}

	// 🔐 IP binding
	if s.IPAddress != "" && s.IPAddress != ip {
		revokeUserSessions(ctx, s.UserID)
		return nil, ErrSessionHijacked
	}

	return &s, nil
}

// ========================
// ROTATE REFRESH TOKEN
// ========================
//
// - delete old session (single-use)
// - issue new opaque refresh
func RotateRefreshToken(
	ctx context.Context,
	session *models.Session,
) (string, string, error) {

	// new opaque refresh token
	refresh, err := GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	hash, err := HashPassword(refresh)
	if err != nil {
		return "", "", err
	}

	sha := refreshTokenSHA(refresh)

	// 🔥 delete old refresh session (single-use)
	_, _ = sessionCol.DeleteOne(ctx, bson.M{"_id": session.ID})

	// ✅ create new refresh session (new ID)
	newSession := models.Session{
		ID:          primitive.NewObjectID(),
		UserID:      session.UserID,
		TokenHash:   hash,
		TokenSHA:    sha,
		Role:        session.Role,
		UserAgent:   session.UserAgent,
		IPAddress:   session.IPAddress,
		RotatedFrom: session.ID,
		CreatedAt:   time.Now().UTC(),
		ExpiresAt:   time.Now().UTC().Add(refreshTTL),
	}

	if _, err := sessionCol.InsertOne(ctx, newSession); err != nil {
		return "", "", err
	}

	// ✅ access JWT must bind jti == newSession.ID
	access, err := GenerateAccessTokenForSession(
		newSession.UserID.Hex(),
		newSession.Role,
		newSession.ID.Hex(),
	)
	if err != nil {
		return "", "", err
	}

	return access, refresh, nil
}

// ========================
// REVOKE ALL USER SESSIONS
// ========================

func revokeUserSessions(ctx context.Context, userID primitive.ObjectID) {
	_, _ = sessionCol.DeleteMany(ctx, bson.M{"user_id": userID})
}
