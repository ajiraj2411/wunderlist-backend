package config

import (
	"os"
	"strconv"
	"time"
)

type appConfig struct {
	Port             string
	MongoURI         string
	MongoDBName      string
	MongoTimeout     time.Duration
	JWTSecret        string
	JWTRefreshSecret string
	AccessTokenTTL   time.Duration
	RefreshTokenTTL  time.Duration
	Environment      string
}

var AppConfig = appConfig{
	Port:             getEnv("PORT", "8080"),
	MongoURI:         getEnv("MONGO_URI", "mongodb://localhost:27017"),
	MongoDBName:      getEnv("MONGO_DB", "wunderlist"),
	MongoTimeout:     getEnvAsDuration("MONGO_TIMEOUT_SEC", 10),
	JWTSecret:        getEnv("JWT_SECRET", "secret123"),
	JWTRefreshSecret: getEnv("JWT_REFRESH_SECRET", "refresh_secret_123"),
	AccessTokenTTL:   getEnvAsDuration("JWT_ACCESS_TTL_MIN", 15),
	RefreshTokenTTL:  getEnvAsDuration("JWT_REFRESH_TTL_HOURS", 168), // 7 days
	Environment:      getEnv("APP_ENV", "development"),
}

/* ----------------- Helpers ----------------- */

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvAsDuration(key string, fallback int) time.Duration {
	valStr := os.Getenv(key)
	if valStr == "" {
		return time.Duration(fallback) * time.Second
	}

	val, err := strconv.Atoi(valStr)
	if err != nil || val <= 0 {
		return time.Duration(fallback) * time.Second
	}

	return time.Duration(val) * time.Second
}
