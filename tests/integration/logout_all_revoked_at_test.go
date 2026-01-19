package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogoutAll_BlocksOldAccessImmediately(t *testing.T) {
	signupTestUser(t)

	access, _ := loginAndGetTokens(t)

	// logout all
	req := httptest.NewRequest(http.MethodPost, "/api/logout/all", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("logout all failed: %d %s", w.Code, w.Body.String())
	}

	// old access token must now fail immediately
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
		t.Fatalf("expected 401 after logout all, got %d %s", w.Code, w.Body.String())
	}
}
