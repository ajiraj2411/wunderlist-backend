package middleware

import (
	"net/http"
	"strings"

	"github.com/ajiraj2411/wunderlist-backend/internal/auth"
	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
)

const (
	UserIDKey    = "userID"
	RoleKey      = "role"
	SessionIDKey = "sessionID"
)

// AuthMiddleware validates access JWT and injects userID into context
func AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {

		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "missing authorization token",
			})
			c.Abort()
			return
		}

		token := strings.TrimPrefix(header, "Bearer ")

		userID, role, sessionID, issuedAt, err := auth.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid or expired token",
			})
			c.Abort()
			return
		}

		// ✅ USER-WIDE REVOKE CHECK (Logout all sessions / Admin force logout)
		if revokedAt, ok := auth.GetUserRevokedAt(userID); ok {
			// block if issuedAt <= revokedAt
			if !issuedAt.After(revokedAt) {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error: "token revoked",
				})
				c.Abort()
				return
			}
		}

		// 🔥 BLACKLIST CHECK
		if auth.IsJWTRevoked(sessionID) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "token revoked",
			})
			c.Abort()
			return
		}

		// store user id in context
		c.Set(UserIDKey, userID)
		c.Set(RoleKey, role)
		c.Set(SessionIDKey, sessionID)
		c.Next()
	}
}
