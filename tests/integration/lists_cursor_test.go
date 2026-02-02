package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListsCursorPagination_NewestFirst(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	for i := 0; i < 25; i++ {
		createListFor(t, access, "list-"+itoa(i))
	}

	req := httptest.NewRequest(http.MethodGet, "/api/lists?limit=10", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("page1 failed: %d %s", w.Code, w.Body.String())
	}

	var page1 struct {
		Items      []map[string]any `json:"items"`
		NextCursor string           `json:"next_cursor"`
		HasMore    bool             `json:"has_more"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &page1)

	if len(page1.Items) != 10 || page1.NextCursor == "" || !page1.HasMore {
		t.Fatalf("invalid page1 response: %s", w.Body.String())
	}

	req = httptest.NewRequest(
		http.MethodGet,
		"/api/lists?limit=10&cursor="+page1.NextCursor,
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+access)

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("page2 failed: %d %s", w.Code, w.Body.String())
	}
}

func TestListsCursorPagination_InvalidCursor(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/lists?cursor=not-a-valid-cursor",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
	}
}
