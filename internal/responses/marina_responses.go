package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/models"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// MarinaResponse represents a marina in the system
// @Description Marina data including location, contact information, and operational details
type MarinaResponse struct {
	ID             uuid.UUID             `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OrganizationID uuid.UUID             `json:"organizationId" example:"550e8400-e29b-41d4-a716-446655440001"`
	Name           string                `json:"name" example:"Harbor Bay Marina"`
	Email          string                `json:"email" example:"info@harborbay.com"`
	Location       *string               `json:"location,omitempty" example:"Miami Beach"`
	Phone          *string               `json:"phone,omitempty" example:"+15551234567"`
	Country        *string               `json:"country,omitempty" example:"USA"`
	Currency       *string               `json:"currency,omitempty" example:"USD"`
	WorkingHours   *models.WorkingHours  `json:"workingHours,omitempty"`
	Website        *string               `json:"website,omitempty" example:"https://harborbay.com"`
	Image          *string               `json:"image,omitempty" example:"/images/marinas/harborbay.jpg"`
	MaxUsers       *int32                `json:"maxUsers,omitempty" example:"100"`
	IsActive       *bool                 `json:"isActive,omitempty" example:"true"`
	IsTest         *bool                 `json:"isTest,omitempty" example:"false"`
	CreatedAt      *time.Time            `json:"createdAt,omitempty"`
	UpdatedAt      *time.Time            `json:"updatedAt,omitempty"`
	AddressID      uuid.UUID             `json:"addressId" example:"550e8400-e29b-41d4-a716-446655440003"`
	SystemID       *string               `json:"systemId,omitempty" example:"SYS123456"`
	StorageUsage   *utils.StorageUsageGB `json:"storageUsage,omitempty" example:"0.00"`
	EmailUsage     *int16                `json:"emailUsage,omitempty" example:"0"`
	TextUsage      *int16                `json:"textUsage,omitempty" example:"0"`
}

// MarinaWithAddressResponse represents a marina with its address details
// @Description Marina data with address details
type MarinaWithAddressResponse struct {
	MarinaResponse
	Address *AddressResponse `json:"address,omitempty"`
}

// ConvertMarinaToResponse converts a database marina to a response model
func ConvertMarinaToResponse(marina db.Marina) MarinaResponse {
	var workingHours *models.WorkingHours

	if marina.WorkingHours != nil {
		workingHours = &models.WorkingHours{}
		if err := workingHours.FromBytes(marina.WorkingHours); err != nil {
			// Log error but continue with nil working hours
			workingHours = nil
		}
	}

	// Convert storage usage from bytes to GB
	var storageUsageGB *utils.StorageUsageGB
	if marina.StorageUsage != nil {
		const bytesInGB = 1024 * 1024 * 1024 // 1 GB in bytes
		usageGB := utils.StorageUsageGB(float64(*marina.StorageUsage) / float64(bytesInGB))
		storageUsageGB = &usageGB
	}

	return MarinaResponse{
		ID:             marina.ID,
		OrganizationID: marina.OrganizationID,
		Name:           marina.Name,
		Email:          marina.Email,
		Location:       marina.Location,
		Phone:          marina.Phone,
		Country:        marina.Country,
		Currency:       marina.Currency,
		WorkingHours:   workingHours,
		Website:        marina.Website,
		Image:          utils.GetFullImageURL(marina.Image),
		MaxUsers:       marina.MaxUsers,
		IsActive:       marina.IsActive,
		IsTest:         marina.IsTest,
		CreatedAt:      utils.PgTimeToTimePtr(marina.CreatedAt),
		UpdatedAt:      utils.PgTimeToTimePtr(marina.UpdatedAt),
		AddressID:      marina.AddressID,
		SystemID:       marina.SystemID,
		StorageUsage:   storageUsageGB,
		EmailUsage:     marina.EmailUsage,
		TextUsage:      marina.TextUsage,
	}
}

// NewMarinaResponseSuccess creates a new successful marina response
func NewMarinaResponseSuccess(marina db.Marina) BaseResponse {
	return NewSuccessResponse(ConvertMarinaToResponse(marina))
}

// NewMarinaWithAddressResponse creates a new marina with address response
func NewMarinaWithAddressResponse(marina db.Marina, address *db.Address) BaseResponse {
	marinaResponse := ConvertMarinaToResponse(marina)
	response := MarinaWithAddressResponse{
		MarinaResponse: marinaResponse,
	}

	if address != nil {
		addrResponse := ConvertAddressToResponse(*address)
		response.Address = &addrResponse
	}

	return NewSuccessResponse(response)
}

// NewMarinasPaginatedResponse creates a paginated response of marinas
func NewMarinasPaginatedResponse(marinas []db.Marina, total int64, perPage, page int32) BaseResponse {
	marinaResponses := make([]MarinaResponse, len(marinas))
	for i, marina := range marinas {
		marinaResponses[i] = ConvertMarinaToResponse(marina)
	}

	return NewPaginatedResponse(marinaResponses, total, perPage, page)
}

// MarinaListResponse is purely for Swagger documentation
type MarinaListResponse struct {
	Data        []MarinaResponse `json:"data"`
	Total       int64            `json:"total" example:"42"`
	PerPage     int32            `json:"perPage" example:"10"`
	CurrentPage int32            `json:"currentPage" example:"1"`
	LastPage    int32            `json:"lastPage" example:"5"`
}

// MarinaResponseWrapper is purely for Swagger documentation
type MarinaResponseWrapper struct {
	Data MarinaResponse `json:"data"`
}

// MarinaWithAddressResponseWrapper is purely for Swagger documentation
type MarinaWithAddressResponseWrapper struct {
	Data MarinaWithAddressResponse `json:"data"`
}
