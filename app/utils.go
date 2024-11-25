package app

import (
	"fmt"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translate "github.com/go-playground/validator/v10/translations/en"
	"github.com/gofiber/fiber/v2"
)

// FieldValidator creates a contextual validation of all the struct fields
// and their struct tags, returning validation errors if any.
func FieldValidator(s any) []FieldResult {
	validate := validator.New(validator.WithRequiredStructEnabled())
	resultErr := make([]FieldResult, 0)

	// Setting English translator
	eng := en.New()
	translate := ut.New(eng, eng)
	translator, ok := translate.GetTranslator("en")

	if !ok {
		fmt.Println("Unable to set English translator")
		return []FieldResult{}
	}

	err := en_translate.RegisterDefaultTranslations(validate, translator)
	if err != nil {
		fmt.Println("Cannot register default translation")
		return []FieldResult{}
	}

	err = validate.Struct(s)
	if err != nil {
		errs := err.(validator.ValidationErrors)
		for _, e := range errs {
			resultErr = append(resultErr, FieldErrorsChecker(e))
		}
		return resultErr
	}
	return nil
}

// FieldResult is a struct that holds the result of the validation of a single field
type FieldResult struct {
	NameSpace, Message string
}

// FieldErrorsChecker checks the error message and returns the result of the validation
func FieldErrorsChecker(err validator.FieldError) FieldResult {
	param := err.Param()
	var msg string
	switch err.Tag() {
	case "required":
		msg = fmt.Sprintf("this %s is required", param)
	case "min":
		msg = fmt.Sprintf("the minimum length is %s", param)
	case "max":
		msg = fmt.Sprintf("the maximum length is %s", param)
	}
	return FieldResult{
		NameSpace: err.Namespace(),
		Message:   msg,
	}
}

// RespondWithError is a helper function to respond with error in Fiber
func RespondWithError(c *fiber.Ctx, status int, message string, err error) error {
	return c.Status(status).JSON(fiber.Map{
		"error":   err.Error(),
		"message": message,
	})
}
