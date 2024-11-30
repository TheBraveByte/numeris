package main

import (
	// "log/slog"
	"net/http"
	// "os"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/envvar"
	// "github.com/joho/godotenv"

	"github.com/thebravebyte/numeris/app"
	infra "github.com/thebravebyte/numeris/db"
	"github.com/thebravebyte/numeris/db/repository"
	"github.com/thebravebyte/numeris/db/service"
)

// main is the entry point of the Numeris application. It sets up the server,
// initializes the application, connects to the database, and starts listening for incoming requests.
func main() {
    // set the app and the environment variable
    // err := godotenv.Load(".env")
    // if err != nil {
    // 	slog.Error("Error loading .env file")
    // }

    // initialize the app
    srv := fiber.New(fiber.Config{
        Prefork:           true,
        ServerHeader:      "Fiber",
        StrictRouting:     true,
        CaseSensitive:     true,
        AppName:           "Numeris Test App",
        EnablePrintRoutes: true,
    })

    // add configurations for environment variables
    // Configure and use envvar middleware to expose specific environment variables
    srv.Use("/expose/envvars", envvar.New(
        envvar.Config{
            ExportVars:  map[string]string{"API_KEY": "numeris_api_key"},
            ExcludeVars: map[string]string{"DATABASE_URI": ""},
        },
    ))

    // get connected to the database
    client := infra.Init("mongodb+srv://ayaaakinleye:2701Akin2000@cluster0.opv1wfb.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0")
    // deferring the disconnection of the database
    defer infra.ShutDown(client)

    // initialize all the services and repository
    // initialize the user, invoice and activity repository and any serivce available
    userRepository := repository.UserRepository{}
    invoiceRepository := repository.InvoiceRepository{}
    activityRepository := repository.ActivityRepository{}
    passwordHasher := service.PasswordHasher{}
    authenticatejwt := service.AuthenticateJWT{}

    // connect to the database and other services to the application server
    app := app.NewApplication(
        client,
        passwordHasher,
        authenticatejwt,
        activityRepository,
        userRepository,
        invoiceRepository,
    )

    // initialize the notification
    // TODO [Yusuf]

    Router(srv, app)

    err := srv.Listen(":8080")
    if err != nil && err != http.ErrServerClosed {
        panic(err)
    }

    err = srv.ShutdownWithTimeout(5 * time.Second)
    if err != nil {
        return
    }
}
