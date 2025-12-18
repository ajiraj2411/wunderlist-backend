package auth

import "go.mongodb.org/mongo-driver/bson/primitive"

// Used for OAuth logic & tests
func FindOrCreateGoogleUser(email, name string) (primitive.ObjectID, error) {
	return primitive.NewObjectID(), nil
}
