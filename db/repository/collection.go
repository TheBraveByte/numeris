package repository

import "go.mongodb.org/mongo-driver/mongo"

// UserCol Setting up the database for the user data collection
func UserCol(dbClient *mongo.Client, collectionName string) *mongo.Collection {
	var userCollection = dbClient.Database("track_space").Collection(collectionName)
	return userCollection
}
