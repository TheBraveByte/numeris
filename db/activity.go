package infra

const (
	UserCreatedAccountActivity string = "user_created_account"
	UserLoginActivity          string = "user_login_activity"
	CreateInvoiceActivity      string = "create_invoice_activity"
	ViewInvoiceActivity        string = "view_invoice_activity"
	ListInvoicesActivity       string = "list_invoices_activity"
	UpdateInvoiceActivity      string = "update_invoice_activity"

	IssueInvoiceActivity  string = "issue_invoice_activity"
	DeleteInvoiceActivity string = "delete_invoice_activity"

	DownloadInvoiceActivity string = "download_invoice_activity"

	UserUpdatedAccountActivity string = "user_updated_account"

	InvoiceReminderActivity  string = "invoice_reminder_activity"
	InvoicePaidActivity      string = "invoice_paid_activity"
	InvoiceCancelledActivity string = "invoice_cancelled_activity"

	InvoiceRefundedActivity string = "invoice_refunded_activity"

	// i dont need this now
	// PaymentFailedActivity    string = "payment_failed_activity"
	// PaymentMadeActivity        string = "payment_made_activity"
)
