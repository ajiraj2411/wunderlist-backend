package auth

import (
	"crypto/rand"
	"encoding/base64"
)

// GenerateAccessToken creates a short-lived JWT for API access
func GenerateAccessToken(userID, role string) (string, error) {
	return GenerateJWT(userID, role, accessTTL, "")
}

// GenerateAccessTokenForSession creates access token where jti == sessionID.
// This is required for /api/sessions/current (O(1)).
func GenerateAccessTokenForSession(userID, role, sessionID string) (string, error) {
	return GenerateJWT(userID, role, accessTTL, sessionID)
}

// GenerateRefreshToken creates a cryptographically secure opaque token
// This token is NEVER parsed, only bcrypt-hashed and stored
func GenerateRefreshToken() (string, error) {
	b := make([]byte, 32) // 256-bit
	if _, err := rand.Read(b); err != nil {
		return "", err
	}

	// URL-safe, no padding
	return base64.RawURLEncoding.EncodeToString(b), nil
}
