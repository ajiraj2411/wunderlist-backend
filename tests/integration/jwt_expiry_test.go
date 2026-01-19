package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wunderlist-backend/internal/auth"
)

func TestAccessTokenExpiryAndRefresh(t *testing.T) {
	// 🔑 Override JWT TTLs for this test only
	auth.InitJWT(
		"test-secret",
		2*time.Second, // VERY SHORT access TTL
		1*time.Minute, // refresh TTL
	)

	defer auth.InitJWT(
		"test-secret",
		15*time.Second, // VERY SHORT access TTL
		7*24*time.Hour, // refresh TTL
	)

	// ---------- Signup ----------
	signup := `{"email":"expiry@test.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(signup))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)
	if w.Code != http.StatusCreated {
		t.Fatalf("signup failed: %s", w.Body.String())
	}

	// ---------- Login ----------
	resetLimiters(t)
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(signup))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %s", w.Body.String())
	}

	var loginRes map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &loginRes)

	access := loginRes["access_token"]
	refresh := loginRes["refresh_token"]

	if access == "" || refresh == "" {
		t.Fatal("tokens missing")
	}

	// ---------- Wait for access token to expire ----------
	time.Sleep(3 * time.Second)

	// ---------- Access protected route (should FAIL) ----------
	req = httptest.NewRequest(
		http.MethodPost,
		"/api/lists",
		bytes.NewBufferString(`{"title":"Expired Test"}`),
	)
	req.Header.Set("Authorization", "Bearer "+access)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d", w.Code)
	}

	// ---------- Refresh token ----------
	refreshReq := `{"refresh_token":"` + refresh + `"}`

	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(refreshReq))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("refresh failed: %s", w.Body.String())
	}

	var refreshRes map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &refreshRes)

	newAccess := refreshRes["access_token"]
	newRefresh := refreshRes["refresh_token"]

	if newAccess == "" || newRefresh == "" {
		t.Fatal("refresh did not issue new tokens")
	}

	// ---------- New access token should work ----------
	req = httptest.NewRequest(
		http.MethodPost,
		"/api/lists",
		bytes.NewBufferString(`{"title":"New Token Works"}`),
	)
	req.Header.Set("Authorization", "Bearer "+newAccess)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("new access token failed: %d %s", w.Code, w.Body.String())
	}

	// ---------- Old refresh token should now FAIL ----------
	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(refreshReq))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected refresh reuse rejection, got %d", w.Code)
	}
}
