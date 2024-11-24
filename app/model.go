package app

import (
	"time"
)

type Auth struct {
	Token string `json:"token" bson:"token" validate:"required"`
}

// User: user details and informations
type User struct {
	ID             string         `json:"id" bson:"_id,omitempty" validate:"required"`
	Name           string         `json:"name" bson:"name" validate:"required"`
	Email          string         `json:"email" bson:"email" validate:"required,email"`
	PhoneNumber    string         `json:"phone_number" bson:"phone_number" validate:"required"`
	CreatedAt      time.Time      `json:"created_at" bson:"created_at" validate:"required"`
	UpdatedAt      time.Time      `json:"updated_at" bson:"updated_at" validate:"required"`
	InvoiceSummary InvoiceSummary `json:"invoice_summary" bson:"invoice_summary"`
	ActivityLog    []Activity     `json:"activity_log" bson:"activity_log"`
	Token          string         `json:"token,omitempty" bson:"token,omitempty"`
}

// Invoice: invoice information for every user activities
type Invoice struct {
	ID              string             `json:"id" bson:"_id,omitempty" validate:"required"`
	InvoiceNumber   string             `json:"invoice_number" bson:"invoice_number" validate:"required"`
	IssueDate       time.Time          `json:"issue_date" bson:"issue_date" validate:"required"`
	DueDate         time.Time          `json:"due_date" bson:"due_date" validate:"required"`
	BillingCurrency string             `json:"billing_currency" bson:"billing_currency" validate:"required"`
	Items           []Item             `json:"items" bson:"items" validate:"required,dive"`
	Discount        float64            `json:"discount,omitempty" bson:"discount,omitempty"`
	TotalAmountDue  float64            `json:"total_amount_due" bson:"total_amount_due" validate:"required"`
	PaymentInfo     PaymentInformation `json:"payment_info" bson:"payment_info" validate:"required"`
	ActivityLog     []Activity         `json:"activity_log,omitempty" bson:"activity_log,omitempty"`
	Customer        CustomerDetails    `json:"customer" bson:"customer" validate:"required"`
	Sender          SenderDetails      `json:"sender" bson:"sender" validate:"required"`
	Notes           string             `json:"notes,omitempty" bson:"notes,omitempty"`
	Reminders       []InvoiceReminder  `json:"reminders,omitempty" bson:"reminders,omitempty"`
	CreatedAt       time.Time          `json:"created_at" bson:"created_at" validate:"required"`
	UpdatedAt       time.Time          `json:"updated_at" bson:"updated_at" validate:"required"`
}

type Item struct {
	Description string  `json:"description" bson:"description" validate:"required"`
	Quantity    int     `json:"quantity" bson:"quantity" validate:"required,min=1"`
	UnitPrice   float64 `json:"unit_price" bson:"unit_price" validate:"required"`
	TotalPrice  float64 `json:"total_price" bson:"total_price" validate:"required"`
}

type PaymentInformation struct {
	AccountName   string `json:"account_name" bson:"account_name" validate:"required"`
	AccountNumber string `json:"account_number" bson:"account_number" validate:"required"`
	RoutingNumber string `json:"routing_number" bson:"routing_number" validate:"required"`
	BankName      string `json:"bank_name" bson:"bank_name" validate:"required"`
}

// Activity this describes the activities of the user invoices created, issues and the likes
type Activity struct {
	Actor     string    `json:"actor" bson:"actor" validate:"required"`
	Action    string    `json:"action" bson:"action" validate:"required"`
	Timestamp time.Time `json:"timestamp" bson:"timestamp" validate:"required"`
	Details   string    `json:"details,omitempty" bson:"details,omitempty"`
}

// CustomerDetails details of a customer an invoices is created for
type CustomerDetails struct {
	Name    string `json:"name" bson:"name" validate:"required"`
	Phone   string `json:"phone" bson:"phone" validate:"required"`
	Email   string `json:"email" bson:"email" validate:"required,email"`
	Address string `json:"address" bson:"address" validate:"required"`
}

// SenderDetails describes the sender details
type SenderDetails struct {
	Name    string `json:"name" bson:"name" validate:"required"`
	Phone   string `json:"phone" bson:"phone" validate:"required"`
	Email   string `json:"email" bson:"email" validate:"required,email"`
	Address string `json:"address" bson:"address" validate:"required"`
}

// InvoiceReminder this is more like a notification for the invoice for the user.
type InvoiceReminder struct {
	DaysBeforeDueDate int    `json:"days_before_due_date" bson:"days_before_due_date" validate:"required,min=1"`
	Message           string `json:"message" bson:"message" validate:"required"`
}

// InvoiceSummary this is the overall summary of the invoice for the users
type InvoiceSummary struct {
	TotalPaid    float64 `json:"total_paid" bson:"total_paid"`
	TotalOverdue float64 `json:"total_overdue" bson:"total_overdue"`
	TotalDraft   float64 `json:"total_draft" bson:"total_draft"`
	TotalUnpaid  float64 `json:"total_unpaid" bson:"total_unpaid"`
}

type EmailTemplate struct {
	UUID     string `json:"uuid" bson:"uuid" validate:"required"`
	Subject  string `json:"subject" bson:"subject" validate:"required"`
	Content  string `json:"content" bson:"content" validate:"required"`
	Receiver string `json:"receiver" bson:"receiver" validate:"required"`
	Sender   string `json:"sender" bson:"sender" validate:"required"`
	Template string `json:"template,omitempty" bson:"template,omitempty"`
}

type IPInfo struct {
	IP        string `json:"ip" bson:"ip" validate:"required"`
	City      string `json:"city" bson:"city,omitempty"`
	Region    string `json:"region" bson:"region,omitempty"`
	Country   string `json:"country" bson:"country,omitempty"`
	Latitude  string `json:"latitude" bson:"latitude,omitempty"`
	Longitude string `json:"longitude" bson:"longitude,omitempty"`
}
