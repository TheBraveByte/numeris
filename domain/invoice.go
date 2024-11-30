package domain

import (
	"errors"
	"fmt"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type Invoice struct {
	InvoiceID       string
	InvoiceNumber   string
	IssueDate       string
	DueDate         string
	BillingCurrency string
	Discount        float64
	TotalAmountDue  float64
	Notes           string
	CreatedAt       time.Time
	UpdatedAt       time.Time
	PaymentInfo     PaymentInformation
	Items           []Item
	Customer        CustomerDetails
	Sender          SenderDetails
}

type Item struct {
	Description string
	Quantity    int
	UnitPrice   float64
	TotalPrice  float64
}

type PaymentInformation struct {
	AccountName   string
	AccountNumber string
	RoutingNumber string
	BankName      string
}

type CustomerDetails struct {
	Name    string
	Phone   string
	Email   string
	Address string
}

type SenderDetails struct {
	Name    string
	Phone   string
	Email   string
	Address string
}

// NewInvoice creates a new Invoice object with the provided details.
// It validates the input parameters and returns an error if any validation fails.
//
// Parameters:
//   - userID: A string representing the user ID associated with the invoice.
//   - invoiceNumber: A string representing the unique invoice number.
//   - billingCurrency: A string representing the currency used for billing.
//   - discount: A float64 representing the discount percentage to be applied to the total amount.
//   - issueDate: A time.Time representing the date the invoice was issued.
//   - dueDate: A time.Time representing the date the invoice is due.
//   - items: A slice of Item structs representing the items included in the invoice.
//   - paymentInfo: A PaymentInformation struct containing the payment details for the invoice.
//   - customer: A CustomerDetails struct containing the customer's information.
//   - sender: A SenderDetails struct containing the sender's information.
//
// Returns:
//   - A pointer to an Invoice struct if the creation is successful.
//   - An error if any validation fails or if required fields are missing.
func NewInvoice(
	userID,
	invoiceNumber, billingCurrency string,
	discount float64,
	issueDate, dueDate string,
	items []Item,
	paymentInfo PaymentInformation,
	customer CustomerDetails,
	sender SenderDetails,
) (*Invoice, error) {
	// validate inputs
	if invoiceNumber == "" {
		return nil, errors.New("invoice number cannot be empty")
	}
	if billingCurrency == "" {
		return nil, errors.New("billing currency cannot be empty")
	}
	if discount < 0 || discount > 100 {
		return nil, errors.New("discount must be between 0 and 100")
	}
	if len(items) == 0 {
		return nil, errors.New("invoice must have at least one item")
	}

	// validate customer details
	if err := validateDetails(customer.Name, customer.Phone, customer.Email, customer.Address); err != nil {
		return nil, errors.New("invalid customer details: " + err.Error())
	}

	// Validate sender details
	if err := validateDetails(sender.Name, sender.Phone, sender.Email, sender.Address); err != nil {
		return nil, errors.New("invalid sender details: " + err.Error())
	}

	for _, item := range items {
		if err := validateItem(item); err != nil {
			return nil, err
		}
	}

	if _, err := validateDates(issueDate, dueDate); err != nil {
		return nil, err
	}

	// validate payment info
	if err := validatePaymentInfo(paymentInfo); err != nil {
		return nil, err
	}

	// calculate total amount
	totalAmount := calculateTotalAmount(items, discount)

	// construct  the invoice model
	invoice := &Invoice{
		InvoiceID:       generateID(),
		InvoiceNumber:   invoiceNumber,
		IssueDate:       issueDate,
		DueDate:         dueDate,
		BillingCurrency: billingCurrency,
		Discount:        discount,
		TotalAmountDue:  totalAmount,
		Notes:           "",
		CreatedAt:       time.Now(),
		UpdatedAt:       time.Now(),
		PaymentInfo:     paymentInfo,
		Items:           items,
	}

	return invoice, nil
}

// AddItem adds an item to the invoice
func (i *Invoice) AddItem(item Item) error {
	if err := validateItem(item); err != nil {
		return err
	}
	i.Items = append(i.Items, item)
	i.TotalAmountDue = calculateTotalAmount(i.Items, i.Discount)
	i.UpdatedAt = time.Now()
	return nil
}

// UpdateDiscount updates the discount and recalculates the total amount due
func (i *Invoice) UpdateDiscount(discount float64) error {
	if discount < 0 || discount > 100 {
		return errors.New("discount must be between 0 and 100")
	}
	i.Discount = discount
	i.TotalAmountDue = calculateTotalAmount(i.Items, i.Discount)
	i.UpdatedAt = time.Now()
	return nil
}

// UpdatePaymentInfo updates the payment information for the invoice
func (i *Invoice) UpdatePaymentInfo(paymentInfo PaymentInformation) error {
	if err := validatePaymentInfo(paymentInfo); err != nil {
		return err
	}
	i.PaymentInfo = paymentInfo
	i.UpdatedAt = time.Now()
	return nil
}

func validateDates(issueDateStr, dueDateStr string) (string, error) {
	const dateFormat = "2006-01-02"

	issueDate, err := time.Parse(dateFormat, issueDateStr)
	if err != nil {
		return "", fmt.Errorf("invalid issue date format: %v", err)
	}

	dueDate, err := time.Parse(dateFormat, dueDateStr)
	if err != nil {
		return "", fmt.Errorf("invalid due date format: %v", err)
	}

	currentDate := time.Now().Truncate(24 * time.Hour)

	if currentDate.After(issueDate) {
		return "", errors.New("issue date cannot be in the past")
	}

	if issueDate.After(dueDate) {
		return "", errors.New("issue date cannot be after due date")
	}

	if currentDate.After(dueDate) {
		return "", errors.New("due date cannot be in the past")
	}

	return "Dates are valid", nil
}

func validateDetails(name, phone, email, address string) error {
	if name == "" {
		return errors.New("name cannot be empty")
	}
	if phone == "" {
		return errors.New("phone cannot be empty")
	}

	if err := validateEmail(email); err != nil {
		return errors.New("email cannot be empty")
	}

	if address == "" {
		return errors.New("address cannot be empty")
	}
	return nil
}

// validatePaymentInfo checks the validity of the payment information
func validatePaymentInfo(paymentInfo PaymentInformation) error {
	if paymentInfo.AccountName == "" {
		return errors.New("account name cannot be empty")
	}
	if paymentInfo.AccountNumber == "" && len(paymentInfo.AccountNumber) < 11 {
		return errors.New("account number cannot be empty")
	}
	if paymentInfo.RoutingNumber == "" && len(paymentInfo.RoutingNumber) < 7 {
		return errors.New("routing number cannot be empty")
	}
	if paymentInfo.BankName == "" {
		return errors.New("bank name cannot be empty")
	}
	return nil
}

// validateItem checks the validity of an item
func validateItem(item Item) error {
	if item.Description == "" {
		return errors.New("item description cannot be empty")
	}
	if item.Quantity <= 0 {
		return errors.New("item quantity must be greater than 0")
	}
	if item.UnitPrice < 0 {
		return errors.New("item unit price must be greater than or equal to 0")
	}
	return nil
}

// calculateTotalAmount calculates the total amount due
func calculateTotalAmount(items []Item, discount float64) float64 {
	total := 0.0
	for _, item := range items {
		total += float64(item.Quantity) * item.UnitPrice
	}
	return total * (1 - discount/100)
}

// generateID generates a unique ID for the invoice
func generateID() string {
	return fmt.Sprintf("INV-%s", primitive.NewObjectID().Hex())
}
