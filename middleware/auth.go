package middleware

import (
	"net/http"
	"strings"

	"github.com/ajiraj2411/wunderlist-backend/internal/auth"
	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
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

		// ✅ USER-WIDE REVOKE CHECK
		if revokedAt, ok := auth.GetUserRevokedAt(userID); ok {
			if !issuedAt.After(revokedAt) {
				c.JSON(http.StatusUnauthorized, models.ErrorResponse{
					Error: "token revoked",
				})
				c.Abort()
				return
			}
		}

		// 🔥 JWT BLACKLIST CHECK
		if auth.IsJWTRevoked(sessionID) {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "token revoked",
			})
			c.Abort()
			return
		}

		// 🔒 SESSION EXISTENCE CHECK (Logout current session hard revoke)
		sid, err := primitive.ObjectIDFromHex(sessionID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid session",
			})
			c.Abort()
			return
		}

		uid, err := primitive.ObjectIDFromHex(userID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid user",
			})
			c.Abort()
			return
		}

		exists, err := auth.SessionExists(c.Request.Context(), sid, uid)
		if err != nil || !exists {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid or revoked session",
			})
			c.Abort()
			return
		}

		// store values in context
		c.Set(UserIDKey, userID)
		c.Set(RoleKey, role)
		c.Set(SessionIDKey, sessionID)
		c.Next()
	}
}
