package integration

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
	"github.com/ajiraj2411/wunderlist-backend/internal/auth"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// resetLimiters clears memory rate limiter buckets so tests don't affect each other.
// Without this, tests become flaky because they share the same in-memory limiter state.
func resetLimiters(t *testing.T) {
	t.Helper()

	// reset memory rate limiter buckets
	auth.ResetRateLimitersForTests(testlimiters)

	// ✅ reset redis state: blacklist + revoked_at keys
	if TestRedis != nil {
		_ = TestRedis.FlushDB(context.Background()).Err()
	}

	t.Cleanup(func() {
		auth.ResetRateLimitersForTests(testlimiters)
		if TestRedis != nil {
			_ = TestRedis.FlushDB(context.Background()).Err()
		}
	})
}

/*
internal test context helper
*/
func ctx() context.Context {
	c, _ := context.WithTimeout(context.Background(), 5*time.Second)
	return c
}

/*
loginAndGetTokens
- logs in default test user
- returns access + refresh tokens
*/
func loginAndGetTokens(t *testing.T) (access string, refresh string) {
	t.Helper()
	resetLimiters(t)

	payload := `{"email":"int@test.com","password":"password123"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBufferString(payload),
	)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "test-agent")
	req.Header.Set("X-Forwarded-For", "10.0.0.99")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}

	var res struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	}

	if err := json.Unmarshal(w.Body.Bytes(), &res); err != nil {
		t.Fatalf("invalid login response: %v", err)
	}

	return res.AccessToken, res.RefreshToken
}

/*
loginAsAdmin
- creates admin user if not exists
- promotes to admin via DB
- returns admin access token
*/
func loginAsAdmin(t *testing.T) string {
	t.Helper()
	resetLimiters(t)
	payload := `{"email":"admin@test.com","password":"password123"}`

	// signup
	req := httptest.NewRequest(
		http.MethodPost,
		"/signup",
		bytes.NewBufferString(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)
	// 201 or 409 both OK

	// promote to admin (DB-level)
	_, err := TestDB.Collection("users").UpdateOne(
		ctx(),
		bson.M{"email": "admin@test.com"},
		bson.M{"$set": bson.M{"role": "admin"}},
	)
	if err != nil {
		t.Fatalf("failed to promote admin: %v", err)
	}

	// login
	req = httptest.NewRequest(
		http.MethodPost,
		"/login",
		bytes.NewBufferString(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	w = httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("admin login failed: %s", w.Body.String())
	}

	var res struct {
		AccessToken string `json:"access_token"`
	}

	_ = json.Unmarshal(w.Body.Bytes(), &res)
	return res.AccessToken
}

/*
signupSecondUser
- creates another normal user
- returns userID as hex string
*/
func signupSecondUser(t *testing.T) string {
	t.Helper()

	payload := `{"email":"user2@test.com","password":"password123"}`

	req := httptest.NewRequest(
		http.MethodPost,
		"/signup",
		bytes.NewBufferString(payload),
	)
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusCreated && w.Code != http.StatusConflict {
		t.Fatalf("signup second user failed: %s", w.Body.String())
	}

	// fetch user id
	var result struct {
		ID primitive.ObjectID `bson:"_id"`
	}

	err := TestDB.Collection("users").
		FindOne(ctx(), bson.M{"email": "user2@test.com"}).
		Decode(&result)

	if err != nil {
		t.Fatalf("failed to fetch second user: %v", err)
	}

	return result.ID.Hex()
}
