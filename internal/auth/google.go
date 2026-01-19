package auth

import (
	"context"
	"errors"
	"os"

	"google.golang.org/api/idtoken"
)

var googleVerifier = VerifyGoogleIDToken

type GoogleUser struct {
	GoogleID string
	Email    string
	Name     string
}

func VerifyGoogleIDToken(ctx context.Context, idToken string) (*GoogleUser, error) {

	clientID := os.Getenv("GOOGLE_CLIENT_ID")
	if clientID == "" {
		return nil, errors.New("google client id not configured")
	}

	payload, err := idtoken.Validate(ctx, idToken, clientID)
	if err != nil {
		return nil, err
	}

	email, ok := payload.Claims["email"].(string)
	if !ok || email == "" {
		return nil, errors.New("email not found in google token")
	}

	name, _ := payload.Claims["name"].(string)

	return &GoogleUser{
		GoogleID: payload.Subject,
		Email:    email,
		Name:     name,
	}, nil
}
