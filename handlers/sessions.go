package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/auth"
	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// @Summary List active sessions
// @Tags Sessions
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.SessionResponse
// @Router /api/sessions [get]
func ListSessions(c *gin.Context) {

	userIDHex := c.GetString("userID")
	userID, _ := primitive.ObjectIDFromHex(userIDHex)

	currentSessionIDHex := c.GetString("sessionID")
	var currentSessionID primitive.ObjectID

	if currentSessionIDHex != "" {
		currentSessionID, _ = primitive.ObjectIDFromHex(currentSessionIDHex)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	sessions, err := auth.ListUserSessions(ctx, userID, currentSessionID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to list sessions",
		})
		return
	}

	c.JSON(http.StatusOK, sessions)
}

// @Summary Revoke a session
// @Tags Sessions
// @Security BearerAuth
// @Param id path string true "Session ID"
// @Success 204
// @Router /api/sessions/{id} [delete]
func RevokeSession(c *gin.Context) {

	userIDHex := c.GetString("userID")
	userID, _ := primitive.ObjectIDFromHex(userIDHex)

	sessionID, err := primitive.ObjectIDFromHex(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid session id",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := auth.DeleteSessionByID(ctx, sessionID, userID)
	if err != nil || res == 0 {
		c.JSON(http.StatusNotFound, models.ErrorResponse{
			Error: "session not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}
