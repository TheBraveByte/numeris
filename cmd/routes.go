package main

import (
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/requestid"

	"github.com/thebravebyte/numeris/app"
)

// Router sets up the routes and middleware for the Fiber application.
//
// Parameters:
//
//	srv: A pointer to the Fiber application instance where routes and middleware will be registered.
//	app: A pointer to the Application struct that contains the handlers for the routes.
//
// The function does not return any value. It configures the server with various middleware and//
// sets up the HTTP routes for handling user registration, login, invoice management, and activity tracking.//
func Router(srv *fiber.App, app *app.Application) {
	router := srv.Use(requestid.New())
	srv.Use(logger.New(logger.Config{
		Format:        "${pid} ${locals:requestid} ${status} - ${method} ${path}​\n",
		TimeFormat:    time.RFC3339Nano,
		TimeInterval:  time.Nanosecond,
		Output:        nil,
		DisableColors: false,
	}))

	srv.Use(cors.New())
	ConfigDefault := cors.Config{
		Next:             nil,
		AllowOriginsFunc: nil,
		AllowOrigins:     "*",
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
		}, ","),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization, Content-Length, Accept-Encoding, X-CSRF-Token, X-HTTP-Method-Override, X-Requested-With",
		AllowCredentials: false,
		ExposeHeaders:    "Content-Length",
		MaxAge:           int(24 * time.Hour),
	}
	// recall am using a wildcard format to allow external origin
	router.Use(cors.New(ConfigDefault))

	// 
	router.Get("/", func(c *fiber.Ctx) error{
		return c.SendString("Welcome to the numeris API!")
	})

	// let configure the endpoints with the routers http methods
	router.Post("/api/register", app.SignUpHandler())
	router.Post("/api/login", app.LoginHandler())

	// invoices routes
	router.Post("/api/invoice/:userID/create", app.CreateInvoiceHandler())
	router.Get("/api/invoice/:userID/get/:invoiceID", app.GetInvoiceHandler())
	router.Get("/api/invoice/:userID/all/:invoiceID", app.ListAllInvoiceHandler())

	router.Put("/api/invoice/:userID/update/:invoiceID", app.UpdateUnIssuedInvoiceHandler())
	router.Delete("/api/invoice/:userID/delete/:invoiceID", app.DeleteInvoiceHandler())

	router.Get("/api/invoice/:userID/stats/:invoiceID", app.GetUserInvoiceStatHandler())
	router.Post("/api/invoice/:userID/send/:invoiceID", app.SendIssuedInvoiceToCustomer())
	router.Get("/api/invoice/:userID/download/:invoiceID", app.DownloadInvoicePDFHandler())

	// activity routes
	router.Get("/api/invoice/:userID/activities/:invoiceID", app.GetInvoiceActivitiesHandler())

}
