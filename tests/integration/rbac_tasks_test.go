package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserCannotUpdateOtherUsersTask(t *testing.T) {
	// user A
	signupUser(t, "a@test.com")
	tokenA := loginAndGetAccessTokenFor(t, "a@test.com")

	listID := createListFor(t, tokenA, "A-List")
	taskID := createTaskFor(t, tokenA, listID, "A-Task")

	// user B
	signupUser(t, "b@test.com")
	tokenB := loginAndGetAccessTokenFor(t, "b@test.com")

	req := httptest.NewRequest(
		http.MethodPut,
		"/api/tasks/"+taskID,
		bytes.NewBufferString(`{"title":"Hacked"}`),
	)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
