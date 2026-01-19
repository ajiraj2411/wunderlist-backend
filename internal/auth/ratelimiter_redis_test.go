package auth

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

func TestRedisRateLimiter(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	limiter := NewRedisRateLimiter(rdb, "test", 2, time.Second)

	if !limiter.Allow("ip") {
		t.Fatal("first request should be allowed")
	}
	if !limiter.Allow("ip") {
		t.Fatal("second request should be allowed")
	}
	if limiter.Allow("ip") {
		t.Fatal("third request should be blocked")
	}
}
