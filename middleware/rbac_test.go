package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"wunderlist-backend/internal/auth"

	"github.com/gin-gonic/gin"
)

func makeValidToken(t *testing.T, role string) string {
	t.Helper()

	auth.InitJWT("test-secret", time.Minute, time.Hour)

	token, err := auth.GenerateAccessToken("test-user-id", role)
	if err != nil {
		t.Fatal(err)
	}

	return token
}

func TestRequireRole_AdminAllowed(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/admin",
		AuthMiddleware(),
		RequireRole("admin"),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+makeValidToken(t, "admin"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}
}

func TestRequireRole_UserDenied(t *testing.T) {
	gin.SetMode(gin.TestMode)

	r := gin.New()
	r.GET("/admin",
		AuthMiddleware(),
		RequireRole("admin"),
		func(c *gin.Context) {
			c.Status(http.StatusOK)
		},
	)

	req := httptest.NewRequest("GET", "/admin", nil)
	req.Header.Set("Authorization", "Bearer "+makeValidToken(t, "user"))

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d", w.Code)
	}
}
