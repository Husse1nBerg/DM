package requests

import (
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

type ConfirmTokenRequest struct {
	Token string `json:"token" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type AcceptInvitationRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=12,letters,number,specialchar" example:"SecureP@ssw0rd"`
}

type RefreshInviteRequest struct {
	UserID *uuid.UUID `json:"userId,omitempty" validate:"required_without=Email" example:"550e8400-e29b-41d4-a716-446655440000"`
	Email  *string    `json:"email,omitempty" validate:"required_without=UserID,omitempty,email" example:"user@example.com"`
}

func (c *ConfirmTokenRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

func (a *AcceptInvitationRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}

func (r *RefreshInviteRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
