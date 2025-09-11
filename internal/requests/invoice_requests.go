package requests

// GetCustomerInvoicesRequest represents a request to get invoices for a customer
type GetCustomerInvoicesRequest struct {
	CustomerID  string `query:"customerId" validate:"required"`
	InvoiceDate string `query:"invoiceDate"`
}

// GetInvoicesByIDsRequest represents a request to retrieve invoices by their IDs
type GetInvoicesByIDsRequest struct {
	InvoiceIDs []string `json:"invoiceIds" validate:"required,min=1"`
}

// InitiatePaymentRequest represents a request to initiate a payment for an invoice
type InitiatePaymentRequest struct {
	CustomerID string  `json:"customerId" validate:"required"`
	InvoiceID  string  `json:"invoiceId" validate:"required"`
	Amount     float64 `json:"amount" validate:"required,gt=0"`
}
