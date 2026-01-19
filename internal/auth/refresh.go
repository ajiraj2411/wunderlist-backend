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
	cursor, err := sessionCol.Find(ctx, bson.M{
		"token_sha":  sha,
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	})
	if err != nil {
		return nil, ErrInvalidRefreshToken
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var s models.Session
		if err := cursor.Decode(&s); err != nil {
			continue
		}

		// 🔑 bcrypt compare
		if bcrypt.CompareHashAndPassword(
			[]byte(s.TokenHash),
			[]byte(rawToken),
		) != nil {
			continue
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

		// ✅ VALID SESSION
		return &s, nil
	}

	return nil, ErrInvalidRefreshToken
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

	// new access token
	access, err := GenerateAccessToken(session.UserID.Hex(), session.Role)
	if err != nil {
		return "", "", err
	}

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

	// 🔥 delete old refresh (single-use)
	_, _ = sessionCol.DeleteOne(ctx, bson.M{"_id": session.ID})

	// insert rotated session
	_, err = sessionCol.InsertOne(ctx, models.Session{
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
	})

	return access, refresh, err
}

// ========================
// REVOKE ALL USER SESSIONS
// ========================

func revokeUserSessions(ctx context.Context, userID primitive.ObjectID) {
	_, _ = sessionCol.DeleteMany(ctx, bson.M{"user_id": userID})
}
