package integration

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestUserCannotDeleteOtherUsersList(t *testing.T) {
	// user A
	signupUser(t, "a@test.com")
	tokenA := loginAndGetAccessTokenFor(t, "a@test.com")

	// create list as A
	listID := createListFor(t, tokenA, "A-List")

	// user B
	signupUser(t, "b@test.com")
	tokenB := loginAndGetAccessTokenFor(t, "b@test.com")

	req := httptest.NewRequest(
		http.MethodDelete,
		"/api/lists/"+listID,
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+tokenB)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", w.Code)
	}
}
