package requests

import "github.com/google/uuid"

type BasicAuth struct {
	Email    string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Password string `json:"password" validate:"required,min=12,letters,number,specialchar" example:"Pa$$w0rd123"`
}

type LoginRequest struct {
	BasicAuth
}

type RegisterRequest struct {
	BasicAuth
	FirstName string    `json:"first_name" validate:"required,min=2" example:"John"`
	LastName  string    `json:"last_name" validate:"required,min=2" example:"Doe"`
	RoleID    uuid.UUID `json:"role_id" validate:"required" example:"admin"`
	Username  string    `json:"username" validate:"required,min=3,alphanum" example:"johndoe"`
}

type RefreshRequest struct {
	Token string `json:"token" validate:"required" example:"refresh_token"`
}
