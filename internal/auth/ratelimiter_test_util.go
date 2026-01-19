package auth

// ResetRateLimitersForTests clears limiter state between tests
func ResetRateLimitersForTests(limiters RateLimiterSet) {
	if limiters.Login != nil {
		limiters.Login.Reset()
	}
	if limiters.Refresh != nil {
		limiters.Refresh.Reset()
	}
	if limiters.Google != nil {
		limiters.Google.Reset()
	}

}
