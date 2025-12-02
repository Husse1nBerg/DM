package requests

import (
	"github.com/adyen/adyen-go-api-library/v14/src/checkout"
	"github.com/go-playground/validator/v10"
)

// CreatePaymentSessionRequest represents the request to create a payment session
type CreatePaymentSessionRequest struct {
	Amount      int64               `json:"amount" validate:"required,min=1"`
	Currency    string              `json:"currency" validate:"required,len=3"`
	CountryCode string              `json:"country_code" validate:"required,len=2"`
	ReturnURL   string              `json:"return_url" validate:"required,url"`
	ShopperIP   string              `json:"shopper_ip,omitempty"`
	LineItems   []checkout.LineItem `json:"line_items,omitempty"`
	Metadata    *map[string]string  `json:"metadata,omitempty"`
	Token       string              `json:"token,omitempty"` // Optional short-lived payment token
}

// PaymentDetailsRequest represents the request to handle payment details
type PaymentDetailsRequest struct {
	PaymentData    string `json:"payment_data" validate:"required"`
	RedirectResult string `json:"redirect_result,omitempty"`
	Payload        string `json:"payload,omitempty"`
}

// Validate validates the CreatePaymentSessionRequest
func (r *CreatePaymentSessionRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// Validate validates the PaymentDetailsRequest
func (r *PaymentDetailsRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// CreatePaymentLinkRequest represents a request to generate a short-lived payment link
type CreatePaymentLinkRequest struct {
	MarinaID      string `json:"marinaId" validate:"required,uuid4"`
	CustomerID    string `json:"customerId" validate:"required"`
	TTLMinutes    int    `json:"ttlMinutes" validate:"omitempty,min=5,max=4320"`
	Email         string `json:"email" validate:"omitempty,email"`
	Recipient     string `json:"recipient" validate:"omitempty"`
	Name          string `json:"name" validate:"omitempty"`
	ReplyName     string `json:"replyName" validate:"omitempty"`
	CustomMessage string `json:"customMessage" validate:"omitempty"`
	InvoiceID     string `json:"invoiceId" validate:"omitempty"`
	Amount        string `json:"amount" validate:"omitempty"`
}

// (SendPaymentLinkEmailRequest removed; use CreatePaymentLinkRequest with optional email fields)

// Validate validates the CreatePaymentLinkRequest
func (r *CreatePaymentLinkRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// TokenInvoicesRequest represents a request to fetch invoices using a token
type TokenInvoicesRequest struct {
	Token string `query:"token" validate:"required"`
}

// Validate validates the TokenInvoicesRequest
func (r *TokenInvoicesRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// ValidatePaymentTokenRequest represents a request to validate a payment token
type ValidatePaymentTokenRequest struct {
	Token string `query:"token" validate:"required"`
}

// Validate validates the ValidatePaymentTokenRequest
func (r *ValidatePaymentTokenRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// CreatePaymentSessionWithTokenRequest represents creating an Adyen session using a token
type CreatePaymentSessionWithTokenRequest struct {
	Token       string              `json:"token" validate:"required"`
	Amount      int64               `json:"amount" validate:"required,min=1"`
	Currency    string              `json:"currency" validate:"required,len=3"`
	CountryCode string              `json:"country_code" validate:"required,len=2"`
	ReturnURL   string              `json:"return_url" validate:"required,url"`
	ShopperIP   string              `json:"shopper_ip,omitempty"`
	LineItems   []checkout.LineItem `json:"line_items,omitempty"`
	Metadata    *map[string]string  `json:"metadata,omitempty"`
}

// Validate validates the CreatePaymentSessionWithTokenRequest
func (r *CreatePaymentSessionWithTokenRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// ListPaymentsRequest represents the request to list payments with filtering
type ListPaymentsRequest struct {
	Page       int    `query:"page" validate:"required,min=1"`
	PageSize   int    `query:"pageSize" validate:"required,min=1,max=100"`
	Status     string `query:"status"`     // Optional: pending, authorized, completed, failed
	EntityType string `query:"entityType"` // Optional: invoice, boat, customer, etc.
	EntityID   string `query:"entityId"`   // Optional: ID of the entity
	StartDate  string `query:"startDate"`  // Optional: YYYY-MM-DD format
	EndDate    string `query:"endDate"`    // Optional: YYYY-MM-DD format
}

// GetPaymentRequest represents the request to get a single payment
type GetPaymentRequest struct {
	PaymentID string `param:"id" validate:"required,uuid"`
}

// GetPaymentsByEntityRequest represents the request to get payments for a specific entity
type GetPaymentsByEntityRequest struct {
	EntityType string `query:"entityType" validate:"required"`
	EntityID   string `query:"entityId" validate:"required"`
}

// GetPaymentStatsRequest represents the request to get payment statistics
type GetPaymentStatsRequest struct {
	StartDate string `query:"startDate"` // Optional: YYYY-MM-DD format
	EndDate   string `query:"endDate"`   // Optional: YYYY-MM-DD format
}
