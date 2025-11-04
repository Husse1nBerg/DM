package requests

import "github.com/google/uuid"

// GetCustomerInvoicesRequest represents a request to get invoices for a customer
type GetCustomerInvoicesRequest struct {
	CustomerID  string    `query:"customerId" validate:"required"`
	InvoiceDate string    `query:"invoiceDate"`
	MarinaID    uuid.UUID `query:"marinaId" validate:"required"`
	Token       string    `query:"token"` // Optional: short-lived payment token
}

// GetNextReferenceRequest represents a request to get next AR reference number for a customer
type GetNextReferenceRequest struct {
	CustomerID string `query:"customerId" validate:"required"`
}
