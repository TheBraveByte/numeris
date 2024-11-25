package service

import infra "github.com/thebravebyte/numeris/db"

type AuthenticateJWT interface {
	GenerateJWTToken(userUUID, email string) (string, error)
	ParseToken(tokenValue string) (*infra.AuthAccessToken, error)
}
