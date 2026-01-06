package handlers

import (
	"context"
	"net/http"
	"time"

	"wunderlist-backend/internal/auth"
	"wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// @Summary Reset test account
// @Tags Debug
// @Router /debug/reset-test-account [post]
func ResetTestAccount(c *gin.Context) {

	email := "testuser@example.com"
	rawPass := "password123"

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// remove user + sessions
	_, _ = UserCollection.DeleteOne(ctx, bson.M{"email": email})
	_, _ = SessionCollection.DeleteMany(ctx, bson.M{})

	hash, err := auth.HashPassword(rawPass)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "password hash failed",
		})
		return
	}

	user := models.User{
		ID:        primitive.NewObjectID(),
		Email:     email,
		Password:  hash, // ✅ GUARANTEED bcrypt hash
		CreatedAt: time.Now().UTC(),
	}

	if _, err := UserCollection.InsertOne(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "insert failed",
		})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "reset OK",
	})
}
