package integration

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/mongo"

	"wunderlist-backend/db"
	"wunderlist-backend/handlers"
	"wunderlist-backend/internal/auth"
	"wunderlist-backend/middleware"
)

var (
	TestDB *mongo.Database
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
	auth.InitSessionStore(TestDB.Collection("sessions"))

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

	api := TestRouter.Group("/api")
	api.Use(middleware.AuthMiddleware())
	api.POST("/lists", handlers.CreateList)

	code := m.Run()

	_ = TestDB.Drop(context.Background())
	os.Exit(code)
}
