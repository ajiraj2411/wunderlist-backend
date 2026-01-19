package auth

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisRateLimiter struct {
	rdb    *redis.Client
	limit  int
	window time.Duration
	prefix string
}

func NewRedisRateLimiter(
	rdb *redis.Client,
	prefix string,
	limit int,
	window time.Duration,
) *RedisRateLimiter {
	return &RedisRateLimiter{
		rdb:    rdb,
		prefix: prefix,
		limit:  limit,
		window: window,
	}
}

func (r *RedisRateLimiter) Allow(key string) bool {
	ctx := context.Background()
	redisKey := fmt.Sprintf("%s:%s", r.prefix, key)

	pipe := r.rdb.TxPipeline()
	incr := pipe.Incr(ctx, redisKey)
	pipe.Expire(ctx, redisKey, r.window)

	if _, err := pipe.Exec(ctx); err != nil {
		// fail open (do not block auth if Redis is down)
		return true
	}

	return incr.Val() <= int64(r.limit)
}

func (r *RedisRateLimiter) Reset() {
	// no-op (used only in memory limiter tests)
}
