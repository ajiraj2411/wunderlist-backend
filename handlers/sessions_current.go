package handlers

import (
	"context"
	"net/http"
	"time"

	"wunderlist-backend/internal/auth"
	"wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// @Summary Logout current session
// @Tags Sessions
// @Security BearerAuth
// @Success 204
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/sessions/current [delete]
func LogoutCurrentSession(c *gin.Context) {

	sessionIDHex := c.GetString("sessionID")
	if sessionIDHex == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "missing session id"})
		return
	}

	sid, err := primitive.ObjectIDFromHex(sessionIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid session id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err = auth.DeleteSessionByIDOnly(ctx, sid)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to revoke session"})
		return
	}

	// revoke current access token jti immediately
	auth.RevokeJWT(sessionIDHex, auth.AccessTokenExpiry())

	c.Status(http.StatusNoContent)
}
