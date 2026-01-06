package auth

import "errors"

var (
	ErrInvalidCredentials  = errors.New("invalid credentials")
	ErrInvalidToken        = errors.New("invalid token")
	ErrInvalidRefreshToken = errors.New("invalid refresh token")
	ErrTokenReused         = errors.New("refresh token reused")
	ErrSessionHijacked     = errors.New("session hijacked")
)
