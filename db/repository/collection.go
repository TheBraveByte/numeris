package repository

import "go.mongodb.org/mongo-driver/mongo"

// UserCol Setting up the database for the user data collection
func UserData(db *mongo.Client, collectionName string) *mongo.Collection {
	return db.Database("numeris_book").Collection(collectionName)
}

func InvoiceData(db *mongo.Client, collectionName string) *mongo.Collection {
	return db.Database("numeris_book").Collection(collectionName)
}

func RecordActivityData(db *mongo.Client, collectionName string) *mongo.Collection {
	return db.Database("numeris_book").Collection("activity")
}
