package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

var TestRouter *gin.Engine

func signupTestUser(t *testing.T) {
	payload := map[string]string{
		"email":    "int@test.com",
		"password": "password123",
	}

	body, _ := json.Marshal(payload)

	req := httptest.NewRequest(
		http.MethodPost,
		"/signup",
		bytes.NewReader(body),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	// 201 = created
	// 409 = already exists (acceptable for repeat runs)
	if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
		t.Fatalf("signup failed: %s", w.Body.String())
	}
}

func signupUser(t *testing.T, email string) {
	payload := `{"email":"` + email + `","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)
}

func signupUserWithPassword(t *testing.T, email string, password string) {
	t.Helper()

	payload := `{"email":"` + email + `","password":"` + password + `"}`

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("signup failed: %d %s", w.Code, w.Body.String())
	}
}

func loginAndGetAccessTokenFor(t *testing.T, email string) string {
	payload := `{"email":"` + email + `","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	var res map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res["access_token"]
}

func createListFor(t *testing.T, token, title string) string {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/lists",
		bytes.NewBufferString(`{"title":"`+title+`"}`),
	)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	var res map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res["id"]
}

func createTaskFor(t *testing.T, token, listID, title string) string {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/tasks",
		bytes.NewBufferString(
			`{"title":"`+title+`","list_id":"`+listID+`"}`,
		),
	)
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	var res map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res["id"]
}

func createListAndGetID(t *testing.T, access string) string {
	req := httptest.NewRequest(
		http.MethodPost,
		"/api/lists",
		bytes.NewBufferString(`{"title":"Concurrency List"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+access)

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("list creation failed: %s", w.Body.String())
	}

	var res map[string]string
	json.Unmarshal(w.Body.Bytes(), &res)
	return res["id"]
}

// ===========================
// SECOND TEST USER HELPERS
// ===========================

const secondTestEmail = "second@test.com"
const secondTestPassword = "password123"

func signupSecondTestUser(t *testing.T) {
	t.Helper()

	payload := `{"email":"` + secondTestEmail + `","password":"` + secondTestPassword + `"}`

	req := httptest.NewRequest(http.MethodPost, "/signup", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
		t.Fatalf("second signup failed: %d %s", w.Code, w.Body.String())
	}
}

func loginAndGetAccessTokenForSecondUser(t *testing.T) string {
	t.Helper()

	payload := `{"email":"` + secondTestEmail + `","password":"` + secondTestPassword + `"}`

	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("second login failed: %d %s", w.Code, w.Body.String())
	}

	var res struct {
		AccessToken string `json:"access_token"`
	}
	_ = json.Unmarshal(w.Body.Bytes(), &res)

	if res.AccessToken == "" {
		t.Fatal("missing access token for second user")
	}

	return res.AccessToken
}
