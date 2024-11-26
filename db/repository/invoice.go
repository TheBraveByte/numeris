package repository

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/mongo"

	"github.com/thebravebyte/numeris/domain"
)

type InvoiceRepository struct{}

func (i *InvoiceRepository) AddNewInvoice(db *mongo.Client, invoice *domain.Invoice) interface{} {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)

	defer cancelCtx()
	//insert and no need to check for duplicate invoices because, the invoices are uniquely created
	res, err := InvoiceData(db, "invoice").InsertOne(ctx, invoice)
	if err != nil {
		panic("error while inserting new invoice")
	}

	return res.InsertedID

}

func (i *InvoiceRepository) UpdatePreviousInvoice() {
	return
}

func (i *InvoiceRepository) InvoiceStatSummary() {
	return
}

func (i *InvoiceRepository) InvoiceActivities() {
	return
}

func (i *InvoiceRepository) InvoiceItemSummary() {
	return
}
