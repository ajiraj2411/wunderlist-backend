package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

var (
	jwtSecret   []byte
	accessTTL   time.Duration
	refreshTTL  time.Duration
	jwtIssuer   = "wunderlist-backend"
	jwtAudience = "wunderlist-client"
)

type AccessClaims struct {
	jwt.RegisteredClaims
	Role string `json:"role"`
}

func InitJWT(secret string, access, refresh time.Duration) {
	jwtSecret = []byte(secret)
	accessTTL = access
	refreshTTL = refresh
}

func GenerateJWT(userID string, role string, ttl time.Duration) (string, error) {
	if jwtSecret == nil {
		return "", errors.New("jwt not initialized")
	}

	now := time.Now().UTC()

	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    jwtIssuer,
			Audience:  []string{jwtAudience},
			IssuedAt:  jwt.NewNumericDate(now),
			ExpiresAt: jwt.NewNumericDate(now.Add(ttl)),
			ID:        uuid.NewString(),
		},
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}

func ValidateAccessToken(tokenString string) (string, string, string, error) {
	token, err := jwt.ParseWithClaims(
		tokenString,
		&AccessClaims{},
		func(t *jwt.Token) (interface{}, error) {
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, ErrInvalidToken
			}
			return jwtSecret, nil
		},
		jwt.WithAudience(jwtAudience),
		jwt.WithIssuer(jwtIssuer),
	)

	if err != nil || !token.Valid {
		return "", "", "", ErrInvalidToken
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || claims.Subject == "" {
		return "", "", "", ErrInvalidToken
	}

	return claims.Subject, claims.Role, claims.ID, nil
}

// AccessTokenExpiry returns the expiry time for a newly issued access token
func AccessTokenExpiry() time.Time {
	return time.Now().UTC().Add(accessTTL)
}
