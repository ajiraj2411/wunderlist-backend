package auth

import (
	"github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/mongo"
)

var jwtBlacklist *JWTBlacklist

func InitJWTBlacklist(rdb *redis.Client) {
	jwtBlacklist = NewJWTBlacklist(rdb)
}

func InitPasswordResetStore(col *mongo.Collection) {
	resetTokenCol = col
}
