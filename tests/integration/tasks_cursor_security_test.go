package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestTasksCursorPagination_TamperedCursorRejected(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	listID := createListFor(t, access, "Secure Cursor List")
	createTaskFor(t, access, listID, "task-1")
	createTaskFor(t, access, listID, "task-2")

	// get valid cursor
	req := httptest.NewRequest(http.MethodGet, "/api/tasks?limit=1", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	var page struct {
		NextCursor string `json:"next_cursor"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &page)

	if page.NextCursor == "" {
		t.Fatal("expected cursor")
	}

	// 🔥 tamper with cursor
	tampered := page.NextCursor[:len(page.NextCursor)-2] + "xx"

	req = httptest.NewRequest(
		http.MethodGet,
		"/api/tasks?limit=1&cursor="+tampered,
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+access)

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for tampered cursor, got %d", w.Code)
	}
}

func TestTasksCursorPagination_UserMismatchRejected(t *testing.T) {
	// user A
	signupTestUser(t)
	accessA := loginAndGetAccessToken(t)

	listID := createListFor(t, accessA, "User A List")
	createTaskFor(t, accessA, listID, "task-A")

	req := httptest.NewRequest(http.MethodGet, "/api/tasks?limit=1", nil)
	req.Header.Set("Authorization", "Bearer "+accessA)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	var page struct {
		NextCursor string `json:"next_cursor"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &page)

	if page.NextCursor == "" {
		t.Fatal("expected cursor")
	}

	// user B
	signupSecondTestUser(t)
	accessB := loginAndGetAccessTokenForSecondUser(t)

	req = httptest.NewRequest(
		http.MethodGet,
		"/api/tasks?limit=1&cursor="+page.NextCursor,
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+accessB)

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for cursor user mismatch, got %d", w.Code)
	}
}
