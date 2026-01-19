package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"wunderlist-backend/internal/auth"
)

func TestListSessions(t *testing.T) {
	auth.ResetRateLimitersForTests(testlimiters)
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	req := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}
