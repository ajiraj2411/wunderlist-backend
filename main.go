package main

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"wunderlist-backend/config"
	"wunderlist-backend/db"
	_ "wunderlist-backend/docs"
	"wunderlist-backend/handlers"
	"wunderlist-backend/middleware"
)

// @title Wunderlist Backend API
// @version 1.0
// @description Wunderlist (Microsoft To Do) clone backend
// @host localhost:8080
// @BasePath /
// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
func main() {
	// -------------------------------
	// Connect to MongoDB
	// -------------------------------
	client := db.ConnectMongo()
	database := client.Database(config.AppConfig.MongoDBName)

	// -------------------------------
	// Inject collections into handlers
	// -------------------------------
	handlers.SetUserCollection(database.Collection("users"))
	handlers.SetSessionCollection(database.Collection("sessions"))
	handlers.SetListCollection(database.Collection("lists"))
	handlers.SetTaskCollection(database.Collection("tasks"))

	// -------------------------------
	// Ensure Indexes
	// -------------------------------
	db.EnsureIndexes(database)

	// -------------------------------
	// Setup Gin
	// -------------------------------
	r := gin.Default()

	// -------------------------------
	// Public routes
	// -------------------------------
	r.POST("/signup", handlers.Signup)
	r.POST("/login", handlers.Login)
	r.POST("/google-login", handlers.GoogleLogin) // Stub, can implement later

	// -------------------------------
	// Protected routes
	// -------------------------------
	api := r.Group("/api")
	api.Use(middleware.AuthMiddleware()) // Checks JWT and sets UserID

	// Lists
	api.POST("/lists", handlers.CreateList)
	api.GET("/lists", handlers.GetLists)
	api.PUT("/lists/:id", handlers.UpdateList)
	api.DELETE("/lists/:id", handlers.DeleteList)

	// Tasks
	api.POST("/tasks", handlers.CreateTask)
	api.GET("/tasks", handlers.GetTasks)              // Uses compound index
	api.GET("/tasks/active", handlers.GetActiveTasks) // Uses partial index
	api.GET("/tasks/search", handlers.SearchTasks)    // Uses text index with relevance
	api.PUT("/tasks/:id", handlers.UpdateTask)
	api.DELETE("/tasks/:id", handlers.DeleteTask)

	// Logout
	api.POST("/logout", handlers.Logout)
	api.POST("/logout/all", handlers.LogoutAll)

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// -------------------------------
	// Start server
	// -------------------------------
	log.Printf("🚀 Server running on http://localhost:%s", config.AppConfig.Port)
	r.Run(":" + config.AppConfig.Port)
}
