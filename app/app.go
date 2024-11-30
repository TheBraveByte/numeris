package app

import (
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"

	infra "github.com/thebravebyte/numeris/db"
	"github.com/thebravebyte/numeris/db/repository"
	"github.com/thebravebyte/numeris/db/service"
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

// NewApplication initializes a new application with the provided dependencies.
//
// Parameters:
//   - db: *mongo.Client, a MongoDB client for connecting to the database.
//   - passwordHasher: service.PasswordHasher, a service for hashing and verifying passwords.
//   - authorizeJwt: service.AuthenticateJWT, a service for generating and verifying JWT tokens.
//   - activityRepository: repository.ActivityRepository, a repository for storing and retrieving user activities.
//   - userRepository: repository.UserRepository, a repository for storing and retrieving user information.
//   - invoiceRepository: repository.InvoiceRepository, a repository for storing and retrieving invoice information.
//
// Returns:
//   - *Application, a pointer to a new Application instance with the provided dependencies.
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

// SignUpHandler handles the user registration process.
// It parses the request body, validates the input, hashes the password,
// creates a new user, and attempts to add the user to the database.
//
// Returns:
//   - fiber.Handler: A function that processes the request and returns an error if any occurs during the sign-up process.
func (app *Application) SignUpHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		data := new(SignUpRequestModel)
		print(data)

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
		user, err = app.userRepository.AddUser(app.db, user, user.Email)
		if err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "cannot add existing user",
				"message": fmt.Sprintf("%s: %q", ErrUserAlreadyExists.Error(), "go back to login"),
			})
		}

		go func() {
			activity := &domain.Activity{
				UserID:    user.ID,
				Action:    infra.UserCreatedAccountActivity,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"email": user.Email,
				},
			}
			if err := app.activityRepository.Save(app.db, activity); err != nil {
				slog.Error("Failed to record user activity", "error", err)
			}
		}()

		// If the user is created successfully
		return c.Status(fiber.StatusCreated).JSON(fiber.Map{
			"message": fmt.Sprintf("User: %s has been created successfully", user.ID),
		})
	}
}

// LoginHandler manages the user login process, including validation of input,
// verification of credentials, and generation of a JWT token.
//
// Returns:
//   - fiber.Handler: A function that processes the login request and returns an error if any occurs during the process.
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

		go func() {
			activity := &domain.Activity{
				UserID:    user.ID,
				Action:    infra.UserLoginActivity,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"email": user.Email,
				},
			}
			if err := app.activityRepository.Save(app.db, activity); err != nil {
				slog.Error("Failed to record user activity", "error", err)
			}
		}()

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
			"token":   token,
		})
	}
}

// CreateInvoiceHandler handles the creation of a new invoice for a user.
// It checks for authentication, validates input, and stores the invoice in the database.
//
// Returns:
//   - fiber.Handler: A function that processes the request and returns an error if any occurs during the creation process.
func (app *Application) CreateInvoiceHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := app.contextWithAuth(c, true); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   ErrUnauthorized.Error(),
				"message": "You are not authorized to perform this action",
			})
		}

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
			userID,
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
		res := app.invoiceRepository.AddNewInvoice(app.db, userID, invoice)
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
					"invoiceID":       invoice.InvoiceID,
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
			"message": fmt.Sprintf("Invoice: %s has been created successfully", invoice.InvoiceID),
		})
	}
}

// GetInvoiceHandler retrieves a specific invoice for a user.
// It checks for authentication, validates request parameters, and fetches the invoice from the database.
//
// Returns:
//   - fiber.Handler: A function that processes the request and returns an error if any occurs during the retrieval process.
func (app *Application) GetInvoiceHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := app.contextWithAuth(c, true); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   ErrUnauthorized.Error(),
				"message": "You are not authorized to perform this action",
			})
		}
		params := c.AllParams()
		if params == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID and invoiceID must be provided",
			})
		}
		userID := params["userID"]
		invoiceID := params["invoiceID"]
		if userID == "" || invoiceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID and invoiceID must be provided",
			})
		}

		invoice, err := app.invoiceRepository.FindUserInvoiceByID(app.db, userID, invoiceID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Invoice not found",
					"message": fmt.Sprintf("No invoice found with ID %s for user %s", invoiceID, userID),
				})
			}
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to retrieve invoice",
				"message": err.Error(),
			})
		}

		// recording user activity
		go func() {
			activity := &domain.Activity{
				UserID:    userID,
				Action:    infra.ViewInvoiceActivity,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"invoiceID":     invoiceID,
					"invoiceNumber": invoice.InvoiceNumber,
				},
			}
			if err := app.activityRepository.Save(app.db, activity); err != nil {
				slog.Error("Failed to record user activity", "error", err)
			}
		}()

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Invoice retrieved successfully",
			"data":    invoice,
		})
	}
}

// ListAllInvoice retrieves all invoices for a specific user.
// It checks for authentication, validates the user ID, and fetches the invoices from the database.
//
// Parameters:
//   - c: fiber.Ctx, the context for the current request, which includes request and response objects.
//
// Returns:
//   - fiber.Handler: A function that processes the request and returns an error if any occurs during the retrieval process.
func (app *Application) ListAllInvoiceHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// checking authentication
		if err := app.contextWithAuth(c, true); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   ErrUnauthorized.Error(),
				"message": "You are not authorized to perform this action",
			})
		}

		// get userID from params
		userID := c.Params("userID")
		if userID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID must be provided",
			})
		}

		// validate userID
		if _, err := primitive.ObjectIDFromHex(userID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid user ID",
				"message": "The provided userID is not a valid ObjectID",
			})
		}

		// get all invoices for the user
		invoices, err := app.invoiceRepository.FindAllInvoice(app.db, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to retrieve invoices",
				"message": err.Error(),
			})
		}

		// cecord user activity for this action
		go func() {
			activity := &domain.Activity{
				UserID:    userID,
				Action:    infra.ListInvoicesActivity,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"invoiceCount": len(invoices),
				},
			}
			if err := app.activityRepository.Save(app.db, activity); err != nil {
				slog.Error("Failed to record user activity", "error", err)
			}
		}()

		// return all the invoices
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Invoices retrieved successfully",
			"data":    invoices,
		})
	}
}

// UpdateUnIssuedInvoice handles the update of an unissued invoice for a specific user.
// It checks for authentication, validates input, and updates the invoice in the database.
//
// Parameters:
//   - c: fiber.Ctx, the context for the current request, which includes request and response objects.
//
// Returns:
//   - fiber.Handler: A function that processes the request and returns an error if any occurs during the update process.
func (app *Application) UpdateUnIssuedInvoiceHandler() fiber.Handler {

	return func(c *fiber.Ctx) error {
		// Check authentication
		if err := app.contextWithAuth(c, true); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   ErrUnauthorized.Error(),
				"message": "You are not authorized to perform this action",
			})
		}

		userID := c.Params("userID")
		invoiceID := c.Params("invoiceID")
		if userID == "" || invoiceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID and invoiceID must be provided",
			})
		}

		updatedInvoice := new(InvoiceRequestModel)
		if err := c.BodyParser(updatedInvoice); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   err.Error(),
				"message": "Invalid input received from the client",
			})
		}

		validateData := FieldValidator(updatedInvoice)
		if len(validateData) > 0 {
			for _, f := range validateData {
				return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
					"error":   "Invalid input",
					"message": fmt.Sprintf("%s: %s", f.Message, f.NameSpace),
				})
			}
		}

		// convert the updated invoice data to domain.Invoice
		items := make([]domain.Item, len(updatedInvoice.Items))
		for i, item := range updatedInvoice.Items {
			items[i] = domain.Item{
				Description: item.Description,
				Quantity:    item.Quantity,
				UnitPrice:   item.UnitPrice,
				TotalPrice:  item.TotalPrice,
			}
		}

		domainInvoice, err := domain.NewInvoice(
			userID,
			updatedInvoice.InvoiceNumber,
			updatedInvoice.BillingCurrency,
			updatedInvoice.Discount,
			updatedInvoice.IssueDate,
			updatedInvoice.DueDate,
			items,
			domain.PaymentInformation(updatedInvoice.PaymentInfo),
			domain.CustomerDetails(updatedInvoice.Customer),
			domain.SenderDetails(updatedInvoice.Sender),
		)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to create updated invoice",
				"message": err.Error(),
			})
		}

		// update the invoice
		err = app.invoiceRepository.UpdateInvoiceBeforeDueDate(app.db, userID, invoiceID, domainInvoice)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to update invoice",
				"message": err.Error(),
			})
		}

		// record user activity
		go func() {
			activity := &domain.Activity{
				UserID:    userID,
				Action:    infra.UpdateInvoiceActivity,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"invoiceID":     invoiceID,
					"invoiceNumber": domainInvoice.InvoiceNumber,
				},
			}
			if err := app.activityRepository.Save(app.db, activity); err != nil {
				slog.Error("Failed to record user activity", "error", err)
			}
		}()

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Invoice updated successfully",
			"data":    domainInvoice,
		})
	}
}

func (app *Application) GetUserInvoiceStatHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// checking authentication
		if err := app.contextWithAuth(c, true); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   ErrUnauthorized.Error(),
				"message": "You are not authorized to perform this action",
			})
		}

		userID := c.Params("userID")
		if userID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID must be provided",
			})
		}

		if _, err := primitive.ObjectIDFromHex(userID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid user ID",
				"message": "The provided userID is not a valid ObjectID",
			})
		}

		// get the invoice statistic aggregated value
		invoiceStatSummary, err := app.invoiceRepository.InvoiceStatSummary(app.db, userID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to get invoice statistic",
				"message": err.Error(),
			})
		}

		// return the invoice statistic summary
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Invoice statistic retrieved successfully",
			"data":    invoiceStatSummary,
		})
	}
}

func (app *Application) SendIssuedInvoiceToCustomer() fiber.Handler {
	return func(c *fiber.Ctx) error {
		data := new(UpdateInvoiceStatusRequestModel)
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

		// checking authentication
		if err := app.contextWithAuth(c, true); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   ErrUnauthorized.Error(),
				"message": "You are not authorized to perform this action",
			})
		}
		// get all the parameters
		params := c.AllParams()
		if params == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID and invoiceID must be provided",
			})
		}
		userID := params["userID"]
		invoiceID := params["invoiceID"]
		if userID == "" || invoiceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID and invoiceID must be provided",
			})
		}

		if _, err := primitive.ObjectIDFromHex(userID); err != nil {
			slog.Error("Invalid user", "Error", "the userID must be a valid UUID")
			panic(err)
		}

		if _, err := primitive.ObjectIDFromHex(invoiceID); err != nil {
			slog.Error("Invalid invoice", "Error", "the invoiceID must be a valid UUID")
			panic(err)
		}

		err := app.invoiceRepository.UpdateInvoiceStatusToIssued(app.db, userID, invoiceID)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to update invoice status",
				"message": err.Error(),
			})
		}

		go func() {
			activity := &domain.Activity{
				UserID:    userID,
				Action:    infra.IssueInvoiceActivity,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"invoiceID": invoiceID,
				},
			}
			if err := app.activityRepository.Save(app.db, activity); err != nil {
				slog.Error("Failed to record user activity", "error", err, "userID", userID, "action", infra.IssueInvoiceActivity)
			}
		}()

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Invoice status updated successfully",
		})

	}

}

func (app *Application) DeleteInvoiceHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		if err := app.contextWithAuth(c, true); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   ErrUnauthorized.Error(),
				"message": "You are not authorized to perform this action",
			})
		}

		params := c.AllParams()
		if params == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID and invoiceID must be provided",
			})
		}

		userID := params["userID"]
		invoiceID := params["invoiceID"]

		if userID == "" || invoiceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID and invoiceID must be provided",
			})
		}

		if _, err := primitive.ObjectIDFromHex(userID); err != nil {
			slog.Error("Invalid userID", "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid userID",
				"message": "userID must be a valid ObjectID",
			})
		}

		if _, err := primitive.ObjectIDFromHex(invoiceID); err != nil {
			slog.Error("Invalid invoiceID", "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid invoiceID",
				"message": "invoiceID must be a valid ObjectID",
			})
		}

		err := app.invoiceRepository.DeleteInvoice(app.db, userID, invoiceID)
		if err != nil {
			slog.Error("Failed to delete invoice", "userID", userID, "invoiceID", invoiceID, "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to delete invoice",
				"message": err.Error(),
			})
		}

		go func() {
			activity := &domain.Activity{
				UserID:    userID,
				Action:    infra.DeleteInvoiceActivity,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"invoiceID": invoiceID,
				},
			}
			if err := app.activityRepository.Save(app.db, activity); err != nil {
				slog.Error("Failed to record user activity", "error", err)
			}
		}()

		// Return success response
		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Invoice deleted successfully",
		})
	}
}

func (app *Application) DownloadInvoicePDFHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check authentication
		if err := app.contextWithAuth(c, true); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   ErrUnauthorized.Error(),
				"message": "You are not authorized to perform this action",
			})
		}

		// Get parameters from request
		params := c.AllParams()
		if params == nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID and invoiceID must be provided",
			})
		}
		userID := params["userID"]
		invoiceID := params["invoiceID"]

		// Validate user ID and invoice ID
		if userID == "" || invoiceID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID and invoiceID must be provided",
			})
		}

		if _, err := primitive.ObjectIDFromHex(userID); err != nil {
			slog.Error("Invalid userID", "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid userID",
				"message": "userID must be a valid ObjectID",
			})
		}

		if _, err := primitive.ObjectIDFromHex(invoiceID); err != nil {
			slog.Error("Invalid invoiceID", "error", err)
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid invoiceID",
				"message": "invoiceID must be a valid ObjectID",
			})
		}

		// Get the invoice data
		invoice, err := app.invoiceRepository.FindUserInvoiceByID(app.db, userID, invoiceID)
		if err != nil {
			if err == mongo.ErrNoDocuments {
				return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
					"error":   "Invoice not found",
					"message": "The specified invoice does not exist for the given user",
				})
			}
			slog.Error("Failed to retrieve invoice", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to retrieve invoice",
				"message": err.Error(),
			})
		}

		// Create temporary directory for PDF if it doesn't exist
		tempDir := "temp/invoices"
		if err := os.MkdirAll(tempDir, 0755); err != nil {
			slog.Error("Failed to create temp directory", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to generate invoice document",
				"message": err.Error(),
			})
		}

		// Generate PDF filename
		pdfFilename := fmt.Sprintf("invoice_%s_%s.pdf", userID, invoiceID)
		pdfPath := filepath.Join(tempDir, pdfFilename)

		// Generate the PDF
		if err := GenerateInvoicePDF(invoice, pdfPath); err != nil {
			slog.Error("Failed to generate PDF", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to generate invoice document",
				"message": err.Error(),
			})
		}

		// set response headers for file download
		c.Set("Content-Disposition", fmt.Sprintf(`attachment; filename="%s"`, pdfFilename))
		c.Set("Content-Type", "application/pdf")

		if err := c.SendFile(pdfPath); err != nil {
			slog.Error("Failed to send PDF file", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to send invoice document",
				"message": err.Error(),
			})
		}

		go func() {
			time.Sleep(5 * time.Minute)
			if err := os.Remove(pdfPath); err != nil {
				slog.Error("Failed to cleanup temporary PDF file", "error", err)
			}
		}()

		go func() {
			activity := &domain.Activity{
				UserID:    userID,
				Action:    infra.DownloadInvoiceActivity,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"invoice":   invoice,
					"file_path": pdfPath,
				},
			}
			if err := app.activityRepository.Save(app.db, activity); err != nil {
				slog.Error("Failed to record user activity", "error", err)
			}
		}()

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Invoice downloaded successfully",
		})

	}
}

// GetInvoiceActivitiesHandler returns a handler function that retrieves invoice activities for a specific user.
// It checks for authentication, validates the user ID, and fetches the activities from the database.
//
// Parameters:
//   - c: fiber.Ctx, the context for the current request, which includes request and response objects.
//
// Returns:
//   - fiber.Handler: A function that processes the request and returns an error if any occurs during the retrieval process.
func (app *Application) GetInvoiceActivitiesHandler() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Check authentication
		if err := app.contextWithAuth(c, true); err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error":   ErrUnauthorized.Error(),
				"message": "You are not authorized to perform this action",
			})
		}

		userID := c.Params("userID")
		if userID == "" {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid request parameters",
				"message": "userID must be provided",
			})
		}

		if _, err := primitive.ObjectIDFromHex(userID); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"error":   "Invalid user ID",
				"message": "The provided userID is not a valid ObjectID",
			})
		}

		// get limit from query params, default to 10 if not provided
		limit := int64(10)
		if limitStr := c.Query("limit"); limitStr != "" {
			parsedLimit, err := strconv.ParseInt(limitStr, 10, 64)
			if err == nil && parsedLimit > 0 {
				limit = parsedLimit
			}
		}

		activities, err := app.invoiceRepository.GetInvoiceActivities(app.db, userID, limit)
		if err != nil {
			slog.Error("Failed to retrieve invoice activities", "error", err)
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error":   "Failed to retrieve invoice activities",
				"message": err.Error(),
			})
		}

		go func() {
			activity := &domain.Activity{
				UserID:    userID,
				Action:    infra.ViewInvoiceActivity,
				Timestamp: time.Now(),
				Metadata: map[string]interface{}{
					"limit":           limit,
					"activitiesCount": len(activities),
				},
			}
			if err := app.activityRepository.Save(app.db, activity); err != nil {
				slog.Error("Failed to record user activity", "error", err)
			}
		}()

		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"message": "Invoice activities retrieved successfully",
			"data":    activities,
		})
	}
}
