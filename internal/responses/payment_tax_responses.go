package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// PaymentTaxResponse represents a payment tax configuration record
type PaymentTaxResponse struct {
	ID                        uuid.UUID  `json:"id"`
	MarinaID                  uuid.UUID  `json:"marinaId"`
	ConvenienceFee            float64    `json:"convenienceFee"`
	ConvenienceFeeType        string     `json:"convenienceFeeType"`
	ConvenienceFeeEnabled     bool       `json:"convenienceFeeEnabled"`
	ConvenienceFeeDescription *string    `json:"convenienceFeeDescription,omitempty"`
	Surcharge                 float64    `json:"surcharge"`
	SurchargeType             string     `json:"surchargeType"`
	SurchargeEnabled          bool       `json:"surchargeEnabled"`
	SurchargeDescription      *string    `json:"surchargeDescription,omitempty"`
	PaymentType               string     `json:"paymentType"`
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 *time.Time `json:"updatedAt,omitempty"`
}

// PaymentTaxListResponse represents a paginated list of payment tax configurations
type PaymentTaxListResponse struct {
	Data       []PaymentTaxResponse `json:"data"`
	Total      int64                `json:"total"`
	Page       int                  `json:"page"`
	PageSize   int                  `json:"pageSize"`
	TotalPages int                  `json:"totalPages"`
}

// FeeCalculationResponse represents the calculated fees for a given amount
type FeeCalculationResponse struct {
	BaseAmount         float64 `json:"baseAmount"`
	ConvenienceFee     float64 `json:"convenienceFee"`
	Surcharge          float64 `json:"surcharge"`
	TotalAmount        float64 `json:"totalAmount"`
	ConvenienceFeeType string  `json:"convenienceFeeType"`
	SurchargeType      string  `json:"surchargeType"`
	ConvenienceFeeRate float64 `json:"convenienceFeeRate"`
	SurchargeRate      float64 `json:"surchargeRate"`
	PaymentType        string  `json:"paymentType"`
}

// ConvertPaymentTaxToResponse converts a db.TaxConfiguration to PaymentTaxResponse
func ConvertPaymentTaxToResponse(config db.TaxConfiguration) PaymentTaxResponse {
	// Convert CreatedAt timestamp
	var createdAt time.Time
	if config.CreatedAt.Valid {
		createdAt = config.CreatedAt.Time
	}

	// Convert numeric fields to float64
	convenienceFee, _ := utils.NumericToFloat64(config.ConvenienceFee)
	surcharge, _ := utils.NumericToFloat64(config.Surcharge)

	response := PaymentTaxResponse{
		ID:                    config.ID,
		MarinaID:              config.MarinaID,
		ConvenienceFee:        convenienceFee,
		ConvenienceFeeType:    config.ConvenienceFeeType,
		ConvenienceFeeEnabled: config.ConvenienceFeeEnabled,
		Surcharge:             surcharge,
		SurchargeType:         config.SurchargeType,
		SurchargeEnabled:      config.SurchargeEnabled,
		PaymentType:           config.PaymentType,
		CreatedAt:             createdAt,
	}

	// Convert optional fields
	if config.ConvenienceFeeDescription != nil {
		response.ConvenienceFeeDescription = config.ConvenienceFeeDescription
	}
	if config.SurchargeDescription != nil {
		response.SurchargeDescription = config.SurchargeDescription
	}

	// Convert timestamp fields
	if config.UpdatedAt.Valid {
		t := config.UpdatedAt.Time
		response.UpdatedAt = &t
	}

	return response
}

// swag:response PaymentTaxResponse
type SwagPaymentTaxResponse = PaymentTaxResponse

// swag:response PaymentTaxListResponse
type SwagPaymentTaxListResponse = PaymentTaxListResponse

// swag:response FeeCalculationResponse
type SwagFeeCalculationResponse = FeeCalculationResponse
