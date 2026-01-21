package auth

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func FindOrCreateGoogleUser(
	ctx context.Context,
	email string,
	name string,
	googleSub string,
	emailVerified bool,
) (*models.User, error) {

	if userCol == nil {
		return nil, errors.New("user store not initialized")
	}

	email = strings.ToLower(strings.TrimSpace(email))
	name = strings.TrimSpace(name)

	var user models.User
	err := userCol.FindOne(ctx, bson.M{"email": email}).Decode(&user)

	// ✅ existing user
	if err == nil {

		// If it's password account -> block takeover
		if user.AuthProvider == "password" {
			return nil, errors.New("account already exists with password login")
		}

		// If google account, enforce sub binding
		if user.AuthProvider == "google" {
			if user.GoogleSub != "" && user.GoogleSub != googleSub {
				return nil, errors.New("google identity mismatch")
			}

			// Upgrade missing sub (older accounts)
			if user.GoogleSub == "" && googleSub != "" {
				_, _ = userCol.UpdateOne(ctx,
					bson.M{"_id": user.ID},
					bson.M{"$set": bson.M{
						"google_sub":     googleSub,
						"email_verified": emailVerified,
						"name":           name,
					}},
				)

				// reload
				_ = userCol.FindOne(ctx, bson.M{"_id": user.ID}).Decode(&user)
			}
		}

		return &user, nil
	}

	// ✅ create new google user
	u := models.User{
		ID:            primitive.NewObjectID(),
		Email:         email,
		Password:      "",
		Role:          "user",
		AuthProvider:  "google",
		GoogleSub:     googleSub,
		EmailVerified: emailVerified,
		Name:          name,
		CreatedAt:     time.Now().UTC(),
	}

	if _, err := userCol.InsertOne(ctx, u); err != nil {
		return nil, err
	}

	return &u, nil
}
