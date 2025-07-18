package requests

import (
	"github.com/dockworks/dm-web-backend/pkg/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CreateMarinaRequest represents the required parameters to create a new marina
type CreateMarinaRequest struct {
	OrganizationID      uuid.UUID             `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Name                string                `json:"name" validate:"required" example:"Harbor Bay Marina"`
	Email               string                `json:"email" validate:"required,email" example:"info@harborbay.com"`
	Location            *string               `json:"location,omitempty" example:"Miami Beach"`
	Phone               *string               `json:"phone,omitempty" example:"+15551234567"`
	Country             *string               `json:"country,omitempty" example:"USA"`
	Currency            *string               `json:"currency,omitempty" example:"USD"`
	WorkingHours        *models.WorkingHours  `json:"workingHours,omitempty"`
	Website             *string               `json:"website,omitempty" example:"https://harborbay.com"`
	Image               *string               `json:"image,omitempty" example:"/images/marinas/harborbay.jpg"`
	MaxUsers            *int32                `json:"maxUsers,omitempty" example:"100"`
	IsActive            *bool                 `json:"isActive,omitempty" example:"true"`
	IsTest              *bool                 `json:"isTest,omitempty" example:"false"`
	Address             *CreateAddressRequest `json:"address,omitempty"`
	SystemID            *string               `json:"systemId,omitempty" example:"SYS123456"`
	NotesMessagesPlanID *uuid.UUID            `json:"notesMessagesPlanId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID of the notes and messages plan for this marina"`
	StoragePlanID       *uuid.UUID            `json:"storagePlanId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID of the storage plan for this marina"`
	DocumentPlanID      *uuid.UUID            `json:"documentPlanId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID of the document plan for this marina"`
	Modules             *models.Modules       `json:"modules,omitempty"`
}

// UpdateMarinaRequest represents the parameters that can be updated for a marina
type UpdateMarinaRequest struct {
	Name                 *string               `json:"name,omitempty" example:"Harbor Bay Marina"`
	Email                *string               `json:"email,omitempty" validate:"omitempty,email" example:"info@harborbay.com"`
	Location             *string               `json:"location,omitempty" example:"Miami Beach"`
	Phone                *string               `json:"phone,omitempty" example:"+15551234567"`
	Country              *string               `json:"country,omitempty" example:"USA"`
	Currency             *string               `json:"currency,omitempty" example:"USD"`
	WorkingHours         *models.WorkingHours  `json:"workingHours,omitempty"`
	Website              *string               `json:"website,omitempty" example:"https://harborbay.com"`
	Image                *string               `json:"image,omitempty" example:"/images/marinas/harborbay.jpg"`
	MaxUsers             *int32                `json:"maxUsers,omitempty" example:"100"`
	IsActive             *bool                 `json:"isActive,omitempty" example:"true"`
	IsTest               *bool                 `json:"isTest,omitempty" example:"false"`
	Address              *UpdateAddressRequest `json:"address,omitempty"`
	SystemID             *string               `json:"systemId,omitempty" example:"SYS123456"`
	NotesMessagesPlanID  *uuid.UUID            `json:"notesMessagesPlanId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID of the notes and messages plan for this marina"`
	StoragePlanID        *uuid.UUID            `json:"storagePlanId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID of the storage plan for this marina"`
	DocumentPlanID       *uuid.UUID            `json:"documentPlanId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000" description:"ID of the document plan for this marina"`
	Modules              *models.Modules       `json:"modules,omitempty"`
	InternalAnnouncement *string               `json:"internalAnnouncement,omitempty" example:"This is an internal announcement"`
	ExternalAnnouncement *string               `json:"externalAnnouncement,omitempty" example:"This is an external announcement"`
}

type OverLimitUsageRequest struct {
	OrganizationID *uuid.UUID `json:"organizationId" query:"organizationId" example:"550e8400-e29b-41d4-a716-446655440001"`
	StartMonth     string     `json:"startMonth" query:"startMonth" validate:"required" example:"2024-01"`
	EndMonth       string     `json:"endMonth" query:"endMonth" validate:"required" example:"2024-12"`
}

// MarinaIDParam represents the URL parameter for the marina ID
type MarinaIDParam struct {
	MarinaID uuid.UUID `param:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
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
