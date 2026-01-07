package middleware

import (
	"net/http"
	"strings"

	"wunderlist-backend/internal/auth"
	"wunderlist-backend/internal/models"

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

		userID, role, sessionID, err := auth.ValidateAccessToken(token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, models.ErrorResponse{
				Error: "invalid or expired token",
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
