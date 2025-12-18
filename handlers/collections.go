package handlers

import "go.mongodb.org/mongo-driver/mongo"

var (
	UserCollection    *mongo.Collection
	ListCollection    *mongo.Collection
	TaskCollection    *mongo.Collection
	SessionCollection *mongo.Collection
)

func SetUserCollection(c *mongo.Collection) {
	UserCollection = c
}

func SetListCollection(c *mongo.Collection) {
	ListCollection = c
}

func SetTaskCollection(c *mongo.Collection) {
	TaskCollection = c
}

func SetSessionCollection(c *mongo.Collection) {
	SessionCollection = c
}
