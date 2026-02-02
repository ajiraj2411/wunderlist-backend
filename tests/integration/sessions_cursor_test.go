package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSessionsCursorPagination_NewestFirst(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	// create multiple sessions
	for i := 0; i < 5; i++ {
		loginAndGetAccessToken(t)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/sessions?limit=2", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("page1 failed: %d %s", w.Code, w.Body.String())
	}

	var page struct {
		Items      []map[string]any `json:"items"`
		NextCursor string           `json:"next_cursor"`
		HasMore    bool             `json:"has_more"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &page)

	if len(page.Items) != 2 || page.NextCursor == "" {
		t.Fatalf("invalid response: %s", w.Body.String())
	}
}

func TestSessionsCursorPagination_InvalidCursor(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/sessions?cursor=not-a-valid-cursor",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
	}
}
