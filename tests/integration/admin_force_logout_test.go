package integration

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminForceLogoutUser(t *testing.T) {

	adminAccess := loginAsAdmin(t)
	userID := signupSecondUser(t)

	req := httptest.NewRequest(
		http.MethodPost,
		"/api/admin/users/"+userID+"/force-logout",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+adminAccess)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestAdminForceLogoutRevokesRefreshToken(t *testing.T) {
	// ---- create normal user and login ----
	email := "victim@test.com"
	signupUser(t, email)

	accessVictim := loginAndGetAccessTokenFor(t, email)

	// login again to get refresh token also
	// (we must call /login and parse refresh, so use loginAndGetTokens style manually)
	resetLimiters(t)

	loginPayload := `{"email":"victim@test.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("victim login failed: %d %s", w.Code, w.Body.String())
	}

	// extract refresh_token from response
	body := w.Body.String()

	// quick & safe parse (no struct import)
	refresh := extractJSONField(body, "refresh_token")
	if refresh == "" {
		t.Fatalf("missing refresh_token in login response: %s", body)
	}

	// ---- login admin ----
	adminAccess := loginAsAdmin(t)

	// ---- admin force logout victim ----
	userID := getUserIDByEmail(t, email)

	req = httptest.NewRequest(
		http.MethodPost,
		"/api/admin/users/"+userID+"/force-logout",
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+adminAccess)

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("admin force logout failed: %d %s", w.Code, w.Body.String())
	}

	// ---- refresh token should now FAIL ----
	refreshReq := `{"refresh_token":"` + refresh + `"}`

	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(refreshReq))
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected refresh to fail after force logout, got %d %s", w.Code, w.Body.String())
	}

	// (Optional) victim access token might still work until expiry unless blacklisted
	_ = accessVictim
}
