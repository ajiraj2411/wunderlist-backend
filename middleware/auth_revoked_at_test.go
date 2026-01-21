package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/auth"
	"github.com/ajiraj2411/wunderlist-backend/middleware"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/require"
)

func setupRevokedAtRouter(t *testing.T) *gin.Engine {
	t.Helper()
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.AuthMiddleware())

	r.GET("/protected", func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	return r
}

func TestAuthMiddleware_UserRevokedAt_IssuedBeforeBlocked(t *testing.T) {
	// init jwt
	auth.InitJWT("test-secret", 15*time.Minute, 7*24*time.Hour)

	// init redis revoker
	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	auth.InitUserRevoker(rdb)

	userID := "user123"
	revokedAt := time.Now().UTC().Add(-1 * time.Minute)

	// set revokedAt in redis (millis)
	key := "jwt:user_revoked_at:" + userID
	mr.Set(key, formatMillis(revokedAt))

	issuedAt := revokedAt.Add(-2 * time.Minute) // issued BEFORE revoked => must block
	exp := time.Now().UTC().Add(10 * time.Minute)

	token, err := auth.GenerateJWTWithTimes(userID, "user", issuedAt, exp)
	require.NoError(t, err)

	router := setupRevokedAtRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_UserRevokedAt_IssuedEqualBlocked(t *testing.T) {
	auth.InitJWT("test-secret", 15*time.Minute, 7*24*time.Hour)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	auth.InitUserRevoker(rdb)

	userID := "user123"
	revokedAt := time.Now().UTC()

	key := "jwt:user_revoked_at:" + userID
	mr.Set(key, formatMillis(revokedAt))

	// EXACT equality case
	issuedAt := revokedAt
	exp := time.Now().UTC().Add(10 * time.Minute)

	token, err := auth.GenerateJWTWithTimes(userID, "user", issuedAt, exp)
	require.NoError(t, err)

	router := setupRevokedAtRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAuthMiddleware_UserRevokedAt_IssuedAfterAllowed(t *testing.T) {
	auth.InitJWT("test-secret", 15*time.Minute, 7*24*time.Hour)

	mr, err := miniredis.Run()
	require.NoError(t, err)
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	auth.InitUserRevoker(rdb)

	userID := "user123"
	revokedAt := time.Now().UTC()

	key := "jwt:user_revoked_at:" + userID
	mr.Set(key, formatMillis(revokedAt))

	issuedAt := revokedAt.Add(1 * time.Second) // issued AFTER revoked => allowed
	exp := time.Now().UTC().Add(10 * time.Minute)

	token, err := auth.GenerateJWTWithTimes(userID, "user", issuedAt, exp)
	require.NoError(t, err)

	router := setupRevokedAtRouter(t)

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)
}

// ----- helpers -----

func formatMillis(t time.Time) string {
	return strconv.FormatInt(t.UTC().UnixMilli(), 10)
}
