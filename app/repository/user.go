package repository

import (
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/thebravebyte/numeris/domain"
)

type UserRepository interface {
	AddUser(db *mongo.Client, user *domain.User, email string) (interface{}, *domain.User)
	VerifyLogin(db *mongo.Client, email, password string) (*domain.User, error)
	SaveToken(db *mongo.Client, id string, accessToken string) error
	UpdatePassword(db *mongo.Client, email, password string) error
}
