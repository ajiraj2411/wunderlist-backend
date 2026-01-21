package auth

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var ErrInvalidResetToken = errors.New("invalid reset token")

const resetTokenTTL = 30 * time.Minute

func generateResetToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}

func CreatePasswordResetToken(
	ctx context.Context,
	userID primitive.ObjectID,
) (string, error) {

	raw, err := generateResetToken()
	if err != nil {
		return "", err
	}

	hash, err := HashPassword(raw)
	if err != nil {
		return "", err
	}

	_, err = resetTokenCol.InsertOne(ctx, models.PasswordResetToken{
		UserID:    userID,
		TokenHash: hash,
		ExpiresAt: time.Now().UTC().Add(resetTokenTTL),
		Used:      false,
		CreatedAt: time.Now().UTC(),
	})

	return raw, err
}

func ConsumePasswordResetToken(
	ctx context.Context,
	raw string,
) (*models.PasswordResetToken, error) {

	cursor, err := resetTokenCol.Find(ctx, bson.M{
		"used": false,
		"expires_at": bson.M{
			"$gt": time.Now().UTC(),
		},
	})
	if err != nil {
		return nil, ErrInvalidResetToken
	}
	defer cursor.Close(ctx)

	for cursor.Next(ctx) {
		var t models.PasswordResetToken
		if cursor.Decode(&t) != nil {
			continue
		}

		if ComparePassword(t.TokenHash, raw) == nil {
			// mark used
			_, _ = resetTokenCol.UpdateByID(ctx, t.ID, bson.M{
				"$set": bson.M{"used": true},
			})
			return &t, nil
		}
	}

	return nil, ErrInvalidResetToken
}

// ===== TEST HELPERS =====

// CreatePasswordResetTokenForTests exposes raw token ONLY for tests.
// Do NOT use this in production handlers.
func CreatePasswordResetTokenForTests(
	ctx context.Context,
	userID primitive.ObjectID,
) (string, error) {
	return CreatePasswordResetToken(ctx, userID)
}
