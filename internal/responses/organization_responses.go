package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// OrganizationResponse defines the response for organization data
type OrganizationResponse struct {
	ID        uuid.UUID  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	Email     string     `json:"email" example:"org@example.com"`
	Name      string     `json:"name" example:"Sample Organization"`
	Image     *string    `json:"image,omitempty" example:"https://example.com/logo.png"`
	Website   *string    `json:"website,omitempty" example:"https://example.com"`
	Country   *string    `json:"country,omitempty" example:"USA"`
	Phone     *string    `json:"phone,omitempty" example:"+1 555-123-4567"`
	IsActive  *bool      `json:"is_active,omitempty" example:"true"`
	IsTest    *bool      `json:"is_test,omitempty" example:"false"`
	CreatedAt *time.Time `json:"created_at,omitempty" example:"2023-01-01T00:00:00Z"`
	UpdatedAt *time.Time `json:"updated_at,omitempty" example:"2023-01-01T00:00:00Z"`
	AddressID uuid.UUID  `json:"address_id,omitempty" example:"123e4567-e89b-12d3-a456-426614174001"`
}

// OrganizationsResponse defines the response for a list of organizations
// @Description Response containing a list of organizations
type OrganizationsResponse struct {
	Data []OrganizationResponse `json:"data"`
}

// OrganizationsPaginatedResponse defines the paginated response for organizations
// @Description Paginated response containing a list of organizations
type OrganizationsPaginatedResponse struct {
	Data        []OrganizationResponse `json:"data"`
	Total       int64                  `json:"total" example:"100"`
	PerPage     int32                  `json:"perPage" example:"10"`
	CurrentPage int32                  `json:"currentPage" example:"1"`
	LastPage    int32                  `json:"lastPage" example:"10"`
}

// ConvertOrganizationToResponse converts a database organization to a response model
func ConvertOrganizationToResponse(org db.Organization) OrganizationResponse {
	return OrganizationResponse{
		ID:        org.ID,
		Email:     org.Email,
		Name:      org.Name,
		Image:     org.Image,
		Website:   org.Website,
		Country:   org.Country,
		Phone:     org.Phone,
		IsActive:  org.IsActive,
		IsTest:    org.IsTest,
		CreatedAt: utils.PgTimeToTimePtr(org.CreatedAt),
		UpdatedAt: utils.PgTimeToTimePtr(org.UpdatedAt),
		AddressID: org.AddressID,
	}
}

// ConvertOrganizationsToResponse converts a slice of database organizations to a response model
func ConvertOrganizationsToResponse(orgs []db.Organization) []OrganizationResponse {
	result := make([]OrganizationResponse, len(orgs))
	for i, org := range orgs {
		result[i] = ConvertOrganizationToResponse(org)
	}
	return result
}

// NewOrganizationResponseSuccess creates a successful response with organization data
func NewOrganizationResponseSuccess(org db.Organization) BaseResponse {
	return NewSuccessResponse(ConvertOrganizationToResponse(org))
}

// NewOrganizationsPaginatedResponse creates a paginated response for organizations
func NewOrganizationsPaginatedResponse(orgs []db.Organization, total int64, perPage, page int32) BaseResponse {
	return NewPaginatedResponse(ConvertOrganizationsToResponse(orgs), total, perPage, page)
}
