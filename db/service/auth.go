package service

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"regexp"
	"time"

	"github.com/golang-jwt/jwt/v4"
	"go.mongodb.org/mongo-driver/bson/primitive"

	infra "github.com/thebravebyte/numeris/db"
)

type AuthenticateJWT struct{}

const emailRegex = `^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`

// // AuthAccessToken type struct which is used to create/generate JWT tokens.
// type AuthAccessToken struct {
// 	UserUUID string `json:"_id"`
// 	Email    string `json:"email"`
// 	jwt.RegisteredClaims
// }

// validateEmail checks if the provided email address is in a valid format.
func validateEmail(email string) error {
	re, err := regexp.Compile(emailRegex)
	if err != nil {
		return fmt.Errorf("failed to compile regex: %w", err)
	}

	if re.MatchString(email) {
		return nil
	}

	return errors.New("invalid email format")
}

// NewAuthAccessToken creates a new AuthAccessToken struct.
func (a *AuthenticateJWT) GenerateJWTToken(userUUID, email string) (string, error) {
	if err := validateEmail(email); err != nil {
		slog.Error("invalid email format for this user", "UUID", userUUID, "email", email)
		return "", err
	}

	_, err := primitive.ObjectIDFromHex(userUUID)
	if err != nil {
		slog.Error("invalid UUID", "UUID", userUUID)
		return "", err
	}

	slog.Info("Creating new access token", "UUID", userUUID, "email", email)

	auth := &infra.AuthAccessToken{
		UserUUID: userUUID,
		Email:    email,
		RegisteredClaims: jwt.RegisteredClaims{
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			Subject:   userUUID,
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(48 * time.Hour)),
			Issuer:    "https://github.com/thebravebyte/numeris",
		},
	}

	// GenerateJWTToken creates a JWT token using the ES256 algorithm and the claims.
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, auth).SignedString([]byte(os.Getenv("AUTH_TOKEN_KEY")))
	if err != nil {
		slog.Error("error while generating token", "error", err)
		return "", err
	}

	return token, nil
}

// ParseToken validates the JWT token and returns the claims if valid.
func (a *AuthenticateJWT) ParseToken(tokenValue string) (*infra.AuthAccessToken, error) {
	token, err := jwt.ParseWithClaims(tokenValue, &infra.AuthAccessToken{}, func(t *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("AUTH_TOKEN_KEY")), nil
	})

	// parse and validate token created
	if err != nil {
		var validationErr *jwt.ValidationError
		if errors.As(err, &validationErr) {
			switch {

			case validationErr.Errors&jwt.ValidationErrorSignatureInvalid != 0:
				slog.Error("invalid token signature", "token", tokenValue)
				return nil, errors.New("invalid token signature")

			case validationErr.Errors&jwt.ValidationErrorExpired != 0:
				slog.Error("token has expired", "token", tokenValue)
				return nil, errors.New("token has expired")

			case validationErr.Errors&jwt.ValidationErrorNotValidYet != 0:
				slog.Error("token is not valid yet", "token", tokenValue)
				return nil, errors.New("token is not valid yet")

			default:
				slog.Error("invalid token", "token", tokenValue, "error", err)
				return nil, errors.New("invalid token")
			}
		} else {
			slog.Error("error parsing token", "token", tokenValue, "error", err)
			return nil, err
		}
	}

	// check for the validation and the expiration time
	if claims, ok := token.Claims.(*infra.AuthAccessToken); ok && token.Valid {
		if claims.ExpiresAt != nil && claims.ExpiresAt.Time.Before(time.Now()) {
			slog.Error("token has expired", "token", tokenValue)
			return nil, errors.New("token has expired")
		}
		return claims, nil
	} else {
		slog.Error("invalid token claims or token", "tokenValue", tokenValue)
		return nil, errors.New("invalid token claims or token")
	}
}
