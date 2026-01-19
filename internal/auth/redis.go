package auth

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/redis/go-redis/v9"
)

func NewRedisClient() *redis.Client {
	addr := os.Getenv("REDIS_ADDR")
	if addr == "" {
		log.Println("REDIS_ADDR not set — Redis disabled")
		return nil
	}

	rdb := redis.NewClient(&redis.Options{
		Addr:         addr,
		Password:     os.Getenv("REDIS_PASSWORD"),
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		log.Printf("⚠️ Redis unavailable (%v) — falling back to memory", err)
		return nil
	}

	log.Println("✅ Redis connected")
	return rdb
}
