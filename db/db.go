package infra

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

// Database connection pool configuration
var (
	MaxIdleTime  time.Duration = 2 * time.Minute
	MaxLifeTime  time.Duration = 5 * time.Minute
	MaxOpenConns uint64        = 5
	MaxIdleConns uint64        = 5
)

// InitDB initializes the database connection
// Init initializes the database connection.
// It attempts to connect to the MongoDB database using the provided database URI.
// If the connection fails, it logs the error and retries up to 10 times with a 5-second delay between attempts.
// If all attempts fail, it panics with the last error.
// If the connection is successful, it returns the MongoDB client connection.
//
// Parameters:
// - databaseURI: A string representing the URI of the MongoDB database.
//
// Return:
// - A pointer to the MongoDB client connection if the connection is successful.
// - An error if the connection fails.
func Init(databaseURI string) *mongo.Client {
	// attempting to connect to the database with a maximum of 10 trials
	var trial int

	for {
		db, err := Connect(databaseURI)
		if err != nil {
			trial++

			slog.Error("cannot connect to database", "error", err, "connect attempt: ", trial)

			// if the application tries to connect to the database 10 times and fails, log the error and panic
			if trial == 10 {
				slog.Error("cannot connect to database", "count attempt", trial)
				panic(err)
			}

			time.Sleep(5 * time.Second)
			continue
		}

		// if the application successfully connects to the database, return the database client
		return db
	}
}

// Connect connects to the database
// Connect connects to the MongoDB database using the provided database URI.
// It sets a timeout for the database connection and configures the database pool.
// The function attempts to connect to the database and returns the client connection.
// If an error occurs during the connection process, it returns nil and an error message.
//
// Parameters:
// - databaseURI: A string representing the URI of the MongoDB database.
//
// Return:
// - A pointer to the MongoDB client connection if the connection is successful.
// - An error if the connection fails.
func Connect(databaseURI string) (*mongo.Client, error) {
	// set a timeout for the database connection
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// // configure the database pool
	// opts := &options.ClientOptions{
	// 	ConnectTimeout: &MaxLifeTime,
	// 	MaxPoolSize:    &MaxOpenConns,
	// }

	// try to connect to the database
	client, err := mongo.Connect(ctx, options.Client().ApplyURI(databaseURI))
	if err != nil {
		return nil, fmt.Errorf("unable to connect to database: %v", err)
	}

	// check if the connection is actually working by pinging the database
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		slog.Error("unable to ping database", "Error:", err)
		return nil, fmt.Errorf("unable to ping database: %v", err)
	}

	return client, nil
}

// ShutDown gracefully closes the MongoDB client connection.
// This function uses a deferred function to ensure that the client connection is closed
// even if an error occurs during the execution of the main function.
//
// Parameters:
// - client: A pointer to the MongoDB client connection. This parameter is required and cannot be nil.
// //+
// Return:
// This function does not return any value.//
func ShutDown(client *mongo.Client) {
	defer func(ctx context.Context, client *mongo.Client) {
		if err := client.Disconnect(context.TODO()); err != nil {
			slog.Error("Unable to disconnect from database", "Error", err)
			panic(err)
		}
	}(context.TODO(), client)
}
