package repository

import (
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/thebravebyte/numeris/domain"
)

type InvoiceRepository interface {
	AddNewInvoice(db *mongo.Client, invoice *domain.Invoice) interface{}
}
