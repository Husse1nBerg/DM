package requests

import (
	"github.com/dockworks/dm-web-backend/pkg/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CreateMarinaRequest represents the required parameters to create a new marina
type CreateMarinaRequest struct {
	OrganizationID uuid.UUID             `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Name           string                `json:"name" validate:"required" example:"Harbor Bay Marina"`
	Email          string                `json:"email" validate:"required,email" example:"info@harborbay.com"`
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
	Address        *CreateAddressRequest `json:"address,omitempty"`
}

// UpdateMarinaRequest represents the parameters that can be updated for a marina
type UpdateMarinaRequest struct {
	Name         *string               `json:"name,omitempty" example:"Harbor Bay Marina"`
	Email        *string               `json:"email,omitempty" validate:"omitempty,email" example:"info@harborbay.com"`
	Location     *string               `json:"location,omitempty" example:"Miami Beach"`
	Phone        *string               `json:"phone,omitempty" example:"+15551234567"`
	Country      *string               `json:"country,omitempty" example:"USA"`
	Currency     *string               `json:"currency,omitempty" example:"USD"`
	WorkingHours *models.WorkingHours  `json:"workingHours,omitempty"`
	Website      *string               `json:"website,omitempty" example:"https://harborbay.com"`
	Image        *string               `json:"image,omitempty" example:"/images/marinas/harborbay.jpg"`
	MaxUsers     *int32                `json:"maxUsers,omitempty" example:"100"`
	IsActive     *bool                 `json:"isActive,omitempty" example:"true"`
	IsTest       *bool                 `json:"isTest,omitempty" example:"false"`
	Address      *UpdateAddressRequest `json:"address,omitempty"`
}

// GetMarinasPaginatedRequest represents parameters for paginated marina retrieval
type GetMarinasPaginatedRequest struct {
	Limit  int32 `query:"limit" validate:"gte=0" default:"10"`
	Offset int32 `query:"offset" validate:"gte=0" default:"0"`
}

// Validate performs custom validation on the create marina request
func (r *CreateMarinaRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the update marina request
func (r *UpdateMarinaRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the pagination request
func (r *GetMarinasPaginatedRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
