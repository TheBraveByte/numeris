package main

import (
	"os"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/envvar"

	infra "github.com/thebravebyte/numeris/db"
)

func main() {
	// set the app and the environment variable

	// initialize the app
	app := fiber.New(fiber.Config{
		Prefork:           true,
		ServerHeader:      "Fiber",
		StrictRouting:     true,
		CaseSensitive:     true,
		AppName:           "Numeris Test App",
		EnablePrintRoutes: true,
	})

	// add configurations for environment variables
	// Configure and use envvar middleware to expose specific environment variables
	app.Use("/expose/envvars", envvar.New(
		envvar.Config{
			ExportVars:  map[string]string{"API_KEY": "numeris_api_key"},
			ExcludeVars: map[string]string{"DATABASE_URI": ""},
		},
	))


	// get connected to the database
	client := infra.Init(os.Getenv("DATABASE_URI"))
	// deferring the disconnection of the database
	infra.ShutDown(client)

	// connect to the database to the application server
	

	

	// load up all the routers endpoints
	// while all the middlewares are all set up for authorization
	// and authentication, loggers, monitoring endpoints activities

	// provide the router configuration for static routes and
	// documents

	// connect the application to the requires ports numbers

	// i need to add context for the services swarn in a goroutines
}
