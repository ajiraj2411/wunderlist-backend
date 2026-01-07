package handlers

import (
	"context"
	"net/http"
	"time"

	"wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

// @Summary List all users (admin only)
// @Tags Admin
// @Security BearerAuth
// @Produce json
// @Success 200 {array} models.User
// @Failure 403 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/admin/users [get]
func AdminListUsers(c *gin.Context) {

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	cursor, err := UserCollection.Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to fetch users",
		})
		return
	}
	defer cursor.Close(ctx)

	var users []models.User
	if err := cursor.All(ctx, &users); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "failed to parse users",
		})
		return
	}

	c.JSON(http.StatusOK, users)
}
