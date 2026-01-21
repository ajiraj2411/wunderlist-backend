package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
	"github.com/ajiraj2411/wunderlist-backend/internal/auth"
)

func TestLogoutAllRevokesRefreshAndAccess(t *testing.T) {
	auth.ResetRateLimitersForTests(testlimiters)

	access, refresh := loginAndGetTokens(t)

	// logout all
	req := httptest.NewRequest(http.MethodPost, "/api/logout/all", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	// 🔥 refresh should now fail
	payload := `{"refresh_token":"` + refresh + `"}`

	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected refresh revoked, got %d", w.Code)
	}
}
