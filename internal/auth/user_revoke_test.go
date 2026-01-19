package auth

import (
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func TestGetUserRevokedAt_ParsesLegacySeconds(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis start failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	InitUserRevoker(rdb)

	userID := "user-seconds"
	key := "jwt:user_revoked_at:" + userID

	// legacy seconds value (Jan 1, 2025 UTC approx)
	sec := int64(1735689600)
	mr.Set(key, "1735689600")

	got, ok := GetUserRevokedAt(userID)
	if !ok {
		t.Fatal("expected ok=true")
	}

	if got.Unix() != sec {
		t.Fatalf("expected unix seconds=%d, got=%d", sec, got.Unix())
	}
}

func TestGetUserRevokedAt_ParsesMillis(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis start failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	InitUserRevoker(rdb)

	userID := "user-millis"
	key := "jwt:user_revoked_at:" + userID

	// millis timestamp
	ms := int64(1735689600123)
	mr.Set(key, "1735689600123")

	got, ok := GetUserRevokedAt(userID)
	if !ok {
		t.Fatal("expected ok=true")
	}

	if got.UnixMilli() != ms {
		t.Fatalf("expected unix millis=%d, got=%d", ms, got.UnixMilli())
	}
}

func TestSetUserRevokedNow_StoresMillis(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis start failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	InitUserRevoker(rdb)

	userID := "user-set"
	key := "jwt:user_revoked_at:" + userID

	before := time.Now().UTC().UnixMilli()

	err = SetUserRevokedNow(userID, 10*time.Minute)
	if err != nil {
		t.Fatalf("SetUserRevokedNow failed: %v", err)
	}

	raw, err := mr.Get(key)
	require.NoError(t, err)
	if raw == "" {
		t.Fatalf("expected redis key to be set: %s", key)
	}

	revokedAt, ok := GetUserRevokedAt(userID)
	if !ok {
		t.Fatal("expected ok=true from GetUserRevokedAt")
	}

	after := time.Now().UTC().UnixMilli()
	got := revokedAt.UnixMilli()

	// revokedAt should be within [before, after]
	if got < before || got > after {
		t.Fatalf("revokedAt millis out of range: got=%d before=%d after=%d", got, before, after)
	}

	// TTL should exist (miniredis returns -1 if no TTL)
	ttl := mr.TTL(key)
	if ttl <= 0 {
		t.Fatalf("expected ttl to be set, got=%v", ttl)
	}
}

func TestGetUserRevokedAt_InvalidValue(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis start failed: %v", err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	InitUserRevoker(rdb)

	userID := "user-invalid"
	key := "jwt:user_revoked_at:" + userID

	mr.Set(key, "not-a-number")

	_, ok := GetUserRevokedAt(userID)
	if ok {
		t.Fatal("expected ok=false for invalid value")
	}
}
