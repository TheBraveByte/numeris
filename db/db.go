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
func Init(databaseURI string) *mongo.Client {
	// attempting to connect to the database with a maximum of 5 trials
	var trial int

	for {
		db, err := Connect(databaseURI)
		if err != nil {
			trial++

			slog.Error("Init Timeout -----> Unable to connect to database", "Main Error", err, "Connection Attempt", trial)

			// if the application tries to connect to the database 5 times and fails, log the error and panic
			if trial == 5 {
				slog.Error("Init Timeout -----> Unable to connect to database", "Main Error", err, "Attempt", trial)
				panic(err)
			}

			time.Sleep(5 * time.Second)
			continue
		}

		// if the application successfully connects to the database, return the database client
		return db
	}
}

// ConnectDB connects to the database
func Connect(databaseURI string) (*mongo.Client, error) {
	// set a timeout for the database connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// configure the database pool
	opts := &options.ClientOptions{
		ConnectTimeout: &MaxLifeTime,
		MaxPoolSize:    &MaxOpenConns,
	}

	// try to connect to the database
	client, err := mongo.Connect(ctx, opts.ApplyURI(databaseURI))
	if err != nil {
		return nil, fmt.Errorf("ConnectDB -> Unable to connect to database: %v", err)
	}

	// check if the connection is actually working by pinging the database
	err = client.Ping(ctx, readpref.Primary())
	if err != nil {
		slog.Error("ConnectDB -> Unable to ping database", "Main Error", err)
		return nil, fmt.Errorf("ConnectDB -> Unable to ping database: %v", err)
	}

	return client, nil
}

// ShutDown
func ShutDown(client *mongo.Client) {
	defer func(ctx context.Context, client *mongo.Client) {
		if err := client.Disconnect(context.TODO()); err != nil {
			slog.Error("ShutDown -> Unable to disconnect from database", "Main Error", err)
		}
	}(context.TODO(), client)
}
