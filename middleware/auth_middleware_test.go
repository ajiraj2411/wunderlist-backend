package middleware_test

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wunderlist-backend/internal/auth"
	"wunderlist-backend/middleware"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/require"
)

func setupRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.Use(middleware.AuthMiddleware())

	// Dummy protected route
	r.GET("/protected", func(c *gin.Context) {
		userID, exists := c.Get(middleware.UserIDKey)
		if !exists {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "userID missing"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"user_id": userID})
	})

	return r
}

func TestAuthMiddleware_NoToken(t *testing.T) {
	auth.InitJWT("test-secret", 15*time.Minute, 7*24*time.Hour)

	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestAuthMiddleware_InvalidToken(t *testing.T) {
	auth.InitJWT("test-secret", 15*time.Minute, 7*24*time.Hour)

	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.value")
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestAuthMiddleware_ExpiredToken(t *testing.T) {
	auth.InitJWT("test-secret", -1*time.Minute, 7*24*time.Hour)

	token, err := auth.GenerateAccessToken("user123", "user")
	require.NoError(t, err)

	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusUnauthorized, res.Code)
}

func TestAuthMiddleware_ValidToken(t *testing.T) {
	auth.InitJWT("test-secret", 15*time.Minute, 7*24*time.Hour)

	token, err := auth.GenerateAccessToken("user123", "user")
	require.NoError(t, err)

	router := setupRouter()

	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	res := httptest.NewRecorder()

	router.ServeHTTP(res, req)

	require.Equal(t, http.StatusOK, res.Code)
	require.Contains(t, res.Body.String(), "user123")
}
