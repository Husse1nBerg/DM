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
	TaxRate                   float64    `json:"taxRate"`
	TaxEnabled                bool       `json:"taxEnabled"`
	TaxDescription            *string    `json:"taxDescription,omitempty"`
	IsActive                  bool       `json:"isActive"`
	CreatedAt                 time.Time  `json:"createdAt"`
	UpdatedAt                 *time.Time `json:"updatedAt,omitempty"`
	CreatedBy                 uuid.UUID  `json:"createdBy"`
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
	Tax                float64 `json:"tax"`
	TotalAmount        float64 `json:"totalAmount"`
	ConvenienceFeeType string  `json:"convenienceFeeType"`
	SurchargeType      string  `json:"surchargeType"`
	ConvenienceFeeRate float64 `json:"convenienceFeeRate"`
	SurchargeRate      float64 `json:"surchargeRate"`
	TaxRate            float64 `json:"taxRate"`
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
		TaxRate:               config.TaxRate,
		TaxEnabled:            config.TaxEnabled,
		IsActive:              config.IsActive,
		CreatedAt:             createdAt,
		CreatedBy:             config.CreatedBy,
	}

	// Convert optional fields
	if config.ConvenienceFeeDescription != nil {
		response.ConvenienceFeeDescription = config.ConvenienceFeeDescription
	}
	if config.SurchargeDescription != nil {
		response.SurchargeDescription = config.SurchargeDescription
	}
	if config.TaxDescription != nil {
		response.TaxDescription = config.TaxDescription
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
