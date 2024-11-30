package repository

import (
	"context"
	"errors"
	"fmt"
	"log"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"

	infra "github.com/thebravebyte/numeris/db"
	"github.com/thebravebyte/numeris/domain"
)

type UserRepository struct{}


// AddUser adds a new user to the database or retrieves an existing user.
// It first checks if a user with the given email already exists. If not, it adds the new user.
// If the user exists, it returns the existing user information.
//
// Parameters:
//   - db: A pointer to the MongoDB client used for database operations.
//   - user: A pointer to the domain.User struct containing the user information to be added.
//   - email: The email address of the user, used to check for existing accounts.
//
// Returns:
//   - A pointer to the domain.User struct containing the user information (either newly added or existing).
//   - An error if any database operation fails, or nil if successful.
func (repo *UserRepository) AddUser(db *mongo.Client, user *domain.User, email string) (*domain.User, error) {
    ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancelCtx()

    // existingUser variable is more of a placeholder for data received from the database
    var existingUser infra.User

    filter := bson.D{{Key: "email", Value: email}}

    if err := UserData(db, "user").FindOne(ctx, filter).Decode(&existingUser); err != nil {

        if errors.Is(err, mongo.ErrNoDocuments) {

            _, err := UserData(db, "user").InsertOne(ctx, user)
            if err != nil {
                panic(fmt.Errorf("error while inserting user: %v", err))
            }

            log.Printf("Inserted a new user document: %v", user)
            return user, nil
        }
        panic(fmt.Errorf("error while finding user: %v", err))
    }

    user = infra.UserFromDB(existingUser)
    return user, nil
}

// VerifyLogin function to verify the user login details with respect to the database
func (repo *UserRepository) VerifyLogin(db *mongo.Client, email, password string) (*domain.User, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	var result infra.User

	filter := bson.D{{Key: "email", Value: email}}

	err := UserData(db, "user").FindOne(ctx, filter).Decode(&result)

	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			slog.Error("User not found", "email", email)
			return &domain.User{}, fmt.Errorf("%w:%q", mongo.ErrNoDocuments, infra.ErrUserNotFound)
		}
		return &domain.User{}, fmt.Errorf("%w:%q", err, infra.ErrUserNotFound)
	}

	return infra.UserFromDB(result), nil

}

// SaveToken function to save the token in the database
func (repo *UserRepository) SaveToken(db *mongo.Client, id string, accessToken string) error {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 100*time.Second)
	defer cancelCtx()

	// filter maps to the uuid of the user and update set the new details
	filter := bson.D{{Key: "_id", Value: id}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "token", Value: accessToken}}}}

	// check to make sure that the user exists in the database
	_, err := UserData(db, "user").UpdateOne(ctx, filter, update)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			slog.Error("User not found", "_id", id)
			return fmt.Errorf("%w:%q", mongo.ErrNoDocuments, infra.ErrUserNotFound)
		}
		slog.Error("Error while updating token", "error", err)
		return fmt.Errorf("%w:%q", err, infra.ErrInvalidTokenUpdate)
	}
	return nil
}

// UpdatePassword updates the password of the user with the provided password
func (repo *UserRepository) UpdatePassword(db *mongo.Client, email, password string) error {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancelCtx()

	filter := bson.D{{Key: "email", Value: email}}
	update := bson.D{{Key: "$set", Value: bson.D{{Key: "password", Value: password}}}}

	var result bson.M
	err := UserData(db, "user").FindOneAndUpdate(ctx, filter, update).Decode(&result)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return fmt.Errorf("%v", "cannot reset user password! this acc does not exist")
		}
		slog.Error("Error while updating password", "error", err)
		panic(err)
	}
	return nil
}
