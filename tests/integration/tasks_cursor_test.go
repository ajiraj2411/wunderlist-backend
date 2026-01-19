package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTasksCursorPagination_NewestFirst(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	listID := createListFor(t, access, "Cursor List")

	// create 30 tasks
	for i := 0; i < 30; i++ {
		createTaskFor(t, access, listID, "task-cursor-"+itoa(i))
	}

	// page 1
	req := httptest.NewRequest(http.MethodGet, "/api/tasks?limit=10", nil)
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
	req = httptest.NewRequest(http.MethodGet, "/api/tasks?limit=10&cursor="+page1.NextCursor, nil)
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

	// Ensure no overlap: compare IDs
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

func TestTasksCursorPagination_InvalidCursor(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	req := httptest.NewRequest(http.MethodGet, "/api/tasks?limit=10&cursor=not-a-valid-cursor", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d %s", w.Code, w.Body.String())
	}
}

// tiny helper avoiding strconv import (matches your style)
func itoa(i int) string {
	b := bytes.NewBuffer(nil)
	_, _ = b.WriteString(string(rune('0' + (i / 10))))
	_, _ = b.WriteString(string(rune('0' + (i % 10))))
	s := b.String()
	// handle <10 case
	if i < 10 {
		return s[1:]
	}
	return s
}
