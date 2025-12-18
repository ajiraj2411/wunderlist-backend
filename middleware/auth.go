package middleware

import (
	"net/http"
	"strings"
	"time"

	"wunderlist-backend/config"
	"wunderlist-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

const UserIDKey = "userID"

// MyClaims defines custom JWT claims
type MyClaims struct {
	UserID string `json:"user_id"`
	jwt.RegisteredClaims
}

// AuthMiddleware validates JWT and sets userID in context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Missing or invalid Authorization header",
			})
			c.Abort()
			return
		}

		tokenStr := strings.TrimPrefix(authHeader, "Bearer ")

		claims := &MyClaims{}
		token, err := jwt.ParseWithClaims(tokenStr, claims, func(t *jwt.Token) (interface{}, error) {
			// ✅ Proper signing method validation
			if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrTokenUnverifiable
			}
			return []byte(config.AppConfig.JWTSecret), nil
		})

		if err != nil || !token.Valid {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid or expired token",
			})
			c.Abort()
			return
		}

		// ✅ Defensive expiration check
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now().UTC()) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Token expired",
			})
			c.Abort()
			return
		}

		if claims.UserID == "" {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "Invalid token payload",
			})
			c.Abort()
			return
		}

		// ✅ Store user ID in context
		c.Set(UserIDKey, claims.UserID)
		c.Next()
	}
}
