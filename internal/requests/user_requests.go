package requests

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CreateUserRequest represents the required parameters to create a new user
type CreateUserRequest struct {
	Username       string    `json:"username" example:"johndoe"`
	FirstName      string    `json:"firstName" validate:"required" example:"John"`
	LastName       string    `json:"lastName" validate:"required" example:"Doe"`
	Email          string    `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Phone          *string   `json:"phone,omitempty" example:"+15551234567"`
	Title          *string   `json:"title,omitempty" example:"Manager"`
	Image          *string   `json:"image,omitempty" example:"/images/profiles/johndoe.jpg"`
	Password       string    `json:"password" validate:"required,min=12,letters,number,specialchar" example:"SecureP@ssw0rd"`
	OrganizationID uuid.UUID `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	MarinaID       uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440002"`
	RoleID         uuid.UUID `json:"roleId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440003"`
	IsSuperuser    *bool     `json:"isSuperuser,omitempty" example:"false"`
	IsActive       *bool     `json:"isActive,omitempty" example:"true"`
	UserAnalytics  *bool     `json:"userAnalytics,omitempty" example:"true"`
}

type CreateCustomerUserRequest struct {
	Username       string    `json:"username" example:"johndoe"`
	FirstName      string    `json:"firstName" validate:"required" example:"John"`
	LastName       string    `json:"lastName" validate:"required" example:"Doe"`
	Email          string    `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Phone          *string   `json:"phone,omitempty" example:"+15551234567"`
	Title          *string   `json:"title,omitempty" example:"Manager"`
	Image          *string   `json:"image,omitempty" example:"/images/profiles/johndoe.jpg"`
	OrganizationID uuid.UUID `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	MarinaID       uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440002"`
	RoleID         uuid.UUID `json:"roleId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440003"`
	CustomerID     *string   `json:"customerId,omitempty" validate:"required" example:"1234567890"`
	IsCustomer     *bool     `json:"isCustomer,omitempty" example:"false"`
	IsSuperuser    *bool     `json:"isSuperuser,omitempty" example:"false"`
	IsActive       *bool     `json:"isActive,omitempty" example:"true"`
	UserAnalytics  *bool     `json:"userAnalytics,omitempty" example:"true"`
}

type CreateUserWithInvitationRequest struct {
	Username       string    `json:"username" example:"johndoe"`
	FirstName      string    `json:"firstName" validate:"required" example:"John"`
	LastName       string    `json:"lastName" validate:"required" example:"Doe"`
	Email          string    `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Phone          *string   `json:"phone,omitempty" example:"+15551234567"`
	Title          *string   `json:"title,omitempty" example:"Manager"`
	Image          *string   `json:"image,omitempty" example:"/images/profiles/johndoe.jpg"`
	OrganizationID uuid.UUID `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	MarinaID       uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440002"`
	RoleID         uuid.UUID `json:"roleId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440003"`
	IsSuperuser    *bool     `json:"isSuperuser,omitempty" example:"false"`
	IsActive       *bool     `json:"isActive,omitempty" example:"true"`
	UserAnalytics  *bool     `json:"userAnalytics,omitempty" example:"true"`
}

// UpdateUserRequest represents the parameters that can be updated for a user
type UpdateUserRequest struct {
	FirstName     *string    `json:"firstName,omitempty" validate:"omitempty" example:"John"`
	LastName      *string    `json:"lastName,omitempty" validate:"omitempty" example:"Doe"`
	Email         *string    `json:"email,omitempty" validate:"omitempty,email" example:"john.doe@example.com"`
	Phone         *string    `json:"phone,omitempty" example:"+15551234567"`
	Title         *string    `json:"title,omitempty" example:"Manager"`
	Image         *string    `json:"image,omitempty" example:"/images/profiles/johndoe.jpg"`
	Password      *string    `json:"password,omitempty" validate:"omitempty,min=12,letters,number,specialchar" example:"NewSecureP@ssw0rd"`
	MarinaID      *uuid.UUID `json:"marinaId,omitempty" validate:"omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	RoleID        *uuid.UUID `json:"roleId,omitempty" validate:"omitempty" example:"550e8400-e29b-41d4-a716-446655440003"`
	IsSuperuser   *bool      `json:"isSuperuser,omitempty" example:"false"`
	IsActive      *bool      `json:"isActive,omitempty" example:"true"`
	UserAnalytics *bool      `json:"userAnalytics,omitempty" example:"true"`
}

// UserIDParam represents a URL parameter for user ID
type UserIDParam struct {
	UserID uuid.UUID `param:"userId" validate:"required"`
}

// AssignUserToMarinaRequest represents the parameters to assign a user to a marina
type AssignUserToMarinaRequest struct {
	UserID     uuid.UUID `json:"userId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	MarinaID   uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440002"`
	CustomerID *string   `json:"customerId,omitempty" example:"1234567890"`
	RoleID     uuid.UUID `json:"roleId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440003"`
}

// Validate performs custom validation on the request
func (r *CreateUserRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the request
func (r *CreateUserWithInvitationRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the request
func (r *CreateCustomerUserRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the request
func (r *UpdateUserRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the request
func (r *AssignUserToMarinaRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
