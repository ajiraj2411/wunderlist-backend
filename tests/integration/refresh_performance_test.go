package integration

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
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

func TestRefreshPerformance_WithManySessions(t *testing.T) {
	requirePerfEnabled(t)

	signupTestUser(t)

	// Insert many extra sessions (simulate prod DB size)
	const sessionsToInsert = 1500
	const batchSize = 200

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	userIDHex := getUserIDByEmail(t, "int@test.com")
	userID, _ := primitive.ObjectIDFromHex(userIDHex)

	// ✅ cleanup BEFORE login so we don't delete the refresh token session we need
	_, _ = TestDB.Collection("sessions").DeleteMany(ctx, bson.M{"user_id": userID})

	// login once to get a REAL refresh token that exists in DB
	_, refresh := loginAndGetTokens(t)
	if refresh == "" {
		t.Fatal("missing refresh token")
	}

	opts := options.InsertMany().SetOrdered(false)

	batch := make([]interface{}, 0, batchSize)
	now := time.Now().UTC()

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
			TokenSHA:  authTestSHA(raw),
			Role:      "user",
			UserAgent: "load-test",
			IPAddress: "127.0.0.1",
			CreatedAt: now,
			ExpiresAt: now.Add(7 * 24 * time.Hour),
		}

		batch = append(batch, s)

		// Insert in batches
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

	// Now refresh should STILL be fast (single indexed lookup)
	resetLimiters(t)

	// ✅ Measure refresh 3 times and take median (refresh rotates token!)
	times := make([]time.Duration, 0, 3)

	for i := 0; i < 3; i++ {
		payload := `{"refresh_token":"` + refresh + `"}`

		start := time.Now()

		req := httptest.NewRequest(http.MethodPost, "/refresh", bytes.NewBufferString(payload))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("User-Agent", "test-agent")
		req.Header.Set("X-Forwarded-For", "10.0.0.99")

		w := httptest.NewRecorder()
		TestRouter.ServeHTTP(w, req)

		elapsed := time.Since(start)
		times = append(times, elapsed)

		if w.Code != http.StatusOK {
			t.Fatalf("refresh failed (iteration %d): %d %s", i, w.Code, w.Body.String())
		}

		var res map[string]string
		_ = json.Unmarshal(w.Body.Bytes(), &res)

		if res["access_token"] == "" || res["refresh_token"] == "" {
			t.Fatalf("expected tokens in response, got: %s", w.Body.String())
		}

		// refresh rotates → must update token
		refresh = res["refresh_token"]
	}

	med := medianDuration3(times)

	// ✅ generous threshold so CI won't be flaky
	// This will still catch the major regression (cursor scan O(N)).
	if med > 1200*time.Millisecond {
		t.Fatalf(
			"refresh took too long (median of 3): %s (likely session scan regression). all=%v",
			med, times,
		)
	}

	t.Logf("✅ refresh duration median=%s (runs=%v) with %d extra sessions", med, times, sessionsToInsert)
}

// authTestSHA reproduces internal/auth refresh token SHA algorithm
func authTestSHA(raw string) string {
	h := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(h[:])
}

// medianDuration3 returns median of exactly 3 durations (stable, no sorting allocation)
func medianDuration3(ds []time.Duration) time.Duration {
	// ds length must be 3
	a, b, c := ds[0], ds[1], ds[2]

	if a > b {
		a, b = b, a
	}
	if b > c {
		b, c = c, b
	}
	if a > b {
		a, b = b, a
	}
	return b
}
