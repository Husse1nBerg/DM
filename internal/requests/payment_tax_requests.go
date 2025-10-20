package requests

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CreatePaymentTaxRequest represents the request to create a payment tax configuration
type CreatePaymentTaxRequest struct {
	MarinaID                  uuid.UUID `json:"marinaId" validate:"required,uuid"`
	ConvenienceFee            float64   `json:"convenienceFee" validate:"required,min=0"`
	ConvenienceFeeType        string    `json:"convenienceFeeType" validate:"required,oneof=percentage fixed"`
	ConvenienceFeeEnabled     bool      `json:"convenienceFeeEnabled"`
	ConvenienceFeeDescription *string   `json:"convenienceFeeDescription"`
	Surcharge                 float64   `json:"surcharge" validate:"required,min=0"`
	SurchargeType             string    `json:"surchargeType" validate:"required,oneof=percentage fixed"`
	SurchargeEnabled          bool      `json:"surchargeEnabled"`
	SurchargeDescription      *string   `json:"surchargeDescription"`
	PaymentType               string    `json:"paymentType" validate:"required,oneof=CC DB CK ACH"`
}

// UpdatePaymentTaxRequest represents the request to update a payment tax configuration
type UpdatePaymentTaxRequest struct {
	ConvenienceFee            *float64 `json:"convenienceFee" validate:"omitempty,min=0"`
	ConvenienceFeeType        *string  `json:"convenienceFeeType" validate:"omitempty,oneof=percentage fixed"`
	ConvenienceFeeEnabled     *bool    `json:"convenienceFeeEnabled"`
	ConvenienceFeeDescription *string  `json:"convenienceFeeDescription"`
	Surcharge                 *float64 `json:"surcharge" validate:"omitempty,min=0"`
	SurchargeType             *string  `json:"surchargeType" validate:"omitempty,oneof=percentage fixed"`
	SurchargeEnabled          *bool    `json:"surchargeEnabled"`
	SurchargeDescription      *string  `json:"surchargeDescription"`
	PaymentType               *string  `json:"paymentType" validate:"omitempty,oneof=CC DB CK ACH"`
}

// GetPaymentTaxRequest represents the request to get a payment tax configuration by ID
type GetPaymentTaxRequest struct {
	ID string `param:"id" validate:"required,uuid"`
}

// ListPaymentTaxRequest represents the request to list payment tax configurations
type ListPaymentTaxRequest struct {
	Page     int `query:"page" validate:"required,min=1"`
	PageSize int `query:"pageSize" validate:"required,min=1,max=100"`
}

// CalculateFeeRequest represents the request to calculate fees for an amount
type CalculateFeeRequest struct {
	MarinaID    uuid.UUID `json:"marinaId" validate:"required,uuid"`
	PaymentType string    `json:"paymentType" validate:"required,oneof=CC DB CK ACH"`
	Amount      float64   `json:"amount" validate:"required,min=0"`
}

// Validate validates the CreatePaymentTaxRequest
func (r *CreatePaymentTaxRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// Validate validates the UpdatePaymentTaxRequest
func (r *UpdatePaymentTaxRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// Validate validates the GetPaymentTaxRequest
func (r *GetPaymentTaxRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// Validate validates the ListPaymentTaxRequest
func (r *ListPaymentTaxRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}

// Validate validates the CalculateFeeRequest
func (r *CalculateFeeRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}
