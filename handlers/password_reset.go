package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/internal/auth"
	"github.com/ajiraj2411/wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
)

type forgotPasswordReq struct {
	Email string `json:"email" validate:"required,email"`
}

func ForgotPassword(c *gin.Context) {
	var req forgotPasswordReq
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusOK, gin.H{"message": "if account exists, reset link sent"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var user models.User
	if err := UserCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user); err == nil {
		_, _ = auth.CreatePasswordResetToken(ctx, user.ID)
		// 🔔 email sending happens async (future)
	}

	c.JSON(http.StatusOK, gin.H{"message": "if account exists, reset link sent"})
}
