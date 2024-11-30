package domain

type InvoiceSummary struct {
	TotalPaid    float64
	TotalOverdue float64
	TotalPending float64
	TotalUnpaid  float64
}
