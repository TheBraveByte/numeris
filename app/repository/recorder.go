package repository

import (
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/thebravebyte/numeris/domain"
)

type ActivityRepository interface {
	Save(db *mongo.Client, activity *domain.Activity) error
	GetInvoiceActivities(db *mongo.Client, userID string, limit int64) ([]domain.Activity, error)
}
