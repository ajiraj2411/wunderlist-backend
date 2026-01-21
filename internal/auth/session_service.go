package auth

import (
	"context"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func ListUserSessions(
	ctx context.Context,
	userID primitive.ObjectID,
	currentSessionID primitive.ObjectID,
) ([]models.SessionResponse, error) {

	cursor, err := sessionCol.Find(ctx, bson.M{
		"user_id":    userID,
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	var sessions []models.SessionResponse

	for cursor.Next(ctx) {
		var s models.Session
		if err := cursor.Decode(&s); err != nil {
			continue
		}

		sessions = append(sessions, models.SessionResponse{
			ID:        s.ID,
			UserAgent: s.UserAgent,
			IPAddress: s.IPAddress,
			CreatedAt: s.CreatedAt,
			ExpiresAt: s.ExpiresAt,
			IsCurrent: s.ID == currentSessionID,
		})
	}

	return sessions, nil
}

func DeleteSessionByID(
	ctx context.Context,
	sessionID primitive.ObjectID,
	userID primitive.ObjectID,
) (int64, error) {

	res, err := sessionCol.DeleteOne(ctx, bson.M{
		"_id":     sessionID,
		"user_id": userID,
	})
	if err != nil {
		return 0, err
	}

	return res.DeletedCount, nil
}

func DeleteSessionByIDOnly(
	ctx context.Context,
	sessionID primitive.ObjectID,
) (int64, error) {
	res, err := sessionCol.DeleteOne(ctx, bson.M{"_id": sessionID})
	if err != nil {
		return 0, err
	}
	return res.DeletedCount, nil
}
