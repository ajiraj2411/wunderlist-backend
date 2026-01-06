package models

import "time"

// MessageResponse is a generic success response
type MessageResponse struct {
	Message string `json:"message"`
}

// ErrorResponse is a generic error response
type ErrorResponse struct {
	Error string `json:"error"`
}

// AuthResponse represents JWT tokens returned on login
type AuthResponse struct {
	AccessToken  string    `json:"access_token"`
	RefreshToken string    `json:"refresh_token"`
	ExpiresAt    time.Time `json:"expires_at"` // access token expiry
}
