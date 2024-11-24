package service

import (
	"errors"

	"golang.org/x/crypto/bcrypt"

	infra "github.com/thebravebyte/numeris/db"
)

type PasswordHasher struct{}

// CreateHash generates a cryptographic key from the given password using bcrypt
func (hs *PasswordHasher) CreateHash(password string) (string, error) {
	// check if the password is empty
	if password == "" {
		return "", errors.New("invalid input password from user")
	}

	// create a new hash key from the password
	hashedString, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", errors.New("unable to generate hash key from the password")
	}

	// return the hash key
	return string(hashedString), nil
}

// VerifyPassword checks and verifies the hashed password and input password.
func (hs *PasswordHasher) VerifyPassword(hashPassword, password string) (bool, error) {
	// check if the password is empty
	if password == "" || hashPassword == "" {
		return false, infra.ErrInvalidLoginDetails
	}

	// check if the hashPassword is empty
	err := bcrypt.CompareHashAndPassword([]byte(password), []byte(hashPassword))
	if err != nil {
		if errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, infra.ErrUnMatchedPassword
		}
		// return the error
		return false, err
	}

	return true, nil
}
