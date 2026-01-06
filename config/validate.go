package config

import "log"

func Validate() {
	if AppConfig.JWTSecret == "" {
		log.Fatal("JWT_SECRET must be set")
	}

	if AppConfig.MongoURI == "" {
		log.Fatal("MONGO_URI must be set")
	}

	if AppConfig.Port == "" {
		log.Fatal("PORT must be set")
	}
}
