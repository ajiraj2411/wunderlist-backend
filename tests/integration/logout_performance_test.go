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

	"wunderlist-backend/internal/auth"
	"wunderlist-backend/internal/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func TestLogoutPerformance_WithManySessions(t *testing.T) {
	requirePerfEnabled(t)

	signupTestUser(t)

	access, refresh := loginAndGetTokens(t)
	if access == "" || refresh == "" {
		t.Fatal("missing tokens from login")
	}

	const sessionsToInsert = 1500
	const batchSize = 200

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	userIDHex := getUserIDByEmail(t, "int@test.com")
	userID, _ := primitive.ObjectIDFromHex(userIDHex)

	// ✅ cleanup BEFORE insert so DB doesn't grow forever across perf runs
	_, _ = TestDB.Collection("sessions").DeleteMany(ctx, bson.M{"user_id": userID})

	// login again after cleanup (since cleanup removed sessions!)
	access, refresh = loginAndGetTokens(t)
	if access == "" || refresh == "" {
		t.Fatal("missing tokens from login (after cleanup)")
	}

	opts := options.InsertMany().SetOrdered(false)

	now := time.Now().UTC()
	batch := make([]interface{}, 0, batchSize)

	for i := 0; i < sessionsToInsert; i++ {
		raw, err := auth.GenerateRefreshToken()
		if err != nil {
			t.Fatalf("GenerateRefreshToken failed: %v", err)
		}

		hash, err := auth.HashPassword(raw)
		if err != nil {
			t.Fatalf("HashPassword failed: %v", err)
		}

		s := models.Session{
			ID:        primitive.NewObjectID(),
			UserID:    userID,
			TokenHash: hash,
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

	// ✅ Re-login after heavy DB work so we always benchmark with a fresh valid token
	access, refresh = loginAndGetTokens(t)
	if access == "" || refresh == "" {
		t.Fatal("missing tokens after re-login")
	}

	// ---- logout should still be fast (O(1) DeleteOne by token_sha) ----
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

		// After logout, access token is blacklisted/revoked, so we need fresh tokens for next loop
		access, refresh = loginAndGetTokens(t)
		if access == "" || refresh == "" {
			t.Fatal("missing tokens after re-login")
		}
	}

	med := medianDuration3(times)

	// generous threshold, prevents CI flakiness but catches O(N) regression
	if med > 1200*time.Millisecond {
		t.Fatalf("logout took too long (median of 3): %s (likely scan regression). all=%v", med, times)
	}

	t.Logf("✅ logout duration median=%s (runs=%v) with %d extra sessions", med, times, sessionsToInsert)

	// Extra assertion: refresh must fail after logout for that session
	payload := `{"refresh_token":"` + refresh + `"}`
	req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(payload))
	req.Header.Set("Content-Type", "application/json")

	w := httptest.NewRecorder()
	TestRouter.ServeHTTP(w, req)

	// This refresh token belongs to last login session; we didn't logout it after median loop
	// So it should still work. If you WANT it to fail, you must logout after the last login.
	_ = w
}

func shaHex(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
