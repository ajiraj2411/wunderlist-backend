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
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid payload"})
		return
	}

	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: validationError(err)})
		return
	}

	ip := auth.ExtractClientIP(c.Request)
	if googleLimiter != nil {
		key := "google:" + ip
		if !googleLimiter.Allow(key) {
			c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
				Error: "too many google login attempts",
			})
			return
		}
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	googleUser, err := auth.VerifyGoogleIDTokenWrapped(ctx, req.IDToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid google token"})
		return
	}

	// ✅ must be verified
	if !googleUser.EmailVerified {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "google email not verified"})
		return
	}

	user, err := auth.FindOrCreateGoogleUser(
		ctx,
		googleUser.Email,
		googleUser.Name,
		googleUser.GoogleID,
		googleUser.EmailVerified,
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: err.Error()})
		return
	}

	access, refresh, err := auth.Login(
		ctx,
		user,
		"", // passwordless
		auth.ExtractUserAgent(c.Request),
		auth.ExtractClientIP(c.Request),
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "google login failed"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    auth.AccessTokenExpiry(),
	})
}
