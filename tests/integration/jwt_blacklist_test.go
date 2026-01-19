package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"wunderlist-backend/internal/auth"
)

func TestJWTBlacklistLogoutAll(t *testing.T) {
	auth.ResetRateLimitersForTests(testlimiters)

	signupTestUser(t)
	access, _ := loginAndGetTokens(t)

	// logout all
	req := httptest.NewRequest(http.MethodPost, "/api/logout/all", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("logout all failed: %s", w.Body.String())
	}

	// try using revoked token
	req = httptest.NewRequest(
		http.MethodPost,
		"/api/lists",
		bytes.NewBufferString(`{"title":"should fail"}`),
	)
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected revoked token to fail, got %d", w.Code)
	}
}
