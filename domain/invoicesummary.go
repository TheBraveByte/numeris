package domain

type InvoiceSummary struct {
	TotalPaid    float64
	TotalOverdue float64
	TotalDraft   float64
	TotalUnpaid  float64
}
