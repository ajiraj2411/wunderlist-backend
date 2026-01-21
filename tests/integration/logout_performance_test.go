package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestLogoutPerformance_WithManySessions(t *testing.T) {
	requirePerfEnabled(t)

	// ✅ Use unique user to avoid cross-test pollution
	email := "perf-logout-" + time.Now().UTC().Format("20060102150405.000") + "@test.com"
	signupUser(t, email)

	// ---- login to get real access + refresh ----
	resetLimiters(t)
	loginPayload := `{"email":"` + email + `","password":"password123"}`
	req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginPayload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("login failed: %d %s", w.Code, w.Body.String())
	}

	access := extractJSONField(w.Body.String(), "access_token")
	refresh := extractJSONField(w.Body.String(), "refresh_token")
	if access == "" || refresh == "" {
		t.Fatalf("missing tokens from login: %s", w.Body.String())
	}

	// ---- get user id ----
	userIDHex := getUserIDByEmail(t, email)
	userID, _ := primitive.ObjectIDFromHex(userIDHex)

	// ---- cleanup existing sessions for this user (safety) ----
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	_, _ = TestDB.Collection("sessions").DeleteMany(ctx, bson.M{"user_id": userID})

	// ---- insert lots of sessions (simulate production DB size) ----
	//
	// IMPORTANT:
	// - Do NOT bcrypt-hash synthetic tokens here.
	// - bcrypt is expensive and makes perf tests flaky/slow.
	// - We only care about DB size & indexed lookup regressions.
	const sessionsToInsert = 1500
	const batchSize = 300

	opts := options.InsertMany().SetOrdered(false)

	now := time.Now().UTC()
	batch := make([]interface{}, 0, batchSize)

	for i := 0; i < sessionsToInsert; i++ {
		raw := "synthetic-refresh-" + primitive.NewObjectID().Hex()

		s := models.Session{
			ID:     primitive.NewObjectID(),
			UserID: userID,

			// ✅ fast fake hash (we are NOT validating these refresh tokens)
			TokenHash: "$2a$10$perf.synthetic.hash.not.used",
			TokenSHA:  shaHex(raw),

			Role:      "user",
			UserAgent: "load-test",
			IPAddress: "127.0.0.1",
			CreatedAt: now,
			ExpiresAt: now.Add(7 * 24 * time.Hour),
		}

		batch = append(batch, s)

		if len(batch) == batchSize {
			_, err := TestDB.Collection("sessions").InsertMany(ctx, batch, opts)
			if err != nil {
				t.Fatalf("InsertMany failed: %v", err)
			}
			batch = batch[:0]
		}
	}

	if len(batch) > 0 {
		_, err := TestDB.Collection("sessions").InsertMany(ctx, batch, opts)
		if err != nil {
			t.Fatalf("final InsertMany failed: %v", err)
		}
	}

	// ---- benchmark logout ----
	// logout should still be fast (O(1) delete by token_sha)
	resetLimiters(t)

	times := make([]time.Duration, 0, 3)

	for i := 0; i < 3; i++ {
		start := time.Now()

		req := httptest.NewRequest(http.MethodPost, "/api/logout", nil)
		req.Header.Set("Authorization", "Bearer "+access)
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		TestRouter.ServeHTTP(w, req)

		elapsed := time.Since(start)
		times = append(times, elapsed)

		if w.Code != http.StatusOK {
			t.Fatalf("logout failed (iteration %d): %d %s", i, w.Code, w.Body.String())
		}

		// After logout the access token is revoked/blacklisted.
		// Re-login to get new access & refresh for the next iteration.
		resetLimiters(t)
		req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBufferString(loginPayload))
		req.Header.Set("Content-Type", "application/json")

		w = httptest.NewRecorder()
		TestRouter.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("re-login failed: %d %s", w.Code, w.Body.String())
		}

		access = extractJSONField(w.Body.String(), "access_token")
		if access == "" {
			t.Fatalf("missing access_token on re-login: %s", w.Body.String())
		}
	}

	med := medianDuration3(times)

	if med > 1200*time.Millisecond {
		t.Fatalf("logout took too long (median of 3): %s (likely scan regression). all=%v", med, times)
	}

	t.Logf("✅ logout duration median=%s (runs=%v) with %d extra sessions", med, times, sessionsToInsert)
}

func shaHex(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
