package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRefreshTokenRotation(t *testing.T) {

	signupTestUser(t)

	// ---- login ----
	loginReq := `{"email":"int@test.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginReq))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %s", w.Body.String())
	}

	var loginRes map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &loginRes)

	refresh := loginRes["refresh_token"]

	// ---- refresh ----
	refreshReq := `{"refresh_token":"` + refresh + `"}`

	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(refreshReq))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("refresh failed: %s", w.Body.String())
	}
}
