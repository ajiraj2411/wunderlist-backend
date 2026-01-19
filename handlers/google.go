package handlers

import (
	"context"
	"net/http"
	"time"

	"wunderlist-backend/internal/auth"
	"wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
)

type googleLoginRequest struct {
	IDToken string `json:"id_token" validate:"required"`
}

// @Summary Google OAuth login
// @Tags Auth
// @Accept json
// @Produce json
// @Param body body googleLoginRequest true "Google ID token"
// @Success 200 {object} models.AuthResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /auth/google [post]
func GoogleLogin(c *gin.Context) {

	var req googleLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error: "invalid payload",
		})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 🔐 verify google token
	googleUser, err := auth.VerifyGoogleIDToken(ctx, req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{
			Error: "invalid google token",
		})
		return
	}

	// 👤 find or create user
	user, err := auth.FindOrCreateGoogleUser(
		ctx,
		googleUser.Email,
		googleUser.Name,
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "user creation failed",
		})
		return
	}

	// 🔑 issue tokens + session
	access, refresh, err := auth.Login(
		ctx,
		user,
		"", // no password
		auth.ExtractUserAgent(c.Request),
		auth.ExtractClientIP(c.Request),
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "login failed",
		})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    auth.AccessTokenExpiry(),
	})
}
