package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type JWTBlacklist struct {
	rdb *redis.Client
}

func NewJWTBlacklist(rdb *redis.Client) *JWTBlacklist {
	if rdb == nil {
		return nil
	}
	return &JWTBlacklist{rdb: rdb}
}

func (b *JWTBlacklist) key(jti string) string {
	return fmt.Sprintf("jwt:blacklist:%s", jti)
}

// Revoke token until it naturally expires
func (b *JWTBlacklist) Revoke(jti string, exp time.Time) error {
	if b == nil {
		return nil // fail-open
	}

	ttl := time.Until(exp)
	if ttl <= 0 {
		return nil
	}

	return b.rdb.Set(
		context.Background(),
		b.key(jti),
		"1",
		ttl,
	).Err()
}

// Check if token is revoked
func (b *JWTBlacklist) IsRevoked(jti string) bool {
	if b == nil {
		return false
	}

	val, err := b.rdb.Exists(
		context.Background(),
		b.key(jti),
	).Result()

	return err == nil && val == 1
}
