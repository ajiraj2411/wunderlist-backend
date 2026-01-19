package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestLogoutCurrentSession_RevokesAccessImmediately(t *testing.T) {
	signupTestUser(t)

	access, _ := loginAndGetTokens(t)

	// logout current session
	req := httptest.NewRequest(http.MethodDelete, "/api/sessions/current", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusNoContent {
		t.Fatalf("expected 204, got %d %s", w.Code, w.Body.String())
	}

	// same token must fail immediately
	req = httptest.NewRequest(http.MethodPost, "/api/lists", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after revoke, got %d %s", w.Code, w.Body.String())
	}
}
