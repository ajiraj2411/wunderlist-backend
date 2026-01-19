package auth

import "time"

func IsJWTRevoked(jti string) bool {
	if jwtBlacklist == nil {
		return false
	}
	return jwtBlacklist.IsRevoked(jti)
}

func RevokeJWT(jti string, exp time.Time) {
	if jwtBlacklist != nil {
		_ = jwtBlacklist.Revoke(jti, exp)
	}
}
