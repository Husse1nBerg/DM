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
