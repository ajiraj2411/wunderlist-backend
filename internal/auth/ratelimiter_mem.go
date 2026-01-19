package auth

import (
	"sync"
	"time"
)

type MemoryRateLimiter struct {
	limit  int
	window time.Duration

	mu      sync.Mutex
	buckets map[string]*bucket
}

type bucket struct {
	count     int
	expiresAt time.Time
}

func NewMemoryRateLimiter(limit int, window time.Duration) *MemoryRateLimiter {
	return &MemoryRateLimiter{
		limit:   limit,
		window:  window,
		buckets: make(map[string]*bucket),
	}
}

func (r *MemoryRateLimiter) Allow(key string) bool {
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

func (r *MemoryRateLimiter) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.buckets = make(map[string]*bucket)
}
