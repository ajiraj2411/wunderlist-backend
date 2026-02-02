package auth

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

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

func SessionExists(
	ctx context.Context,
	sessionID primitive.ObjectID,
	userID primitive.ObjectID,
) (bool, error) {

	count, err := sessionCol.CountDocuments(ctx, bson.M{
		"_id":        sessionID,
		"user_id":    userID,
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	})

	if err != nil {
		return false, err
	}

	return count == 1, nil
}
