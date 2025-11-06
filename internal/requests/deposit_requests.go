package requests

import (
	"github.com/adyen/adyen-go-api-library/v14/src/checkout"
	"github.com/go-playground/validator/v10"
)

// CreateDrystackDepositSessionRequest represents a request to create an Adyen session for a Drystack Deposit
type CreateDrystackDepositSessionRequest struct {
	// Payment session fields
	Amount      int64               `json:"amount" validate:"required,min=1"`
	Currency    string              `json:"currency" validate:"required,len=3"`
	CountryCode string              `json:"country_code" validate:"required,len=2"`
	ReturnURL   string              `json:"return_url" validate:"required,url"`
	ShopperIP   string              `json:"shopper_ip,omitempty"`
	LineItems   []checkout.LineItem `json:"line_items,omitempty"`

	// Context
	MarinaID     string `json:"marinaId" validate:"required,uuid4"`
	CustomerID   string `json:"customerId" validate:"required"`
	LocationCode string `json:"locationCode" validate:"required"`

	// Drystack specific
	AgreementNum string `json:"agreementNum" validate:"omitempty"`
	Year         string `json:"year,omitempty"`
	BoatID       string `json:"boatId" validate:"required"`

	// Optional short-lived token (payment link flow)
	Token string `json:"token,omitempty"`
}

func (r *CreateDrystackDepositSessionRequest) Validate(validate *validator.Validate) error {
	return validate.Struct(r)
}
