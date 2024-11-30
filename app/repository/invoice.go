package repository

import (
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/thebravebyte/numeris/domain"
)

type InvoiceRepository interface {
	AddNewInvoice(db *mongo.Client, userID string, invoice *domain.Invoice) error
	FindUserInvoiceByID(db *mongo.Client, userID, invoiceID string) (*domain.Invoice, error)
	FindAllInvoice(db *mongo.Client, userID string) ([]*domain.Invoice, error)
	InvoiceStatSummary(db *mongo.Client, userID string) (*domain.InvoiceSummary, error)
	InvoiceItemSummary(db *mongo.Client, userID string, invoiceID string) ([]domain.Item, error)

	UpdateInvoiceBeforeDueDate(db *mongo.Client, userID string, invoiceID string, updatedInvoice *domain.Invoice) error
	GetIssueInvoiceList(db *mongo.Client, userID string) ([]domain.Invoice, error)
	UpdateInvoiceStatusToIssued(db *mongo.Client, userID string, invoiceID string) error

	DeleteInvoice(db *mongo.Client, userID, invoiceID string) error
}
