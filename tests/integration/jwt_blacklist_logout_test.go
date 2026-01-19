package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestJWTBlacklistLogoutSingleSession(t *testing.T) {
	signupTestUser(t)

	access, _ := loginAndGetTokens(t)

	// ---- logout (single session) ----
	req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("logout failed: %d %s", w.Code, w.Body.String())
	}

	// ---- use same access token again (must fail due to blacklist) ----
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
		t.Fatalf("expected 401 (revoked token), got %d %s", w.Code, w.Body.String())
	}
}
