package auth

import "go.mongodb.org/mongo-driver/mongo"

var userCol *mongo.Collection

var resetTokenCol *mongo.Collection

func InitUserStore(col *mongo.Collection) {
	userCol = col
}
