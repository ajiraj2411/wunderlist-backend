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

func ValidateAccessToken(tokenString string) (string, string, string, time.Time, error) {
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
		return "", "", "", time.Time{}, ErrInvalidToken
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || claims.Subject == "" || claims.IssuedAt == nil {
		return "", "", "", time.Time{}, ErrInvalidToken
	}

	return claims.Subject, claims.Role, claims.ID, claims.IssuedAt.Time, nil
}

// AccessTokenExpiry returns the expiry time for a newly issued access token
func AccessTokenExpiry() time.Time {
	return time.Now().UTC().Add(accessTTL)
}

// GenerateJWTWithTimes is used for tests to build tokens with controlled timestamps.
// Production code should continue using GenerateJWT / GenerateAccessToken.
func GenerateJWTWithTimes(userID string, role string, issuedAt time.Time, expiresAt time.Time) (string, error) {
	if jwtSecret == nil {
		return "", errors.New("jwt not initialized")
	}

	claims := AccessClaims{
		RegisteredClaims: jwt.RegisteredClaims{
			Subject:   userID,
			Issuer:    jwtIssuer,
			Audience:  []string{jwtAudience},
			IssuedAt:  jwt.NewNumericDate(issuedAt.UTC()),
			ExpiresAt: jwt.NewNumericDate(expiresAt.UTC()),
			ID:        uuid.NewString(),
		},
		Role: role,
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(jwtSecret)
}
