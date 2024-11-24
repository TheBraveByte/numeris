package infra

import "errors"

var (
	ErrUserNotFound       = errors.New("user not found")
	ErrInternalError      = errors.New("internal error while finding user")
	ErrInvalidTokenUpdate = errors.New("invalid token update")

	ErrUnMatchedPassword   = errors.New("invalid input password")
	ErrInvalidLoginDetails = errors.New("invalid login details")
)
