package app

import (
	"fmt"
	"time"

	"github.com/go-playground/locales/en"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	en_translate "github.com/go-playground/validator/v10/translations/en"
	"github.com/jung-kurt/gofpdf"

	"github.com/thebravebyte/numeris/domain"
)

const inputDateFormat = "2006-01-02"
const outputDateFormat = "Jan 02, 2006"

// FieldValidator performs validation on the provided struct using the validator package.
// It creates a new validator instance, sets up an English translator, and validates
// the struct fields based on their struct tags.
//
// Parameters:
//   - s: any - The struct to be validated. It should be a pointer to a struct.
//
// Returns:
//   - []FieldResult: A slice of FieldResult structs containing validation errors, if any.
//     Each FieldResult includes the namespace of the field and the error message.
//     Returns nil if no validation errors are found.
func FieldValidator(s any) []FieldResult {
	validate := validator.New(validator.WithRequiredStructEnabled())
	resultErr := make([]FieldResult, 0)

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

// FieldErrorsChecker processes a validator.FieldError and generates a corresponding FieldResult.
// It interprets the error tag and creates an appropriate error message.
//
// Parameters:
//   - err: validator.FieldError - The validation error to be processed.
//
// Returns:
//   - FieldResult: A struct containing the namespace of the field and the generated error message.
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

// GenerateInvoicePDF creates a PDF file for the given invoice data.
// It formats and writes all the invoice details including sender and customer information,
// invoice items, total amount, and payment information to the PDF.
//
// Parameters:
//   - invoice: *domain.Invoice - A pointer to the Invoice struct containing all the invoice data.
//   - filePath: string - The file path where the generated PDF will be saved.
//
// Returns:
//   - error: An error if the PDF generation or saving process fails, nil otherwise.
func GenerateInvoicePDF(invoice *domain.Invoice, filePath string) error {
	pdf := gofpdf.New("P", "mm", "A4", "")
	pdf.SetMargins(10, 10, 10)
	pdf.AddPage()

	pdf.SetFont("Arial", "B", 16)
	pdf.SetTextColor(33, 37, 41)
	pdf.CellFormat(0, 10, fmt.Sprintf("Invoice #%s", invoice.InvoiceNumber), "0", 1, "C", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, "Sender:", "0", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, fmt.Sprintf("%s\n%s\n%s", invoice.Sender.Name, invoice.Sender.Address, invoice.Sender.Email), "", "L", false)
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, "Customer:", "0", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, fmt.Sprintf("%s\n%s\n%s", invoice.Customer.Name, invoice.Customer.Address, invoice.Customer.Email), "", "L", false)
	pdf.Ln(10)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, "Invoice Details:", "0", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	issueDate, err := time.Parse(inputDateFormat, invoice.IssueDate)
	if err != nil {
		return fmt.Errorf("invalid issue date format: %v", err)
	}

	dueDate, err := time.Parse(inputDateFormat, invoice.DueDate)
	if err != nil {
		return fmt.Errorf("invalid due date format: %v", err)
	}

	pdf.CellFormat(0, 6, fmt.Sprintf("Issue Date: %s", issueDate.Format(outputDateFormat)), "0", 1, "L", false, 0, "")
	pdf.CellFormat(0, 6, fmt.Sprintf("Due Date: %s", dueDate.Format(outputDateFormat)), "0", 1, "L", false, 0, "")

	pdf.CellFormat(0, 6, fmt.Sprintf("Billing Currency: %s", invoice.BillingCurrency), "0", 1, "L", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 12)
	pdf.SetFillColor(220, 220, 220)
	pdf.CellFormat(80, 8, "Description", "1", 0, "C", true, 0, "")
	pdf.CellFormat(30, 8, "Quantity", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Unit Price", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 8, "Total Price", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 11)
	for _, item := range invoice.Items {
		pdf.CellFormat(80, 8, item.Description, "1", 0, "L", false, 0, "")
		pdf.CellFormat(30, 8, fmt.Sprintf("%d", item.Quantity), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", item.UnitPrice), "1", 0, "R", false, 0, "")
		pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", item.TotalPrice), "1", 1, "R", false, 0, "")
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.Ln(5)
	pdf.CellFormat(150, 8, "Total Amount Due:", "0", 0, "R", false, 0, "")
	pdf.CellFormat(40, 8, fmt.Sprintf("%.2f", invoice.TotalAmountDue), "1", 1, "R", false, 0, "")
	pdf.Ln(5)

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(0, 8, "Payment Information:", "0", 1, "L", false, 0, "")
	pdf.SetFont("Arial", "", 11)
	pdf.MultiCell(0, 6, fmt.Sprintf("Account Name: %s\nAccount Number: %s\nRouting Number: %s\nBank Name: %s",
		invoice.PaymentInfo.AccountName, invoice.PaymentInfo.AccountNumber, invoice.PaymentInfo.RoutingNumber, invoice.PaymentInfo.BankName), "", "L", false)
	pdf.Ln(5)

	if invoice.Notes != "" {
		pdf.Ln(10)
		pdf.SetFont("Arial", "B", 12)
		pdf.CellFormat(0, 8, "Notes:", "0", 1, "L", false, 0, "")
		pdf.SetFont("Arial", "", 11)
		pdf.MultiCell(0, 6, invoice.Notes, "", "L", false)
	}

	err = pdf.OutputFileAndClose(filePath)
	if err != nil {
		return fmt.Errorf("failed to save PDF: %w", err)
	}

	return nil
}
