package auth

import (
	"context"
	"time"

	"wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var sessionCol *mongo.Collection

func InitSessionStore(col *mongo.Collection) {
	sessionCol = col
}

func createSession(ctx context.Context, s models.Session) error {
	_, err := sessionCol.InsertOne(ctx, s)
	return err
}

// ✅ Find session by refresh token hash ONLY
func findSessionByRefreshToken(
	ctx context.Context,
	rawToken string,
) (*models.Session, error) {

	cursor, err := sessionCol.Find(ctx, bson.M{
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

		if bcrypt.CompareHashAndPassword(
			[]byte(s.TokenHash),
			[]byte(rawToken),
		) == nil {
			return &s, nil
		}
	}

	return nil, ErrInvalidRefreshToken
}
