package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestAdminForceLogout_BlocksAccessImmediately(t *testing.T) {
	// victim
	email := "victim_access@test.com"
	signupUser(t, email)

	resetLimiters(t)
	loginPayload := `{"email":"` + email + `","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("victim login failed: %d %s", w.Code, w.Body.String())
	}

	var res struct {
		Access string `json:"access_token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	victimAccess := res.Access
	if victimAccess == "" {
		t.Fatal("missing victim access token")
	}

	// admin
	adminAccess := loginAsAdmin(t)
	userID := getUserIDByEmail(t, email)

	// force logout
	req = httptest.NewRequest(http.MethodPost, "/api/admin/users/"+userID+"/force-logout", nil)
	req.Header.Set("Authorization", "Bearer "+adminAccess)

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("force logout failed: %d %s", w.Code, w.Body.String())
	}

	// victim old access token must now fail immediately
	req = httptest.NewRequest(
		http.MethodPost,
		"/api/lists",
		bytes.NewBufferString(`{"title":"should fail"}`),
	)
	req.Header.Set("Authorization", "Bearer "+victimAccess)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 after admin force logout, got %d %s", w.Code, w.Body.String())
	}
}
