package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"

	"wunderlist-backend/db"
	"wunderlist-backend/handlers"
	"wunderlist-backend/internal/auth"
	"wunderlist-backend/middleware"
)

var (
	TestDB       *mongo.Database
	testlimiters auth.RateLimiterSet
	TestRedis    *redis.Client
)

func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)

	client := db.ConnectMongo()
	TestDB = client.Database("wunderlist_test")

	_ = TestDB.Drop(context.Background())

	// 🔑 AUTH INITIALIZATION (CRITICAL)
	auth.InitJWT(
		"test-secret",
		15*time.Minute,
		7*24*time.Hour,
	)

	// ✅ Start in-memory redis for blacklist tests
	mr, err := miniredis.Run()
	if err != nil {
		panic(err)
	}
	defer mr.Close()

	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	TestRedis = rdb

	// ✅ init blacklist system
	auth.InitJWTBlacklist(rdb)
	auth.InitUserRevoker(rdb)

	testlimiters = auth.NewRateLimiters(nil)
	handlers.InitAuthRateLimiters(testlimiters)

	auth.InitSessionStore(TestDB.Collection("sessions"))
	auth.InitUserStore(TestDB.Collection("users"))

	// 🔑 HANDLER COLLECTIONS
	handlers.SetUserCollection(TestDB.Collection("users"))
	handlers.SetSessionCollection(TestDB.Collection("sessions"))
	handlers.SetListCollection(TestDB.Collection("lists"))
	handlers.SetTaskCollection(TestDB.Collection("tasks"))

	db.EnsureIndexes(TestDB)

	TestRouter = gin.New()

	TestRouter.POST("/signup", handlers.Signup)
	TestRouter.POST("/login", handlers.Login)
	TestRouter.POST("/refresh", handlers.RefreshToken)
	TestRouter.POST("/auth/google", handlers.GoogleLogin)

	api := TestRouter.Group("/api")
	api.Use(middleware.AuthMiddleware())

	api.GET("/me", handlers.GetMe)
	api.DELETE("/sessions/current", handlers.LogoutCurrentSession)

	api.POST("/lists", handlers.CreateList)
	api.POST("/tasks", handlers.CreateTask)
	api.GET("/tasks", handlers.GetTasks)
	api.GET("/sessions", handlers.ListSessions)

	api.POST("/logout", handlers.Logout)
	api.POST("/logout/all", handlers.LogoutAll)

	admin := api.Group("/admin")
	admin.Use(middleware.RequireRole("admin"))

	admin.GET("/sessions", handlers.AdminListAllSessions)
	admin.GET("/users", handlers.AdminListUsers)
	admin.POST("/users/:userID/force-logout", handlers.AdminForceLogoutUser)

	code := m.Run()

	_ = TestDB.Drop(context.Background())
	os.Exit(code)
}
