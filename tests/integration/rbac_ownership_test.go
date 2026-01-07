package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

// helper: signup + login + return access token
func signupAndLogin(t *testing.T, email string) string {
	payload := `{"email":"` + email + `","password":"password123"}`

	// signup
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
		t.Fatalf("signup failed: %s", w.Body.String())
	}

	// login
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %s", w.Body.String())
	}

	var res map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res["access_token"]
}

func TestRBAC_UserCannotAccessOthersList(t *testing.T) {
	// -------- user A --------
	tokenA := signupAndLogin(t, "userA@test.com")

	// create list as A
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/lists",
		bytes.NewBufferString(`{"title":"Private List A"}`),
	)
	req.Header.Set("Authorization", "Bearer "+tokenA)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("list creation failed: %s", w.Body.String())
	}

	var createRes map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &createRes)
	listID := createRes["id"]

	if listID == "" {
		t.Fatal("list id missing")
	}

	// -------- user B --------
	tokenB := signupAndLogin(t, "userB@test.com")

	// -------- B tries to UPDATE A's list --------
	req = httptest.NewRequest(
		http.MethodPut,
		"/api/lists/"+listID,
		bytes.NewBufferString(`{"title":"Hacked"}`),
	)
	req.Header.Set("Authorization", "Bearer "+tokenB)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on update чужого list, got %d", w.Code)
	}

	// -------- B tries to DELETE A's list --------
	req = httptest.NewRequest(
		http.MethodDelete,
		"/api/lists/"+listID,
		nil,
	)
	req.Header.Set("Authorization", "Bearer "+tokenB)

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusNotFound {
		t.Fatalf("expected 404 on delete чужого list, got %d", w.Code)
	}
}
