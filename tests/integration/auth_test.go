package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSignupAndLoginFlow(t *testing.T) {

	signupTestUser(t)

	body := []byte(`{"email":"int@test.com","password":"password123"}`)

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}
}

func TestProtectedRouteRequiresAuth(t *testing.T) {

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/lists",
		bytes.NewBufferString(`{"title":"Work"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", w.Code)
	}
}

func TestCreateListAuthorized(t *testing.T) {

	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/lists",
		bytes.NewBufferString(`{"title":"Work"}`),
	)
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", w.Code, w.Body.String())
	}
}

func loginAndGetAccessToken(t *testing.T) string {
	resetLimiters(t)
	payload := `{"email":"int@test.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	var res map[string]string
	json.Unmarshal(w.Body.Bytes(), &res)
	return res["access_token"]
}
