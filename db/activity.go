package infra

const (
	UserCreatedAccountActivity string = "user_created_account"
	CreateInvoiceActivity      string = "create_invoice_activity"
	IssueInvoiceActivity       string = "issue_invoice_activity"
	PaymentMadeActivity        string = "payment_made_activity"
	UserUpdatedAccountActivity string = "user_updated_account"
	UserDeletedAccountActivity string = "user_deleted_account"
	InvoiceReminderActivity    string = "invoice_reminder_activity"
	PaymentFailedActivity      string = "payment_failed_activity"
	InvoicePaidActivity        string = "invoice_paid_activity"
	InvoiceCancelledActivity   string = "invoice_cancelled_activity"
	InvoiceRefundedActivity    string = "invoice_refunded_activity"
)
