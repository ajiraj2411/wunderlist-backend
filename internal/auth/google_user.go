package auth

import (
	"context"
	"errors"
	"time"

	"wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func FindOrCreateGoogleUser(
	ctx context.Context,
	email string,
	name string,
) (*models.User, error) {

	if userCol == nil {
		return nil, errors.New("user store not initialized")
	}

	var user models.User

	err := userCol.FindOne(ctx, bson.M{"email": email}).Decode(&user)
	if err == nil {
		return &user, nil
	}

	// create new user
	user = models.User{
		ID:        primitive.NewObjectID(),
		Email:     email,
		Password:  "", // passwordless (Google)
		Role:      "user",
		CreatedAt: time.Now().UTC(),
	}

	if _, err := userCol.InsertOne(ctx, user); err != nil {
		return nil, err
	}

	return &user, nil
}
