package app

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")

	ErrNoDataFound   = errors.New("no data found")
	ErrInternalError = errors.New("internal server error")

	ErrUserAlreadyExists    = errors.New("user already exists")
	ErrInvalidInputReceived = errors.New("invalid input request received")

	ErrLoginFailed        = errors.New("login failed")
	ErrGenerateToken      = errors.New("cannot generate jwt token")
	ErrInvalidUpdateToken = errors.New("invalid token update")
)
