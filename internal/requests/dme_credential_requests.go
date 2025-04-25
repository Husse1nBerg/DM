package requests

import (
	"time"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// CreateDMECredentialRequest represents the request to create DME credentials
type CreateDMECredentialRequest struct {
	OrganizationID uuid.UUID  `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Username       string     `json:"username" validate:"required" example:"api_user"`
	Password       *string    `json:"password,omitempty" example:"secure_password"`
	IsOldAPI       *bool      `json:"isOldApi,omitempty" example:"false"`
	AccessToken    *string    `json:"accessToken,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken   *string    `json:"refreshToken,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiryDate     *time.Time `json:"expiryDate,omitempty"`
}

// UpdateDMECredentialRequest represents the request to update DME credentials
type UpdateDMECredentialRequest struct {
	Username     *string    `json:"username,omitempty" example:"api_user"`
	Password     *string    `json:"password,omitempty" example:"secure_password"`
	IsOldAPI     *bool      `json:"isOldApi,omitempty" example:"false"`
	AccessToken  *string    `json:"accessToken,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken *string    `json:"refreshToken,omitempty" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiryDate   *time.Time `json:"expiryDate,omitempty"`
}

// UpdateDMETokenRequest represents the request to update only the token information
type UpdateDMETokenRequest struct {
	AccessToken  string     `json:"accessToken" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	RefreshToken string     `json:"refreshToken" validate:"required" example:"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."`
	ExpiryDate   *time.Time `json:"expiryDate,omitempty"`
}

// Validate performs custom validation on the create DME credential request
func (r *CreateDMECredentialRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the update DME credential request
func (r *UpdateDMECredentialRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the update DME token request
func (r *UpdateDMETokenRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
