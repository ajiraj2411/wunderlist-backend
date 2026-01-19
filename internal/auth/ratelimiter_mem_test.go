package auth

import (
	"testing"
	"time"
)

func TestMemoryRateLimiter(t *testing.T) {
	limiter := NewMemoryRateLimiter(2, time.Second)

	if !limiter.Allow("ip") {
		t.Fatal("first allowed")
	}
	if !limiter.Allow("ip") {
		t.Fatal("second allowed")
	}
	if limiter.Allow("ip") {
		t.Fatal("third should be blocked")
	}

	// reset window
	time.Sleep(time.Second + 10*time.Millisecond)

	if !limiter.Allow("ip") {
		t.Fatal("after window reset should allow")
	}
}
