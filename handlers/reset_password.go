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

type resetPasswordReq struct {
	Token       string `json:"token" validate:"required"`
	NewPassword string `json:"new_password" validate:"required,min=8"`
}

func ResetPassword(c *gin.Context) {
	var req resetPasswordReq
	if c.ShouldBindJSON(&req) != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid payload"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	token, err := auth.ConsumePasswordResetToken(ctx, req.Token)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid or expired token"})
		return
	}

	hash, err := auth.HashPassword(req.NewPassword)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to reset"})
		return
	}

	_, err = UserCollection.UpdateByID(ctx, token.UserID, bson.M{
		"$set": bson.M{"password": hash},
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "failed to reset"})
		return
	}

	// 🔥 revoke ALL sessions
	_ = auth.SetUserRevokedNow(token.UserID.Hex(), 30*24*time.Hour)

	c.JSON(http.StatusOK, gin.H{"message": "password reset successful"})
}
