package integration

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetMe_Unauthorized(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d %s", w.Code, w.Body.String())
	}
}

func TestGetMe_ReturnsProfile(t *testing.T) {
	signupTestUser(t)
	access := loginAndGetAccessToken(t)

	req := httptest.NewRequest(http.MethodGet, "/api/me", nil)
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d %s", w.Code, w.Body.String())
	}

	var res struct {
		ID        string `json:"id"`
		Email     string `json:"email"`
		Role      string `json:"role"`
		CreatedAt string `json:"created_at"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid json response: %v", err)
	}

	if res.ID == "" || res.Email == "" || res.Role == "" || res.CreatedAt == "" {
		t.Fatalf("missing fields: %s", w.Body.String())
	}

	if res.Email != "int@test.com" {
		t.Fatalf("expected email int@test.com got %s", res.Email)
	}
}
