package requests

import "github.com/google/uuid"

// GetCustomerInvoicesRequest represents a request to get invoices for a customer
type GetCustomerInvoicesRequest struct {
	CustomerID  string    `query:"customerId" validate:"omitempty"`
	InvoiceDate string    `query:"invoiceDate"`
	MarinaID    uuid.UUID `query:"marinaId" validate:"omitempty"`
	Token       string    `query:"token"` // Optional: short-lived payment token
}

// GetNextReferenceRequest represents a request to get next AR reference number for a customer
type GetNextReferenceRequest struct {
	CustomerID string `query:"customerId" validate:"required"`
}

// GetCustomerInvoiceRequest represents a request to get a specific customer invoice
type GetCustomerInvoiceRequest struct {
	InvoiceID  string    `query:"invoiceId" validate:"required"`
	CustomerID string    `query:"customerId" validate:"omitempty"`
	MarinaID   uuid.UUID `query:"marinaId" validate:"omitempty"`
	Token      string    `query:"token" validate:"omitempty"`
}
