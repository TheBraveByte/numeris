package repository

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/thebravebyte/numeris/domain"
)

type InvoiceRepository struct{}

// AddNewInvoice adds a new invoice to the user's document and synchronizes it with the invoices collection.
// It uses a MongoDB transaction to ensure data consistency and integrity.
//
// Parameters:
// - db: A pointer to the MongoDB client.
// - userID: The unique identifier of the user.
// - invoice: A pointer to the Invoice struct representing the new invoice to be added.
//
// Returns:
// - An error if any error occurs during the process, otherwise nil.
func (i *InvoiceRepository) AddNewInvoice(db *mongo.Client, userID string, invoice *domain.Invoice) error {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	// start session for transaction
	session, err := db.StartSession()
	if err != nil {
		return fmt.Errorf("error starting session: %v", err)
	}
	defer session.EndSession(ctx)

	callback := func(sessCtx mongo.SessionContext) (interface{}, error) {
		filter := bson.D{{Key: "_id", Value: userID}}
		update := bson.D{{Key: "$push", Value: bson.D{{Key: "invoices", Value: invoice}}}}

		_, err := UserData(db, "user").UpdateOne(sessCtx, filter, update)
		if err != nil {
			panic("error while inserting new invoice")
		}

		// insert into the invoices collection
		_, err = InvoiceData(db, "invoice").InsertOne(sessCtx, invoice)
		if err != nil {
			return nil, fmt.Errorf("error inserting into invoices: %v", err)
		}
		return nil, nil

	}

	// execute the transaction
	_, err = session.WithTransaction(ctx, callback)
	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}
	slog.Info("Invoice created and synchronized successfully.")

	return nil
}

// FindUserInvoice retrieves a specific invoice for a given user from the database.
// “
// Parameters:
// - db: A pointer to the MongoDB client used for database operations.
// - userID: The unique identifier of the user whose invoice is being searched.
// - invoiceID: The unique identifier of the invoice to be retrieved.
//
// Returns:
// - A pointer to the domain.Invoice if found.
// - An error if the invoice is not found or if any other error occurs during the database operation.
func (i *InvoiceRepository) FindUserInvoiceByID(db *mongo.Client, userID, invoiceID string) (*domain.Invoice, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	filter := bson.M{"_id": userID, "invoices.invoice_id": invoiceID}
	projection := bson.M{"invoices.$": 1}

	var result struct {
		Invoices []domain.Invoice `bson:"invoices"`
	}

	err := UserData(db, "user").FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("invoice not found for userID %s and invoiceID %s", userID, invoiceID)
		}
		return nil, fmt.Errorf("error finding invoice: %v", err)
	}

	if len(result.Invoices) > 0 {
		return &result.Invoices[0], nil
	}
	return nil, fmt.Errorf("invoice not found in user document")
}

// UpdatePreviousInvoice updates the details of a previous invoice for a given user.
func (i *InvoiceRepository) UpdateInvoiceBeforeDueDate(db *mongo.Client, userID string, invoiceID string, updatedInvoice *domain.Invoice) error {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	session, err := db.StartSession()
	if err != nil {
		return fmt.Errorf("error starting session: %v", err)
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		// Start the transaction
		if err := session.StartTransaction(); err != nil {
			return fmt.Errorf("error starting transaction: %v", err)
		}

		// First, fetch the current invoice to check its dates
		currentInvoice, err := i.FindUserInvoiceByID(db, userID, invoiceID)
		if err != nil {
			session.AbortTransaction(sessCtx)
			return fmt.Errorf("error fetching current invoice: %v", err)
		}

		// Check if the invoice can be updated
		now := time.Now()
		// Parse IssueDate
		issueDate, err := time.Parse("2006-01-02", currentInvoice.IssueDate)
		if err != nil {
			log.Fatalf("invalid issue date format: %v", err)
		}

		// Parse DueDate
		dueDate, err := time.Parse("2006-01-02", currentInvoice.DueDate)
		if err != nil {
			log.Fatalf("invalid due date format: %v", err)
		}

		if issueDate.Before(now) || now.After(dueDate) {
			session.AbortTransaction(sessCtx)
			return fmt.Errorf("invoice cannot be updated: it has been issued or the due date has passed")
		}

		// Proceed with the update
		filter := bson.M{"_id": userID, "invoices.invoice_id": invoiceID}
		update := bson.M{"$set": bson.M{"invoices.$": updatedInvoice}}

		result, err := UserData(db, "user").UpdateOne(ctx, filter, update)
		if err != nil {
			session.AbortTransaction(sessCtx)
			return fmt.Errorf("error updating invoice in user collection: %v", err)
		}

		if result.MatchedCount == 0 {
			session.AbortTransaction(sessCtx)
			return fmt.Errorf("no invoice found with ID %s for user %s", invoiceID, userID)
		}

		// Update the external invoice collection
		filter = bson.M{"invoice_id": invoiceID}
		update = bson.M{"$set": updatedInvoice}
		result, err = InvoiceData(db, "invoice").UpdateOne(ctx, filter, update)

		if err != nil {
			session.AbortTransaction(sessCtx)
			return fmt.Errorf("error updating invoice in invoice collection: %v", err)
		}

		if result.MatchedCount == 0 {
			session.AbortTransaction(sessCtx)
			return fmt.Errorf("no invoice found with ID %s in invoice collection", invoiceID)
		}

		if err := session.CommitTransaction(sessCtx); err != nil {
			return fmt.Errorf("error committing transaction: %v", err)
		}

		return nil
	})

	if err != nil {
		return err
	}

	slog.Info("Invoice updated successfully", "invoiceID", invoiceID)
	return nil
}

// FindAllInvoice retrieves all invoices associated with a given user from the database.
//
// Parameters:
// - db: A pointer to the MongoDB client used for database operations.
// - userID: The unique identifier of the user whose invoices are being searched.
//
// Returns:
// - A slice of pointers to domain.Invoice representing the invoices found for the user.
// - An error if any error occurs during the database operation. If no invoices are found, the function returns nil for the error.
func (i *InvoiceRepository) FindAllInvoice(db *mongo.Client, userID string) ([]*domain.Invoice, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	filter := bson.M{"_id": userID}
	projection := bson.M{"invoices": 1}

	var result struct {
		Invoices []*domain.Invoice `bson:"invoices"`
	}

	err := UserData(db, "user").FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, fmt.Errorf("no user found with ID %s", userID)
		}
		return nil, fmt.Errorf("error finding invoices: %v", err)
	}

	return result.Invoices, nil
}

// InvoiceStatSummary retrieves a summary of invoice statistics for a given user.
// The summary includes the total amount paid, total amount overdue, and total amount of draft invoices.
//
// Parameters:
// - db: A pointer to the MongoDB client used for database operations.
// - userID: The unique identifier of the user whose invoice statistics are being retrieved.
//
// Returns:
// - A pointer to domain.InvoiceSummary containing the invoice statistics.
// - An error if any error occurs during the database operation. If no invoice statistics are found, the function returns nil for the error.
func (i *InvoiceRepository) InvoiceStatSummary(db *mongo.Client, userID string) (*domain.InvoiceSummary, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()
	pipeline := mongo.Pipeline{
		bson.D{{Key: "$match", Value: bson.M{"_id": userID}}},
		bson.D{{Key: "$unwind", Value: "$invoices"}},
		bson.D{{Key: "$match", Value: bson.M{"invoices.status": "paid"}}},
		bson.D{
			{Key: "$group", Value: bson.M{
				"_id": nil,
				"totalPaid": bson.M{
					"$sum": "$invoices.total_amount",
				},
				"totalOverdue": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{
							bson.M{"$and": bson.A{
								bson.M{"$eq": bson.A{"$invoices.status", "overdue"}},
								bson.M{"$lt": bson.A{"$invoices.dueDate", time.Now()}},
							}},
							"$invoices.totalAmount",
							0,
						},
					},
				},
				"totalDraft": bson.M{
					"$sum": bson.M{
						"$cond": bson.A{
							bson.M{"$eq": bson.A{"$invoices.status", "draft"}},
							"$invoices.totalAmount",
							0,
						},
					},
				},
			}},
		},
	}

	cursor, err := UserData(db, "user").Aggregate(ctx, pipeline)
	if err != nil {
		return nil, fmt.Errorf("error aggregating invoice stats: %v", err)
	}
	defer cursor.Close(ctx)

	var results domain.InvoiceSummary
	if err = cursor.All(ctx, &results); err != nil {
		return nil, fmt.Errorf("error decoding invoice stats: %v", err)
	}

	if &results == nil {
		return nil, fmt.Errorf("no invoice stats found for user %s", userID)
	}

	return &results, nil
}

// InvoiceItemSummary retrieves a summary of items in a specific invoice for a given user.
//
// Parameters:
// - db: A pointer to the MongoDB client used for database operations.
// - userID: The unique identifier of the user whose invoice is being searched.
// - invoiceID: The unique identifier of the invoice whose items are being summarized.
//
// Returns:
// - A slice of domain.Item representing the items in the invoice.
// - An error if any error occurs during the database operation. If no invoice is found, the function returns nil for the error.
func (i *InvoiceRepository) InvoiceItemSummary(db *mongo.Client, userID string, invoiceID string) ([]domain.Item, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	filter := bson.M{"_id": userID, "invoices.invoice_id": invoiceID}
	projection := bson.M{"invoices.$": 1}

	var result struct {
		Invoices []domain.Invoice `bson:"invoices"`
	}

	err := UserData(db, "user").FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("error finding invoice: %v", err)
	}

	if len(result.Invoices) == 0 {
		return nil, fmt.Errorf("no invoice found with ID %s for user %s", invoiceID, userID)
	}

	invoice := result.Invoices[0]
	summary := make([]domain.Item, len(invoice.Items))
	for i, item := range invoice.Items {
		summary[i] = domain.Item{
			Description: item.Description,
			Quantity:    item.Quantity,
			UnitPrice:   item.UnitPrice,
			TotalPrice:  item.TotalPrice,
		}
	}

	return summary, nil
}

// GetIssueInvoiceList retrieves a list of invoices that are ready to be issued for a given user within the next 30 days.
// The function filters invoices based on their status (pending, draft, or overdue) and issue date.
//
// Parameters:
// - db: A pointer to the MongoDB client used for database operations.
// - userID: The unique identifier of the user whose invoices are being searched.
//
// Returns:
// - A slice of domain.Invoice representing the invoices that are ready to be issued.
// - An error if any error occurs during the database operation. If no invoices are found, the function returns nil for the error.
func (i *InvoiceRepository) GetIssueInvoiceList(db *mongo.Client, userID string) ([]domain.Invoice, error) {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	readyToIssueStatus := []string{"pending", "draft", "overdue"}

	today := time.Now()
	startDate := today
	endDate := today.AddDate(0, 0, 30)

	filter := bson.M{
		"_id":             userID,
		"invoices.status": bson.M{"$in": readyToIssueStatus},
		"invoices.issueDate": bson.M{
			"$gte": startDate,
			"$lte": endDate,
		},
	}

	projection := bson.M{"invoices.$": 1}

	var result struct {
		Invoices []domain.Invoice `bson:"invoices"`
	}

	err := UserData(db, "user").FindOne(ctx, filter, options.FindOne().SetProjection(projection)).Decode(&result)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, nil
		}
		return nil, fmt.Errorf("error finding invoices for user %s: %v", userID, err)
	}

	return result.Invoices, nil
}

// UpdateInvoiceStatusToIssued updates the status of an invoice to "issued" for a given user and invoice ID.
// It starts a MongoDB session and performs the update operation within a transaction.
// If the update is successful, it commits the transaction and returns nil.
// If any error occurs during the operation, it returns an error message.
func (i *InvoiceRepository) UpdateInvoiceStatusToIssued(db *mongo.Client, userID string, invoiceID string) error {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	session, err := db.StartSession()
	if err != nil {
		return fmt.Errorf("error starting session: %v", err)
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return fmt.Errorf("failed to start transaction: %v", err)
		}

		filter := bson.M{
			"_id":                 userID,
			"invoices.invoice_id": invoiceID,
			"invoices.status":     bson.M{"$in": []string{"pending", "draft", "overdue"}},
		}

		update := bson.M{
			"$set": bson.M{
				"invoices.$.status":    "issued",
				"invoices.$.issue_date": time.Now(),
			},
		}

		_, err := UserData(db, "user").UpdateOne(ctx, filter, update)
		if err != nil {
			session.AbortTransaction(ctx)
			return fmt.Errorf("error updating invoice status: %v", err)
		}

		// Update the external invoice collection
		filter = bson.M{
			"invoice_id": invoiceID,
			"status":     bson.M{"$in": []string{"pending", "draft", "overdue"}},
		}
		update = bson.M{"$set": bson.M{"issued_date": time.Now()}}

		_, err = InvoiceData(db, "invoices").UpdateOne(ctx, filter, update)
		if err != nil {
			session.AbortTransaction(ctx)
			return fmt.Errorf("error updating invoices status: %v", err)
		}

		if err := session.CommitTransaction(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %v", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}

	return nil
}

// DeleteInvoice removes an invoice from both the user's document and the invoices collection.
// It uses a MongoDB transaction to ensure data consistency across collections.
//
// Parameters:
//   - db: A pointer to the MongoDB client used for database operations.
//   - userID: The unique identifier of the user whose invoice is being deleted.
//   - invoiceID: The unique identifier of the invoice to be deleted.
//
// Returns:
//   - An error if any step of the deletion process fails, or nil if the operation is successful.
func (i *InvoiceRepository) DeleteInvoice(db *mongo.Client, userID, invoiceID string) error {
	ctx, cancelCtx := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancelCtx()

	session, err := db.StartSession()
	if err != nil {
		return fmt.Errorf("error starting session: %v", err)
	}
	defer session.EndSession(ctx)

	err = mongo.WithSession(ctx, session, func(sessCtx mongo.SessionContext) error {
		if err := session.StartTransaction(); err != nil {
			return fmt.Errorf("failed to start transaction: %v", err)
		}
		// user collection
		filter := bson.M{
			"_id":                 userID,
			"invoices.invoice_id": invoiceID,
		}

		_, err := UserData(db, "user").DeleteOne(ctx, filter)
		if err != nil {
			session.AbortTransaction(ctx)
			return fmt.Errorf("error updating invoice status: %v", err)
		}

		// delete the external invoice collection
		filter = bson.M{
			"invoice_id": invoiceID,
		}

		_, err = InvoiceData(db, "invoices").DeleteOne(ctx, filter)
		if err != nil {
			session.AbortTransaction(ctx)
			return fmt.Errorf("error updating invoices status: %v", err)
		}

		if err := session.CommitTransaction(ctx); err != nil {
			return fmt.Errorf("failed to commit transaction: %v", err)
		}

		return nil
	})

	if err != nil {
		return fmt.Errorf("transaction failed: %v", err)
	}

	return nil
}
