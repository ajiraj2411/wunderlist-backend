package integration

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

/*
	TEST: Login rate limit exceeded
*/

func TestLoginRateLimitExceeded(t *testing.T) {

	signupTestUser(t)

	loginPayload := `{"email":"int@test.com","password":"password123"}`

	for i := 0; i < 6; i++ {
		req := httptest.NewRequest(
			http.MethodPost,
			"/login",
			bytes.NewBufferString(loginPayload),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "10.0.0.1")

		w := httptest.NewRecorder()
		TestRouter.ServeHTTP(w, req)

		if i < 5 && w.Code != http.StatusOK {
			t.Fatalf("expected login OK, got %d: %s", w.Code, w.Body.String())
		}

		if i >= 5 && w.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429, got %d", w.Code)
		}
	}
}

/*
	TEST: Refresh rate limit exceeded (ROTATING TOKEN)
*/

func TestRefreshRateLimitExceeded(t *testing.T) {

	signupTestUser(t)

	// ---- LOGIN ----
	loginPayload := `{"email":"int@test.com","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "10.0.0.2")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %s", w.Body.String())
	}

	var loginRes map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &loginRes)

	refresh := loginRes["refresh_token"]

	// ---- REFRESH WITH ROTATION ----
	for i := 0; i < 11; i++ {

		payload := `{"refresh_token":"` + refresh + `"}`

		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "10.0.0.2")

		w := httptest.NewRecorder()
		TestRouter.ServeHTTP(w, req)

		if i < 10 {
			if w.Code != http.StatusOK {
				t.Fatalf("expected refresh OK, got %d: %s", w.Code, w.Body.String())
			}

			var res map[string]string
			_ = json.Unmarshal(w.Body.Bytes(), &res)

			refresh = res["refresh_token"] // 🔁 rotated token
			continue
		}

		if w.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429, got %d", w.Code)
		}
	}
}
