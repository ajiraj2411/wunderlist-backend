package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/auth"
	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
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
	userIDHex := c.GetString("userID")
	sessionIDHex := c.GetString("sessionID")
	if userIDHex == "" || sessionIDHex == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "missing session"})
		return
	}

	uid, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid user"})
		return
	}

	sid, err := primitive.ObjectIDFromHex(sessionIDHex)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid session id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// delete refresh session (O(1))
	_, _ = SessionCollection.DeleteOne(ctx, bson.M{
		"_id":     sid,
		"user_id": uid,
	})

	// revoke current access token immediately (JWT jti == sessionID)
	auth.RevokeJWT(sessionIDHex, auth.AccessTokenExpiry())

	c.Status(http.StatusNoContent)

}
