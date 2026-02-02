package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestActiveTasksCursorPagination_NewestFirst(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	listID := createListFor(t, access, "Active Cursor List")

	// create 25 active tasks
	for i := 0; i < 25; i++ {
		createTaskFor(t, access, listID, "active-task-"+itoa(i))
	}

	// page 1
	req := httptest.NewRequest(http.MethodGet, "/api/tasks/active?limit=10", nil)
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

	// page 2
	req = httptest.NewRequest(
		http.MethodGet,
		"/api/tasks/active?limit=10&cursor="+page1.NextCursor,
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+access)

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("page2 failed: %d %s", w.Code, w.Body.String())
	}

	var page2 struct {
		Items      []map[string]any `json:"items"`
		NextCursor string           `json:"next_cursor"`
		HasMore    bool             `json:"has_more"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &page2)

	if len(page2.Items) != 10 {
		t.Fatalf("expected 10 items in page2, got %d", len(page2.Items))
	}

	// ensure no overlap
	seen := map[string]bool{}
	for _, it := range page1.Items {
		seen[it["id"].(string)] = true
	}
	for _, it := range page2.Items {
		if seen[it["id"].(string)] {
			t.Fatalf("cursor paging overlap detected: %v", it["id"])
		}
	}
}

func TestActiveTasksCursorPagination_InvalidCursor(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/tasks/active?cursor=not-a-valid-cursor",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
	}
}

func TestActiveTasksCursorPagination_ListFilter(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	listA := createListFor(t, access, "List A")
	listB := createListFor(t, access, "List B")

	createTaskFor(t, access, listA, "task-a1")
	createTaskFor(t, access, listB, "task-b1")

	req := httptest.NewRequest(
		http.MethodGet,
		"/api/tasks/active?list_id="+listA,
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("list filter failed: %d %s", w.Code, w.Body.String())
	}

	var resp struct {
		Items []map[string]any `json:"items"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &resp)

	if len(resp.Items) != 1 {
		t.Fatalf("expected 1 task, got %d", len(resp.Items))
	}

	if resp.Items[0]["list_id"].(string) != listA {
		t.Fatalf("wrong list returned")
	}
}
