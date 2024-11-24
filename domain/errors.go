package domain

import "errors"

var (
	ErrInvalidEmail       = errors.New("invalid email")
	ErrInvalidFirstName   = errors.New("invalid first name")
	ErrInvalidLastName    = errors.New("invalid last name")
	ErrInvalidPassword    = errors.New("invalid password")
	ErrInvalidPhoneNumber = errors.New("invalid phone number")

	// ErrEmailToken      = errors.New("email already token")
	// ErrUUID            = errors.New("cannot create uuid for user")
	// ErrEmptyName       = errors.New("cannot create empty name")
	// ErrEmptyContent    = errors.New("cannot create empty content")
	// ErrEmptyToolsUse   = errors.New("cannot create empty tools use")
	// ErrEmptyTypeFormat = errors.New("cannot create empty type format")
	// ErrEmptyStatus     = errors.New("cannot create empty status")
)