package main

import (
	"log"

	"wunderlist-backend/config"
	"wunderlist-backend/db"
	"wunderlist-backend/internal/auth"
)

func main() {

	client := db.ConnectMongo()
	database := client.Database(config.AppConfig.MongoDBName)

	auth.InitJWT(
		config.AppConfig.JWTSecret,
		config.AppConfig.AccessTokenTTL,
		config.AppConfig.RefreshTokenTTL,
	)

	auth.InitSessionStore(database.Collection("sessions"))

	log.Println("auth initialized")
}
