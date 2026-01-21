package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/ajiraj2411/wunderlist-backend/config"
	"github.com/ajiraj2411/wunderlist-backend/db"
	"github.com/ajiraj2411/wunderlist-backend/handlers"
	"github.com/ajiraj2411/wunderlist-backend/internal/auth"
	"github.com/ajiraj2411/wunderlist-backend/middleware"

	_ "github.com/ajiraj2411/wunderlist-backend/docs"
)

// @title Wunderlist Backend API
// @version 1.0
// @description Wunderlist backend API
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {

	// ---------------------------
	// Validate config & env
	// ---------------------------
	config.Validate()

	if config.AppConfig.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// ---------------------------
	// MongoDB
	// ---------------------------
	client := db.ConnectMongo()
	database := client.Database(config.AppConfig.MongoDBName)

	auth.InitJWT(
		config.AppConfig.JWTSecret,
		config.AppConfig.AccessTokenTTL,
		config.AppConfig.RefreshTokenTTL,
	)
	rdb := auth.NewRedisClient()
	auth.InitJWTBlacklist(rdb)
	auth.InitUserRevoker(rdb)

	limiters := auth.NewRateLimiters(rdb)
	handlers.InitAuthRateLimiters(limiters)

	auth.InitSessionStore(database.Collection("sessions"))
	auth.InitUserStore(database.Collection("users"))
	auth.InitPasswordResetStore(database.Collection("password_reset_tokens"))

	handlers.SetUserCollection(database.Collection("users"))
	handlers.SetSessionCollection(database.Collection("sessions"))
	handlers.SetListCollection(database.Collection("lists"))
	handlers.SetTaskCollection(database.Collection("tasks"))

	db.EnsureIndexes(database)

	// ---------------------------
	// Gin router
	// ---------------------------
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.RequestLogger())
	r.Use(middleware.SecurityHeaders())

	// ---------------------------
	// Public routes
	// ---------------------------
	r.POST("/signup", handlers.Signup)
	r.POST("/login", handlers.Login)
	r.POST("/refresh", handlers.RefreshToken)
	r.POST("/auth/google", handlers.GoogleLogin)

	// Test account reset (only in non-production)
	r.POST("/debug/reset-test-account", handlers.ResetTestAccount)

	// ---------------------------
	// Protected routes
	// ---------------------------
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware())

	api.GET("/me", handlers.GetMe)

	// Lists
	api.POST("/lists", handlers.CreateList)
	api.GET("/lists", handlers.GetLists)
	api.PUT("/lists/:id", handlers.UpdateList)
	api.DELETE("/lists/:id", handlers.DeleteList)

	// Tasks
	api.POST("/tasks", handlers.CreateTask)
	api.GET("/tasks", handlers.GetTasks)
	api.GET("/tasks/active", handlers.GetActiveTasks)
	api.GET("/tasks/search", handlers.SearchTasks)
	api.PUT("/tasks/:id", handlers.UpdateTask)
	api.DELETE("/tasks/:id", handlers.DeleteTask)

	// Sessions
	api.GET("/sessions", handlers.ListSessions)
	api.DELETE("/sessions/:id", handlers.RevokeSession)
	api.DELETE("/sessions/current", handlers.LogoutCurrentSession)

	// Auth session control
	api.POST("/logout", handlers.Logout)
	api.POST("/logout/all", handlers.LogoutAll)

	admin := api.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))

	admin.GET("/users", handlers.AdminListUsers)
	admin.GET("/sessions", handlers.AdminListAllSessions)
	admin.POST("/users/:userID/force-logout", handlers.AdminForceLogoutUser)

	auth := r.Group("/auth")
	auth.POST("/forgot-password", handlers.ForgotPassword)
	auth.POST("/reset-password", handlers.ResetPassword)

	// ---------------------------
	// Health & readiness
	// ---------------------------
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	r.GET("/ready", func(c *gin.Context) {
		if err := client.Ping(context.Background(), nil); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{"status": "not ready"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"status": "ready"})
	})

	// ---------------------------
	// Swagger
	// ---------------------------
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// ---------------------------
	// HTTP server (graceful shutdown)
	// ---------------------------
	srv := &http.Server{
		Addr:    ":" + config.AppConfig.Port,
		Handler: r,
	}

	go func() {
		log.Printf("🚀 Server running on http://localhost:%s", config.AppConfig.Port)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	// Wait for shutdown signal
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
	<-quit

	log.Println("🛑 Shutting down server...")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		log.Printf("server shutdown failed: %v", err)
	}

	db.DisconnectMongo(client)
	log.Println("✅ Server exited cleanly")
}
