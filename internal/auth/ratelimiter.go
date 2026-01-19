package auth

type RateLimiter interface {
	Allow(key string) bool
	Reset() // no-op for Redis, used in tests
}
