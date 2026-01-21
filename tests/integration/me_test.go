package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestMe_ReturnsProfileAndSessionCount(t *testing.T) {
	signupTestUser(t)

	// login twice to create 2 sessions
	access1, _ := loginAndGetTokens(t)
	_, _ = loginAndGetTokens(t)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+access1)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}

	if extractJSONField(w.Body.String(), "email") != "int@test.com" {
		t.Fatalf("expected email int@test.com, got: %s", w.Body.String())
	}

	if extractJSONField(w.Body.String(), "active_sessions_count") == "" {
		t.Fatalf("expected active_sessions_count in response, got: %s", w.Body.String())
	}
}

func TestMe_UnauthorizedWithoutToken(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d %s", w.Code, w.Body.String())
	}
}
