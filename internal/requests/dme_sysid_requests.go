package requests

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CreateDMESysIDRequest represents the request to create a DME system ID
type CreateDMESysIDRequest struct {
	OrganizationID uuid.UUID  `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	MarinaID       *uuid.UUID `json:"marinaId,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	Name           string     `json:"name" validate:"required" example:"Production System"`
	Description    *string    `json:"description,omitempty" example:"Primary production environment for marina operations"`
	SystemID       string     `json:"systemId" validate:"required" example:"SYS123456"`
	IsActive       *bool      `json:"isActive,omitempty" example:"true"`
}

// UpdateDMESysIDRequest represents the request to update a DME system ID
type UpdateDMESysIDRequest struct {
	MarinaID    *uuid.UUID `json:"marinaId,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	Name        *string    `json:"name,omitempty" example:"Production System"`
	Description *string    `json:"description,omitempty" example:"Primary production environment for marina operations"`
	SystemID    *string    `json:"systemId,omitempty" example:"SYS123456"`
	IsActive    *bool      `json:"isActive,omitempty" example:"true"`
}

// LinkDMESysIDRequest represents the request to link a DME system ID to a marina
type LinkDMESysIDRequest struct {
	MarinaID uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440002"`
}

// Validate performs custom validation on the create DME system ID request
func (r *CreateDMESysIDRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the update DME system ID request
func (r *UpdateDMESysIDRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the link DME system ID request
func (r *LinkDMESysIDRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
