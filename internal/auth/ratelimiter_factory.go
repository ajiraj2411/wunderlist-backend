package auth

import (
	"time"

	"github.com/redis/go-redis/v9"
)

type RateLimiterSet struct {
	Login   RateLimiter
	Refresh RateLimiter
}

func NewRateLimiters(rdb *redis.Client) RateLimiterSet {

	// Redis available
	if rdb != nil {
		return RateLimiterSet{
			Login:   NewRedisRateLimiter(rdb, "login", 5, time.Minute),
			Refresh: NewRedisRateLimiter(rdb, "refresh", 10, time.Minute),
		}
	}

	// Memory fallback
	return RateLimiterSet{
		Login:   NewMemoryRateLimiter(5, time.Minute),
		Refresh: NewMemoryRateLimiter(10, time.Minute),
	}
}
