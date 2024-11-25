package service

type PasswordHasher interface {
	CreateHash(password string) (string, error)
	VerifyPassword(hashPassword, password string) (bool, error)
}