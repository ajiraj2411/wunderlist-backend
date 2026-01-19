// internal/auth/user_revoke.go
package auth

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/redis/go-redis/v9"
)

type UserRevoker struct {
	rdb *redis.Client
}

var userRevoker *UserRevoker

func InitUserRevoker(rdb *redis.Client) {
	if rdb == nil {
		userRevoker = nil
		return
	}
	userRevoker = &UserRevoker{rdb: rdb}
}

func (u *UserRevoker) key(userID string) string {
	return fmt.Sprintf("jwt:user_revoked_at:%s", userID)
}

// SetUserRevokedNow stores revoked_at in UnixMillis for precision.
func SetUserRevokedNow(userID string, keep time.Duration) error {
	if userRevoker == nil {
		return nil // fail-open
	}

	nowMs := time.Now().UTC().UnixMilli()

	ctx := context.Background()
	key := userRevoker.key(userID)

	// store integer millis as string
	return userRevoker.rdb.Set(ctx, key, strconv.FormatInt(nowMs, 10), keep).Err()
}

// GetUserRevokedAt loads revoked_at. Supports millis + legacy seconds format.
func GetUserRevokedAt(userID string) (time.Time, bool) {
	if userRevoker == nil {
		return time.Time{}, false
	}

	ctx := context.Background()
	key := userRevoker.key(userID)

	val, err := userRevoker.rdb.Get(ctx, key).Result()
	if err != nil {
		return time.Time{}, false
	}

	n, err := strconv.ParseInt(val, 10, 64)
	if err != nil || n <= 0 {
		return time.Time{}, false
	}

	// ✅ Robust parsing:
	// - legacy seconds are ~1_700_000_000
	// - millis are ~1_700_000_000_000
	// Threshold: treat < 1e11 as seconds (covers up to year 5138)
	if n < 100_000_000_000 {
		return time.Unix(n, 0).UTC(), true
	}

	// millis
	return time.UnixMilli(n).UTC(), true
}
