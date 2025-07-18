package requests

import (
	"github.com/dockworks/dm-web-backend/pkg/models"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// RoleIDParam represents the URL parameter for role ID
type RoleIDParam struct {
	RoleID uuid.UUID `param:"roleId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// CreateRoleRequest represents the required parameters to create a new role
type CreateRoleRequest struct {
	Name           string              `json:"name" validate:"required" example:"Admin"`
	Description    *string             `json:"description,omitempty" example:"Administrator role with full access"`
	Permissions    *models.Permissions `json:"permissions" validate:"required"`
	IsActive       *bool               `json:"isActive,omitempty" example:"true"`
	IsCustomerRole *bool               `json:"isCustomerRole,omitempty" example:"false"`
	Type           string              `json:"type" validate:"required" example:"marina"`
	MarinaID       *uuid.UUID          `json:"marinaId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// UpdateRoleRequest represents the parameters that can be updated for a role
type UpdateRoleRequest struct {
	Name           *string             `json:"name,omitempty" example:"Admin"`
	Description    *string             `json:"description,omitempty" example:"Administrator role with full access"`
	Permissions    *models.Permissions `json:"permissions,omitempty"`
	IsActive       *bool               `json:"isActive,omitempty" example:"true"`
	IsCustomerRole *bool               `json:"isCustomerRole,omitempty" example:"false"`
	Type           string              `json:"type,omitempty" example:"marina"`
	MarinaID       *uuid.UUID          `json:"marinaId,omitempty" example:"550e8400-e29b-41d4-a716-446655440000"`
}

// Validate performs custom validation on the create role request
func (r *CreateRoleRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the update role request
func (r *UpdateRoleRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
