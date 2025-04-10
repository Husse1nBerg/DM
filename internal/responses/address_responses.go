package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// AddressResponse defines the response for address data
// @Description Address response model
// @Schema responses.AddressResponse
type AddressResponse struct {
	ID         uuid.UUID  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Street     *string    `json:"street,omitempty" example:"123 Main St"`
	City       *string    `json:"city,omitempty" example:"San Francisco"`
	State      *string    `json:"state,omitempty" example:"CA"`
	PostalCode *string    `json:"postal_code,omitempty" example:"94105"`
	Country    *string    `json:"country,omitempty" example:"USA"`
	Latitude   *float64   `json:"latitude,omitempty" example:"37.7749"`
	Longitude  *float64   `json:"longitude,omitempty" example:"-122.4194"`
	CreatedAt  *time.Time `json:"created_at,omitempty" example:"2023-01-01T00:00:00Z"`
	UpdatedAt  *time.Time `json:"updated_at,omitempty" example:"2023-01-01T00:00:00Z"`
}

// OrganizationWithAddressResponse extends the organization response to include the address
// @Description Organization with address response model
// @Schema responses.OrganizationWithAddressResponse
type OrganizationWithAddressResponse struct {
	ID        uuid.UUID        `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email     string           `json:"email" example:"org@example.com"`
	Name      string           `json:"name" example:"Example Organization"`
	Image     *string          `json:"image,omitempty" example:"https://example.com/logo.png"`
	Website   *string          `json:"website,omitempty" example:"https://example.com"`
	Country   *string          `json:"country,omitempty" example:"USA"`
	Phone     *string          `json:"phone,omitempty" example:"+1-555-123-4567"`
	IsActive  *bool            `json:"is_active,omitempty" example:"true"`
	IsTest    *bool            `json:"is_test,omitempty" example:"false"`
	Address   *AddressResponse `json:"address,omitempty"`
	CreatedAt *time.Time       `json:"created_at,omitempty" example:"2023-01-01T00:00:00Z"`
	UpdatedAt *time.Time       `json:"updated_at,omitempty" example:"2023-01-01T00:00:00Z"`
}

// ConvertAddressToResponse converts a database address to a response model
func ConvertAddressToResponse(address db.Address) AddressResponse {
	// Prepare latitude and longitude pointers
	var latitude, longitude *float64

	// Copy values to prevent reference issues
	lat := address.Latitude
	lng := address.Longitude

	// Assign pointers
	latitude = &lat
	longitude = &lng

	return AddressResponse{
		ID:         address.ID,
		Street:     address.Street,
		City:       address.City,
		State:      address.State,
		PostalCode: address.PostalCode,
		Country:    address.Country,
		Latitude:   latitude,
		Longitude:  longitude,
		CreatedAt:  utils.PgTimeToTimePtr(address.CreatedAt),
		UpdatedAt:  utils.PgTimeToTimePtr(address.UpdatedAt),
	}
}

// NewAddressResponseSuccess creates a successful response with address data
func NewAddressResponseSuccess(address db.Address) BaseResponse {
	return NewSuccessResponse(ConvertAddressToResponse(address))
}
