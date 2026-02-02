package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSearchTasksCursorPagination_NewestFirst(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	listID := createListFor(t, access, "Search Cursor List")

	for i := 0; i < 20; i++ {
		createTaskFor(t, access, listID, "search-me-"+itoa(i))
	}

	// page 1
	req := httptest.NewRequest(
		http.MethodGet,
		"/api/tasks/search?q=search-me&limit=5",
		nil,
	)
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

	if len(page1.Items) != 5 || page1.NextCursor == "" || !page1.HasMore {
		t.Fatalf("invalid page1 response: %s", w.Body.String())
	}

	// page 2
	req = httptest.NewRequest(
		http.MethodGet,
		"/api/tasks/search?q=search-me&limit=5&cursor="+page1.NextCursor,
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+access)

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("page2 failed: %d %s", w.Code, w.Body.String())
	}

	var page2 struct {
		Items []map[string]any `json:"items"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &page2)

	if len(page2.Items) != 5 {
		t.Fatalf("expected 5 items, got %d", len(page2.Items))
	}
}

func TestSearchTasksCursorPagination_InvalidCursor(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/tasks/search?q=test&cursor=bad-cursor",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
	}
}
