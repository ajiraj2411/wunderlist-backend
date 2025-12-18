package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"wunderlist-backend/config"
	"wunderlist-backend/middleware"
	"wunderlist-backend/models"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
)

// =======================
// Signup
// =======================

// @Summary Signup a new user
// @Description Create a new user with email and password
// @Tags Auth
// @Accept json
// @Produce json
// @Param user body models.User true "User info"
// @Success 201 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 409 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /signup [post]
func Signup(c *gin.Context) {
	var user models.User
	if err := c.ShouldBindJSON(&user); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Failed to check email"})
		return
	}
	if count > 0 {
		c.JSON(http.StatusConflict, models.ErrorResponse{Error: "Email already exists"})
		return
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Password hashing failed"})
		return
	}

	user.ID = primitive.NewObjectID()
	user.Password = string(hashed)
	user.CreatedAt = time.Now().UTC()

	if _, err := UserCollection.InsertOne(ctx, user); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "User creation failed"})
		return
	}

	c.JSON(http.StatusCreated, models.MessageResponse{Message: "User created"})
}

// =======================
// Login
// =======================

// @Summary Login
// @Description Login with email & password
// @Tags Auth
// @Accept json
// @Produce json
// @Param credentials body map[string]string true "Email & password"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /login [post]
func Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email"`
		Password string `json:"password"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: err.Error()})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var user models.User
	if err := UserCollection.FindOne(ctx, bson.M{"email": req.Email}).Decode(&user); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid credentials"})
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.Password)); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid credentials"})
		return
	}

	accessToken, err := generateJWT(user.ID.Hex(), 15*time.Minute)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Access token generation failed"})
		return
	}

	refreshToken, err := generateJWT(user.ID.Hex(), 7*24*time.Hour)
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Refresh token generation failed"})
		return
	}

	session := models.Session{
		ID:        primitive.NewObjectID(),
		UserID:    user.ID,
		Token:     refreshToken,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
	}

	if _, err := SessionCollection.InsertOne(ctx, session); err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Session creation failed"})
		return
	}

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// =======================
// Refresh Token
// =======================

// @Summary Refresh token
// @Description Refresh access token
// @Tags Auth
// @Accept json
// @Produce json
// @Param token body map[string]string true "Refresh token"
// @Success 200 {object} models.AuthResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /refresh [post]
func RefreshToken(c *gin.Context) {
	var req struct {
		RefreshToken string `json:"refresh_token"`
	}

	if err := c.ShouldBindJSON(&req); err != nil || req.RefreshToken == "" {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Missing refresh_token"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	var session models.Session
	if err := SessionCollection.FindOne(ctx, bson.M{"token": req.RefreshToken}).Decode(&session); err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	if session.ExpiresAt.Before(time.Now()) {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Refresh token expired"})
		return
	}

	_, err := jwt.Parse(req.RefreshToken, func(t *jwt.Token) (interface{}, error) {
		if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, jwt.ErrTokenUnverifiable
		}
		return []byte(config.AppConfig.JWTSecret), nil
	})
	if err != nil {
		c.JSON(http.StatusUnauthorized, models.ErrorResponse{Error: "Invalid refresh token"})
		return
	}

	// Rotate session
	_, _ = SessionCollection.DeleteOne(ctx, bson.M{"_id": session.ID})

	accessToken, _ := generateJWT(session.UserID.Hex(), 15*time.Minute)
	refreshToken, _ := generateJWT(session.UserID.Hex(), 7*24*time.Hour)

	newSession := models.Session{
		ID:        primitive.NewObjectID(),
		UserID:    session.UserID,
		Token:     refreshToken,
		CreatedAt: time.Now().UTC(),
		ExpiresAt: time.Now().UTC().Add(7 * 24 * time.Hour),
	}
	_, _ = SessionCollection.InsertOne(ctx, newSession)

	c.JSON(http.StatusOK, models.AuthResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	})
}

// =======================
// Logout
// =======================

// @Summary Logout
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /logout [post]
func Logout(c *gin.Context) {
	auth := c.GetHeader("Authorization")
	if !strings.HasPrefix(auth, "Bearer ") {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Missing token"})
		return
	}

	token := strings.TrimPrefix(auth, "Bearer ")

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	_, err := SessionCollection.DeleteOne(ctx, bson.M{"token": token})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Logout failed"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{Message: "Logged out successfully"})
}

// =======================
// Logout All
// =======================

// @Summary Logout all sessions
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} models.MessageResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 401 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /logout/all [post]
func LogoutAll(c *gin.Context) {
	userID := c.GetString(middleware.UserIDKey)
	uid, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, models.ErrorResponse{Error: "Invalid user ID"})
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), dbTimeout)
	defer cancel()

	res, err := SessionCollection.DeleteMany(ctx, bson.M{"user_id": uid})
	if err != nil {
		c.JSON(http.StatusInternalServerError, models.ErrorResponse{Error: "Logout failed"})
		return
	}

	c.JSON(http.StatusOK, models.MessageResponse{
		Message: fmt.Sprintf("Logged out from all devices (%d sessions)", res.DeletedCount),
	})
}

// =======================
// JWT Helper
// =======================

func generateJWT(userID string, duration time.Duration) (string, error) {
	claims := jwt.MapClaims{
		"user_id": userID,
		"exp":     time.Now().Add(duration).Unix(),
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(config.AppConfig.JWTSecret))
}
