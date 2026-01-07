package handlers

import (
	"context"
	"net/http"
	"time"

	"wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// @Summary List all active sessions (admin only)
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.Session
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/sessions [get]
func AdminListAllSessions(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := SessionCollection.Find(ctx, bson.M{
		"expires_at": bson.M{"$gt": time.Now().UTC()},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to fetch sessions",
		})
		return
	}
	defer cursor.Close(ctx)

	var sessions []models.Session
	if err := cursor.All(ctx, &sessions); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to parse sessions",
		})
		return
	}

	c.JSON(http.StatusOK, sessions)
}
