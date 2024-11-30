package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/thebravebyte/numeris/domain"
)

type ActivityRepository struct {
}

// Save stores a new activity record in the database.
//
// This function creates a new activity entry in the MongoDB database. It uses a
// timeout context to ensure the operation doesn't hang indefinitely.
//
// Parameters:
//   - db: A pointer to the MongoDB client used for database operations.
//   - activity: A pointer to the domain.Activity struct containing the activity data to be saved.
//
// Returns:
//   - An error if there was a problem saving the activity. If successful, returns nil.
//     Note that this function will panic if it encounters an error during the save operation.
func (r *ActivityRepository) Save(db *mongo.Client, activity *domain.Activity) error {
    ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancel()

    _, err := RecordActivityData(db, "activity").InsertOne(ctx, activity)
    if err != nil {
        panic(fmt.Errorf("error while saving application activity: %v", err))
    }
    return nil
}


// GetInvoiceActivities retrieves invoice-related activities for a specific user.
//
// This function queries the database for activities related to creating, issuing,
// and updating invoices for a given user. The results are sorted by timestamp
// in descending order and limited to the specified number of entries.
//
// Parameters:
//   - db: A pointer to the MongoDB client used for database operations.
//   - userID: A string representing the ID of the user whose activities are being retrieved.
//   - limit: An int64 value specifying the maximum number of activities to return.
//
// Returns:
//   - A slice of domain.Activity containing the retrieved invoice activities.
//   - An error if there was a problem querying the database or decoding the results.
func (i *InvoiceRepository) GetInvoiceActivities(db *mongo.Client, userID string, limit int64) ([]domain.Activity, error) {
    ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
    defer cancelCtx()

    filter := bson.M{
        "user_id": userID,
        "action": bson.M{
            "$in": []string{
                "create_invoice_activity",
                "issue_invoice_activity",
                "update_invoice_activity",
            },
        },
    }

    options := options.Find().SetSort(bson.M{"timestamp": -1}).SetLimit(limit)

    cursor, err := RecordActivityData(db, "activity").Find(ctx, filter, options)
    if err != nil {
        return nil, fmt.Errorf("error finding invoice activities: %v", err)
    }
    defer cursor.Close(ctx)

    var activities []domain.Activity
    if err = cursor.All(ctx, &activities); err != nil {
        return nil, fmt.Errorf("error decoding invoice activities: %v", err)
    }

    return activities, nil
}