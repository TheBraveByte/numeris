package domain

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/LukaGiorgadze/gonull"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

// User represents a user.
type User struct {
	ID          string    `json:"id" bson:"_id,omitempty" validate:"required"`
	FirstName   string    `json:"first_name" bson:"first_name" validate:"required"`
	LastName    string    `json:"last_name" bson:"last_name" validate:"required"`
	Email       string    `json:"email" bson:"email" validate:"required,email"`
	Password    string    `json:"password" bson:"password" validate:"required"`
	PhoneNumber string    `json:"phone_number" bson:"phone_number" validate:"required"`
	CreatedAt   time.Time `json:"created_at" bson:"created_at" validate:"required"`
	UpdatedAt   time.Time `json:"updated_at" bson:"updated_at" validate:"required"`
	// InvoiceSummary InvoiceSummary `json:"invoice_summary" bson:"invoice_summary"`
	// ActivityLog    []Activity     `json:"activity_log" bson:"activity_log"`
	Token string `json:"token,omitempty" bson:"token,omitempty"`
}

// NewUser creates a new User
func NewUser(
	firstName, lastName, email, password, phoneNumber string) (*User, error) {

	if err := validateEmail(email); err != nil {
		return &User{}, nil
	}

	if err := validateFields(firstName, lastName, email, gonull.NewNullable(phoneNumber)); err != nil {
		return &User{}, nil
	}

	return &User{
		ID:          primitive.NewObjectID().Hex(),
		FirstName:   firstName,
		LastName:    lastName,
		Email:       strings.ToLower(email),
		Password:    password,
		PhoneNumber: phoneNumber,
	}, nil
}

// validateEmail checks if the provided email address is in a valid format.
func validateEmail(email string) error {
	// using regular expression for validating an email address.

	// compile the regular expression
	re, err := regexp.Compile(emailRegex)
	if err != nil {
		return fmt.Errorf("failed to compile regex: %w", err)
	}

	// using the regex to match the email address of the user
	if re.MatchString(email) {
		return nil
	}

	return errors.New("invalid email format")
}

// validateFields checks if any of the provided fields are empty
func validateFields(firstName, lastName, email string, phoneNumber gonull.Nullable[string]) error {
	if strings.TrimSpace(firstName) == "" {
		return fmt.Errorf("%w:%q", ErrInvalidFirstName, "cannot be empty")
	}
	if strings.TrimSpace(lastName) == "" {
		return fmt.Errorf("%w:%q", ErrInvalidLastName, "cannot be empty")
	}
	if strings.TrimSpace(email) == "" {
		return fmt.Errorf("%w:%q", ErrInvalidEmail, "cannot be empty, provide a valid email")
	}
	if phoneNumber.Valid {
		if strings.TrimSpace(phoneNumber.Val) == "" {
			return fmt.Errorf("%w:%q", ErrInvalidPhoneNumber, "phone number cannot be empty")
		}
	}

	return nil
}
