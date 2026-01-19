package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wunderlist-backend/internal/auth"

	"go.mongodb.org/mongo-driver/bson"
)

func TestGoogleOAuthLogin_CreatesUserAndReturnsTokens(t *testing.T) {
	resetLimiters(t)

	// ✅ stub google verifier
	old := auth.GetGoogleVerifier()
	auth.SetGoogleVerifierForTests(func(ctx context.Context, idToken string) (*auth.GoogleUser, error) {
		return &auth.GoogleUser{
			GoogleID:      "google-sub-123",
			Email:         "googleuser@test.com",
			Name:          "Google User",
			EmailVerified: true,
		}, nil
	})
	t.Cleanup(func() { auth.SetGoogleVerifierForTests(old) })

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/google",
		bytes.NewBufferString(`{"id_token":"fake"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "10.0.0.9")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", w.Code, w.Body.String())
	}

	// validate response tokens
	var res map[string]interface{}
	_ = json.Unmarshal(w.Body.Bytes(), &res)

	if res["access_token"] == "" || res["refresh_token"] == "" {
		t.Fatalf("expected tokens, got: %s", w.Body.String())
	}

	// ensure user created
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	count, err := TestDB.Collection("users").CountDocuments(ctx, bson.M{
		"email":         "googleuser@test.com",
		"auth_provider": "google",
		"google_sub":    "google-sub-123",
	})
	if err != nil {
		t.Fatalf("count users failed: %v", err)
	}
	if count != 1 {
		t.Fatalf("expected google user created, count=%d", count)
	}
}

func TestGoogleOAuthLogin_EmailNotVerifiedRejected(t *testing.T) {
	resetLimiters(t)

	old := auth.GetGoogleVerifier()
	auth.SetGoogleVerifierForTests(func(ctx context.Context, idToken string) (*auth.GoogleUser, error) {
		return &auth.GoogleUser{
			GoogleID:      "sub-unverified",
			Email:         "unverified@test.com",
			Name:          "Nope",
			EmailVerified: false,
		}, nil
	})
	t.Cleanup(func() { auth.SetGoogleVerifierForTests(old) })

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/google",
		bytes.NewBufferString(`{"id_token":"fake"}`),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("google email not verified")) {
		t.Fatalf("expected email verified error, got: %s", w.Body.String())
	}
}

func TestGoogleOAuthLogin_PasswordAccountTakeoverBlocked(t *testing.T) {
	resetLimiters(t)

	// create password account first
	signupUser(t, "victim@test.com")

	old := auth.GetGoogleVerifier()
	auth.SetGoogleVerifierForTests(func(ctx context.Context, idToken string) (*auth.GoogleUser, error) {
		return &auth.GoogleUser{
			GoogleID:      "google-sub-hacker",
			Email:         "victim@test.com", // attempt same email
			Name:          "Hacker",
			EmailVerified: true,
		}, nil
	})
	t.Cleanup(func() { auth.SetGoogleVerifierForTests(old) })

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/google",
		bytes.NewBufferString(`{"id_token":"fake"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Forwarded-For", "10.0.0.66")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	// your handler returns unauthorized when FindOrCreateGoogleUser blocks
	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", w.Code, w.Body.String())
	}
}

func TestGoogleOAuthLogin_RateLimited(t *testing.T) {
	resetLimiters(t)

	old := auth.GetGoogleVerifier()
	auth.SetGoogleVerifierForTests(func(ctx context.Context, idToken string) (*auth.GoogleUser, error) {
		return &auth.GoogleUser{
			GoogleID:      "sub-rate",
			Email:         "ratelimit@test.com",
			Name:          "Rate Limit",
			EmailVerified: true,
		}, nil
	})
	t.Cleanup(func() { auth.SetGoogleVerifierForTests(old) })

	// 5 allowed, 6th blocked (limit=5/min)
	for i := 0; i < 6; i++ {
		req := httptest.NewRequest(
			http.MethodPost,
			"/auth/google",
			bytes.NewBufferString(`{"id_token":"fake"}`),
		)
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Forwarded-For", "10.0.0.88") // same IP key
		req.Header.Set("User-Agent", "test-agent")

		w := httptest.NewRecorder()
		TestRouter.ServeHTTP(w, req)

		if i < 5 && w.Code != http.StatusOK {
			t.Fatalf("expected 200 for attempt %d, got %d: %s", i, w.Code, w.Body.String())
		}

		if i >= 5 && w.Code != http.StatusTooManyRequests {
			t.Fatalf("expected 429 at attempt %d, got %d: %s", i, w.Code, w.Body.String())
		}
	}
}

func TestGoogleOAuthLogin_RefreshWorksWithBoundUAIP(t *testing.T) {
	resetLimiters(t)

	// ✅ stub google verifier
	old := auth.GetGoogleVerifier()
	auth.SetGoogleVerifierForTests(func(ctx context.Context, idToken string) (*auth.GoogleUser, error) {
		return &auth.GoogleUser{
			GoogleID:      "google-sub-refresh-1",
			Email:         "oauthrefresh@test.com",
			Name:          "OAuth Refresh",
			EmailVerified: true,
		}, nil
	})
	t.Cleanup(func() { auth.SetGoogleVerifierForTests(old) })

	// ---- Google Login (creates refresh session bound to UA + IP) ----
	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/google",
		bytes.NewBufferString(`{"id_token":"fake"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "oauth-agent")
	req.Header.Set("X-Forwarded-For", "10.10.10.10")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("google login failed: %d %s", w.Code, w.Body.String())
	}

	var loginRes map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &loginRes)

	refresh := loginRes["refresh_token"]
	if refresh == "" {
		t.Fatalf("missing refresh_token: %s", w.Body.String())
	}

	// ---- Refresh with SAME UA/IP (must pass) ----
	resetLimiters(t)

	refreshPayload := `{"refresh_token":"` + refresh + `"}`

	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(refreshPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "oauth-agent")
	req.Header.Set("X-Forwarded-For", "10.10.10.10")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected refresh OK, got %d %s", w.Code, w.Body.String())
	}

	// ✅ IMPORTANT: capture rotated refresh token from response
	var refreshRes map[string]string
	_ = json.Unmarshal(w.Body.Bytes(), &refreshRes)

	rotated := refreshRes["refresh_token"]
	if rotated == "" {
		t.Fatalf("missing rotated refresh_token: %s", w.Body.String())
	}

	// ---- Refresh with different UA/IP (must hijack + revoke all sessions) ----
	resetLimiters(t)

	// 🔥 use rotated token, not old token (old one is already deleted)
	hijackPayload := `{"refresh_token":"` + rotated + `"}`

	req = httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(hijackPayload))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "hacker-agent")       // mismatch
	req.Header.Set("X-Forwarded-For", "123.123.123.1") // mismatch

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 hijack block, got %d %s", w.Code, w.Body.String())
	}

	if !bytes.Contains(w.Body.Bytes(), []byte("session hijacked")) {
		t.Fatalf("expected session hijacked error, got: %s", w.Body.String())
	}
}

func TestGoogleOAuthLogin_SubMismatchBlocked(t *testing.T) {
	resetLimiters(t)

	// 1) First Google login creates the account with sub A
	old := auth.GetGoogleVerifier()
	auth.SetGoogleVerifierForTests(func(ctx context.Context, idToken string) (*auth.GoogleUser, error) {
		return &auth.GoogleUser{
			GoogleID:      "google-sub-A",
			Email:         "submismatch@test.com",
			Name:          "Sub A",
			EmailVerified: true,
		}, nil
	})
	t.Cleanup(func() { auth.SetGoogleVerifierForTests(old) })

	req := httptest.NewRequest(
		http.MethodPost,
		"/auth/google",
		bytes.NewBufferString(`{"id_token":"fake1"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "10.0.0.77")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("google login A failed: %d %s", w.Code, w.Body.String())
	}

	// 2) Second login attempt with same email but sub B must be blocked
	resetLimiters(t)

	auth.SetGoogleVerifierForTests(func(ctx context.Context, idToken string) (*auth.GoogleUser, error) {
		return &auth.GoogleUser{
			GoogleID:      "google-sub-B", // 🔥 different sub
			Email:         "submismatch@test.com",
			Name:          "Sub B",
			EmailVerified: true,
		}, nil
	})

	req = httptest.NewRequest(
		http.MethodPost,
		"/auth/google",
		bytes.NewBufferString(`{"id_token":"fake2"}`),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "10.0.0.77")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for sub mismatch, got %d %s", w.Code, w.Body.String())
	}

	// your FindOrCreateGoogleUser returns: "google identity mismatch"
	if !bytes.Contains(w.Body.Bytes(), []byte("google identity mismatch")) {
		t.Fatalf("expected sub mismatch error, got: %s", w.Body.String())
	}
}
