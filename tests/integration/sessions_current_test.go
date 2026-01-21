package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogoutCurrentSession_RevokesAccessImmediately(t *testing.T) {
	signupTestUser(t)

	access, refresh := loginAndGetTokens(t)
	if access == "" || refresh == "" {
		t.Fatal("missing tokens")
	}

	// logout current session
	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/current", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d %s", w.Code, w.Body.String())
	}

	// access token should fail immediately
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
		t.Fatalf("expected 401 after logout current, got %d %s", w.Code, w.Body.String())
	}

	// refresh should also fail
	payload := `{"refresh_token":"` + refresh + `"}`
	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "10.0.0.99")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected refresh to fail after logout current, got %d %s", w.Code, w.Body.String())
	}
}
