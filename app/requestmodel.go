package app

// LoginRequestModel represents a request to login
type LoginRequestModel struct {
	Email    string `json:"email" Usage:"required,email"`
	Password string `json:"password" Usage:"min=8,max=20"`
}

// SignUpRequestModel represents a request to sign up a user
type SignUpRequestModel struct {
	FirstName   string `json:"first_name" Usage:"required,alpha"`
	LastName    string `json:"last_name" Usage:"required,alpha"`
	Email       string `json:"email" Usage:"required,email"`
	Password    string `json:"password" Usage:"min=8,max=20"`
	PhoneNumber string `json:"phone_number" Usage:"required"`
	Profession  string `json:"profession"`
}

// ResetPasswordRequestModel to reset user password
type ResetPasswordRequestModel struct {
	Email           string `json:"email" Usage:"required,email"`
	Password        string `json:"password" Usage:"min=8,max=20"`
	ConfirmPassword string `json:"confirm_password" Usage:"required"`
}

// InvoiceRequestModel to create a invoice
type InvoiceRequestModel struct {
	// NetDayRange     int                `json:"net_range" validate:"required"`
	BillingCurrency string             `json:"billing_currency"`
	Items           []Item             `json:"items" validate:"required"`
	InvoiceNumber   string             `json:"invoice_number"`
	Discount        float64            `json:"discount"`
	PaymentInfo     PaymentInformation `json:"payment_info" validate:"required"`
	Notes           string             `json:"notes"`
	Customer        CustomerDetails    `json:"customer" validate:"required"`
	Sender          SenderDetails      `json:"sender" validate:"required"`
	IssueDate       string             `json:"issue_date" validate:"required"`
	DueDate         string             `json:"due_date"`
}

type UpdateInvoiceStatusRequestModel struct {
	Status string `json:"status" validate:"required"`
}
