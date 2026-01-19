package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestRefreshHijack_RevokesAllSessions(t *testing.T) {
	signupTestUser(t)

	resetLimiters(t)
	loginPayload := `{"email":"int@test.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "UA-1")
	req.Header.Set("X-Forwarded-For", "10.0.0.10")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}

	var loginRes map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &loginRes)
	refresh := loginRes["refresh_token"]
	if refresh == "" {
		t.Fatal("missing refresh token")
	}

	// attempt refresh from different UA => hijack
	resetLimiters(t)
	refreshReq := `{"refresh_token":"` + refresh + `"}`
	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(refreshReq))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "UA-HIJACKED")
	req.Header.Set("X-Forwarded-For", "10.0.0.10")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 hijack, got %d %s", w.Code, w.Body.String())
	}

	// now same refresh from correct UA should fail because sessions revoked
	resetLimiters(t)
	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(refreshReq))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "UA-1")
	req.Header.Set("X-Forwarded-For", "10.0.0.10")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected refresh to fail after hijack revocation, got %d %s", w.Code, w.Body.String())
	}
}
