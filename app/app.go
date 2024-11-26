package app

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	"github.com/thebravebyte/numeris/app/repository"
	"github.com/thebravebyte/numeris/app/service"
	infra "github.com/thebravebyte/numeris/db"
	"github.com/thebravebyte/numeris/domain"
)

type Application struct {
	// Define your application's configuration and dependencies here
	db                 *mongo.Client
	passwordHasher     service.PasswordHasher
	authorizeJWT       service.AuthenticateJWT
	activityRepository repository.ActivityRepository
	userRepository     repository.UserRepository
	invoiceRepository  repository.InvoiceRepository
}

func NewApplication(
	db *mongo.Client,
	passwordHasher service.PasswordHasher,
	authorizeJwt service.AuthenticateJWT,
	activityRepository repository.ActivityRepository,
	userRepository repository.UserRepository,
	invoiceRepository repository.InvoiceRepository,
	// well we can add other dependencies as needed

) *Application {
	return &Application{
		// Initialize your application's dependencies and configurations
		db:                 db,
		passwordHasher:     passwordHasher,
		authorizeJWT:       authorizeJwt,
		activityRepository: activityRepository,
		userRepository:     userRepository,
		invoiceRepository:  invoiceRepository,
	}
}

func (app *Application) SignUpHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		data := new(SignUpRequestModel)

		// bind the request body to the data struct
		if err := c.BodyParser(data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   err.Error(),
				"message": fmt.Sprintf("%s: %q", ErrInvalidInputReceived.Error(), "from the client"),
			})
		}

		// validate the data input
		validateData := FieldValidator(data)
		if validateData != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "invalid input",
				"message": fmt.Sprintf("%q: %v", "from the client", validateData),
			})
		}

		//hHash the user's password
		hashedPassword, err := app.passwordHasher.CreateHash(data.Password)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error": err.Error(),
			})
		}

		// server-side validation of the user input
		user, err := domain.NewUser(
			data.FirstName,
			data.LastName,
			data.Email,
			hashedPassword,
			data.PhoneNumber,
		)
		if err != nil {
			slog.Info("No user is created", "user", data)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "cannot create user",
				"message": ErrInvalidInputReceived.Error(),
			})
		}

		// attempt to add the user to the database
		id, existingUser := app.userRepository.AddUser(app.db, user, user.Email)
		if existingUser != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "cannot add existing user",
				"message": fmt.Sprintf("%s: %q", ErrUserAlreadyExists.Error(), "go back to login"),
			})
		}

		// If the user is created successfully
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": fmt.Sprintf("User: %s has been created successfully", id),
		})
	}
}

// LoginHandler validates the user input and manages login, including generating a token and setting it as a cookie.
func (app *Application) LoginHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		data := new(LoginRequestModel)

		// parse the request body into the LoginRequestModel struct
		if err := c.BodyParser(data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   err.Error(),
				"message": "Invalid input received from the client",
			})
		}

		// validate user input
		validateData := FieldValidator(data)
		if len(validateData) > 0 {
			for _, f := range validateData {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Invalid input",
					"message": fmt.Sprintf("%s: %s", f.Message, f.NameSpace),
				})
			}
		}

		// check and verify the stored hashed password in the database
		user, err := app.userRepository.VerifyLogin(app.db, data.Email, data.Password)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   ErrLoginFailed.Error(),
				"message": err.Error(),
			})
		}

		// lets compare login passowrd with the stored hashed password
		ok, err := app.passwordHasher.VerifyPassword(data.Password, user.Password)
		if !ok || err != nil {
			slog.Info("Password does not match", "user", user)
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   ErrInvalidCredentials,
				"message": err.Error(),
			})
		}

		token, err := app.authorizeJWT.GenerateJWTToken(user.ID, user.Email)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   ErrGenerateToken.Error(),
				"message": err.Error(),
			})
		}

		// save the token in the database
		if err := app.userRepository.SaveToken(app.db, user.ID, token); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   ErrInvalidUpdateToken.Error(),
				"message": err.Error(),
			})
		}

		// set response headers and cookies
		c.Set("Authorization", "Bearer "+token)
		c.Cookie(&fiber.Cookie{
			Name:     "bearerToken",
			Value:    token,
			MaxAge:   60 * 60 * 48,
			Path:     "/login",
			Domain:   "numeris.onrender.com",
			Secure:   false,
			HTTPOnly: true,
		})

		// respond with JSON indicating success
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Login successful",
			"data":    user.ID,
		})
	}
}

func (app *Application) CreateInvoiceHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		data := new(InvoiceRequestModel)
		if err := c.BodyParser(data); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   err.Error(),
				"message": "Invalid input received from the client",
			})
		}

		userID := c.Params("userID")
		if userID == "" {
			slog.Error("Invalid user", "Error", "the userID must be provided")
			panic("user id is invalid")
		}

		if _, err := primitive.ObjectIDFromHex(userID); err != nil {
			slog.Error("Invalid user", "Error", "the userID must be a valid UUID")
			panic(err)
		}

		validateData := FieldValidator(data)
		if len(validateData) > 0 {
			for _, f := range validateData {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Invalid input",
					"message": fmt.Sprintf("%s: %s", f.Message, f.NameSpace),
				})
			}
		}

		// load and store the values in a list
		items := make([]domain.Item, 0)
		for _, val := range data.Items {
			items = append(items,
				domain.Item{
					Description: val.Description,
					Quantity:    val.Quantity,
					UnitPrice:   val.UnitPrice,
					TotalPrice:  val.TotalPrice,
				})
		}

		// create a new invoice object from the input data and store it in memory
		invoice, err := domain.NewInvoice(
			data.InvoiceNumber,
			data.BillingCurrency,
			data.Discount,
			data.IssueDate,
			data.DueDate,
			items,
			domain.PaymentInformation{
				AccountName:   data.PaymentInfo.AccountName,
				AccountNumber: data.PaymentInfo.AccountNumber,
				RoutingNumber: data.PaymentInfo.RoutingNumber,
				BankName:      data.PaymentInfo.BankName,
			},
			domain.CustomerDetails{
				Email:   data.Customer.Email,
				Name:    data.Customer.Name,
				Phone:   data.Customer.Phone,
				Address: data.Customer.Address,
			},
			domain.SenderDetails{
				Email:   data.Sender.Email,
				Name:    data.Sender.Name,
				Phone:   data.Sender.Phone,
				Address: data.Sender.Address,
			},
		)

		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to create invoice",
				"message": err.Error(),
			})
		}

		// Add the invoice to the database
		res := app.invoiceRepository.AddNewInvoice(app.db, invoice)
		if res != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to create invoice",
				"message": err,
			})
		}

		// Record user activity
		go func() {
			activity := &domain.Activity{
				UserID:    userID,
				Action:    infra.CreateInvoiceActivity,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"invoiceID":       invoice.ID,
					"invoiceNumber":   invoice.InvoiceNumber,
					"billingCurrency": invoice.BillingCurrency,
					"totalAmount":     invoice.TotalAmountDue,
				},
			}
			if err := app.activityRepository.Save(app.db, activity); err != nil {
				slog.Error("Failed to record user activity", "error", err)
			}
		}()

		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": fmt.Sprintf("Invoice: %s has been created successfully", invoice.ID),
		})
	}
}

func (a *Application) InvoiceHandler() {
	// Implement your invoice dashboard loading logic here this gonna be alot for me
}
