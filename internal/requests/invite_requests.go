package requests

import (
	"github.com/go-playground/validator/v10"
)

type ConfirmTokenRequest struct {
	Token string `json:"token" validate:"required"`
	Email string `json:"email" validate:"required,email"`
}

type AcceptInvitationRequest struct {
	Token    string `json:"token" validate:"required"`
	Password string `json:"password" validate:"required,min=12,letters,number,specialchar" example:"SecureP@ssw0rd"`
}

func (c *ConfirmTokenRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(c)
}

func (a *AcceptInvitationRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(a)
}
