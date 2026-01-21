package db

import (
	"context"
	"log"
	"time"

	"github.com/ajiraj2411/wunderlist-backend/config"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

// ConnectMongo initializes and returns a MongoDB client
func ConnectMongo() *mongo.Client {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		config.AppConfig.MongoTimeout,
	)
	defer cancel()

	clientOpts := options.Client().
		ApplyURI(config.AppConfig.MongoURI).
		SetMaxPoolSize(50).
		SetMinPoolSize(5).
		SetConnectTimeout(config.AppConfig.MongoTimeout).
		SetServerSelectionTimeout(config.AppConfig.MongoTimeout)

	client, err := mongo.Connect(ctx, clientOpts)
	if err != nil {
		log.Fatal("❌ MongoDB connection failed:", err)
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		log.Fatal("❌ MongoDB ping failed:", err)
	}

	log.Println("✅ MongoDB connected successfully")
	return client
}

// DisconnectMongo gracefully closes MongoDB connection
func DisconnectMongo(client *mongo.Client) {
	if client == nil {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Disconnect(ctx); err != nil {
		log.Println("⚠️ MongoDB disconnect error:", err)
	} else {
		log.Println("🛑 MongoDB disconnected cleanly")
	}
}
