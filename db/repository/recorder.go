package repository

import (
	"context"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/thebravebyte/numeris/domain"
)

type ActivityRepository struct {
}

func (r *ActivityRepository) Save(db *mongo.Client, activity *domain.Activity) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	_, err := RecordActivityData(db, "activity").InsertOne(ctx, activity)
	if err != nil {
		panic(fmt.Errorf("error while saving application activity: %v", err))
	}
	return nil
}
