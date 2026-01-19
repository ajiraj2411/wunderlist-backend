package handlers

import (
	"context"
	"net/http"
	"strings"
	"time"

	"wunderlist-backend/internal/auth"
	"wunderlist-backend/internal/models"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var (
	loginLimiter   auth.RateLimiter
	refreshLimiter auth.RateLimiter
	googleLimiter  auth.RateLimiter
)

func InitAuthRateLimiters(limiters auth.RateLimiterSet) {
	loginLimiter = limiters.Login
	refreshLimiter = limiters.Refresh
	googleLimiter = limiters.Google
}

type signupRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8,max=64"`
}

type loginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type refreshRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

/* =====================================================
   Signup
===================================================== */

// @Summary Signup
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body map[string]string true "Signup payload"
// @Success 201 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Router /signup [post]
func Signup(c *gin.Context) {

	var req signupRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid payload"})
		return
	}

	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: validationError(err)})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	count, _ := UserCollection.CountDocuments(ctx, bson.M{"email": req.Email})
	if count > 0 {
		c.JSON(http.StatusConflict, models.ErrorResponse{
			Error: "email already exists",
		})
		return
	}

	hash, err := auth.HashPassword(req.Password)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "password hashing failed",
		})
		return
	}

	user := models.User{
		ID:           primitive.NewObjectID(),
		Email:        req.Email,
		Password:     hash,
		Role:         "user",
		AuthProvider: "password",
		CreatedAt:    time.Now().UTC(),
	}

	if _, err := UserCollection.InsertOne(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{
			Error: "user creation failed",
		})
		return
	}

	c.JSON(http.StatusCreated, models.MessageResponse{
		Message: "user created",
	})
}

/* =====================================================
   Login
===================================================== */

// @Summary Login
// @Description Login with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body map[string]string true "Email & Password"
// @Success 200 {object} models.AuthResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /login [post]
func Login(c *gin.Context) {

	var req loginRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid payload"})
		return
	}

	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: validationError(err)})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	ip := auth.ExtractClientIP(c.Request)
	key := "login:" + ip

	if !loginLimiter.Allow(key) {
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
			Error: "too many login attempts, try again later",
		})
		return
	}

	var user models.User
	if err := UserCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials"})
		return
	}

	if user.Password == "" && user.AuthProvider == "google" {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "use google login"})
		return
	}

	access, refresh, err := auth.Login(
		ctx,
		&user,
		strings.TrimSpace(req.Password),
		auth.ExtractUserAgent(c.Request),
		auth.ExtractClientIP(c.Request),
	)

	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid credentials"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    auth.AccessTokenExpiry(),
	})
}

/* =====================================================
   Refresh Token
===================================================== */

// @Summary Refresh token
// @Description Rotate refresh token and issue new access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param token body map[string]string true "Refresh token"
// @Success 200 {object} models.AuthResponse
// @Failure 401 {object} models.ErrorResponse
// @Router /refresh [post]
func RefreshToken(c *gin.Context) {
	var req refreshRequest

	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "invalid payload"})
		return
	}

	if err := validate.Struct(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: validationError(err)})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	ip := auth.ExtractClientIP(c.Request)
	key := "refresh:" + ip

	if !refreshLimiter.Allow(key) {
		c.JSON(http.StatusTooManyRequests, models.ErrorResponse{
			Error: "too many refresh attempts",
		})
		return
	}

	session, err := auth.ValidateRefreshTokenStrict(ctx, req.RefreshToken, auth.ExtractUserAgent(c.Request),
		auth.ExtractClientIP(c.Request),
	)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: err.Error()})
		return
	}

	access, refresh, err := auth.RotateRefreshToken(ctx, session)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "token rotation failed"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  access,
		RefreshToken: refresh,
		ExpiresAt:    auth.AccessTokenExpiry(),
	})
}

/* =====================================================
   Logout (single session)
===================================================== */

// @Summary Logout
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.MessageResponse
// @Router /logout [post]
func Logout(c *gin.Context) {

	authHeader := c.GetHeader("Authorization")
	if !strings.HasPrefix(authHeader, "Bearer ") {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "missing token"})
		return
	}

	token := strings.TrimPrefix(authHeader, "Bearer ")

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if err := auth.DeleteSessionByToken(ctx, token); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "logout failed"})
		return
	}

	sessionID := c.GetString("sessionID")

	auth.RevokeJWT(
		sessionID,
		auth.AccessTokenExpiry(),
	)

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "logged out",
	})
}

/* =====================================================
   Logout all sessions
===================================================== */

// @Summary Logout all sessions
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.MessageResponse
// @Router /logout/all [post]
func LogoutAll(c *gin.Context) {

	userID := c.GetString("userID")
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "invalid user"})
		return
	}

	// ✅ revoke ALL access tokens for this user (instant)
	_ = auth.SetUserRevokedNow(uid.Hex(), 30*24*time.Hour)

	// Optional: also revoke current jti (not required anymore, but harmless)
	sessionID := c.GetString("sessionID")
	if sessionID != "" {
		auth.RevokeJWT(
			sessionID,
			auth.AccessTokenExpiry(),
		)
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	if err := auth.DeleteAllSessions(ctx, uid); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "logout failed"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: "logged out from all devices",
	})
}
