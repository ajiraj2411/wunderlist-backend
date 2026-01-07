package auth

import (
	"sync"
	"time"
)

/*
RateLimiter implements a simple token-bucket rate limiter.

- Keyed by string (IP, user, etc.)
- Safe for concurrent use
- Supports test resets
*/

type RateLimiter struct {
	limit  int
	window time.Duration

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	count     int
	expiresAt time.Time
}

/*
Constructor
*/
func NewRateLimiter(limit int, window time.Duration) *RateLimiter {
	return &RateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string]*bucket),
	}
}

/*
Allow returns true if the request is allowed
*/
func (r *RateLimiter) Allow(key string) bool {
	if !rateLimitEnabled {
		return true
	}

	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	b, ok := r.buckets[key]
	if !ok || now.After(b.expiresAt) {
		r.buckets[key] = &bucket{
			count:     1,
			expiresAt: now.Add(r.window),
		}
		return true
	}

	if b.count >= r.limit {
		return false
	}

	b.count++
	return true
}

/*
Reset clears all buckets
Used ONLY in tests
*/
func (r *RateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buckets = make(map[string]*bucket)
}

//
// ==============================
// GLOBAL AUTH LIMITERS
// ==============================
//

var (
	LoginLimiter   = NewRateLimiter(5, 1*time.Minute)
	RefreshLimiter = NewRateLimiter(10, 1*time.Minute)
)

/*
ResetRateLimitersForTest clears global limiters.
Call from TestMain ONLY.
*/

func ResetRateLimitersForTest() {
	if LoginLimiter != nil {
		LoginLimiter.Reset()
	}
	if RefreshLimiter != nil {
		RefreshLimiter.Reset()
	}
}

var rateLimitEnabled = true

func EnableRateLimitForTests() {
	rateLimitEnabled = true
}

func DisableRateLimitForTests() {
	rateLimitEnabled = false
}
