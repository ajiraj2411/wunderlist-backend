package auth

import "github.com/redis/go-redis/v9"

var jwtBlacklist *JWTBlacklist

func InitJWTBlacklist(rdb *redis.Client) {
	jwtBlacklist = NewJWTBlacklist(rdb)
}
