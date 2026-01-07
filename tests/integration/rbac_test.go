package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"wunderlist-backend/internal/auth"
)

func TestAdminRoutesBlockedForUser(t *testing.T) {
	access, _ := auth.GenerateAccessToken("user-id", "user")

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/admin/sessions",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
