package app

import (
	"log/slog"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// contextWithAuth returns the authorization information to give user access to private routes
func (app *Application) contextWithAuth(c *fiber.Ctx, forceAuth bool) error {
	// getting the authorization header from the request
	if forceAuth {
		authString := c.Get("Authorization")
		if authString == "" {
			return c.Status(fiber.StatusNoContent).JSON(fiber.Map{"error": "No token provided"})
		}

		tokenSlices := strings.Split(authString, " ")
		slog.Info("Checking token", tokenSlices, len(tokenSlices))
		if len(tokenSlices) != 2 || tokenSlices[0] != "Bearer" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": "Invalid authorization header"})
		}

		// getting the token to parse to the access token
		token := tokenSlices[1]
		parse, err := app.authorizeJWT.ParseToken(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{"error": err.Error()})
		}

		c.Locals("token", token)
		c.Locals("id", parse.ID)
		c.Locals("email", parse.Email)

		return c.Next()
	}

	return c.Next()
}
