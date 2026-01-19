package handlers

import (
	"context"
	"net/http"
	"time"

	"wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// @Summary Get current user profile
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.MeResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /api/me [get]
func GetMe(c *gin.Context) {

	userIDHex := c.GetString(UserIDKey)
	if userIDHex == "" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "unauthorized"})
		return
	}

	uid, err := primitive.ObjectIDFromHex(userIDHex)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid user id"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	if err := UserCollection.FindOne(ctx, bson.M{"_id": uid}).Decode(&user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to fetch user"})
		return
	}

	c.JSON(http.StatusOK, models.MeResponse{
		ID:        user.ID.Hex(),
		Email:     user.Email,
		Role:      user.Role,
		CreatedAt: user.CreatedAt,
	})
}
