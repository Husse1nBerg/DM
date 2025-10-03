package requests

import "github.com/google/uuid"

type BasicAuth struct {
	Email string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	//Password string `json:"password" validate:"required,min=12,letters,number,specialchar" example:"Pa$$w0rd123"`
	Password string `json:"password" validate:"required" example:"Pa$$w0rd123"`
}

type LoginRequest struct {
	BasicAuth
}

type RefreshRequest struct {
	Token string `json:"token" validate:"required" example:"refresh_token"`
}

// ResetPasswordRequest defines the request parameters for password reset
// @Description Password reset request payload
// @Schema requests.ResetPasswordRequest
type ResetPasswordRequest struct {
	UserID      uuid.UUID `json:"userId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	OldPassword string    `json:"oldPassword" validate:"required" example:"OldPa$$w0rd123"`
	NewPassword string    `json:"newPassword" validate:"required,min=12,letters,number,specialchar" example:"NewPa$$w0rd123"`
}

// ForgotPasswordRequest defines the request to initiate password recovery
// @Description Forgot password request payload
// @Schema requests.ForgotPasswordRequest
type ForgotPasswordRequest struct {
	Email string `json:"email" validate:"required,email" example:"john.doe@example.com"`
}

// CompletePasswordRecoveryRequest defines the request to complete the password recovery process
// @Description Complete password recovery request payload
// @Schema requests.CompletePasswordRecoveryRequest
type CompletePasswordRecoveryRequest struct {
	Email       string `json:"email" validate:"required,email" example:"john.doe@example.com"`
	Token       string `json:"token" validate:"required" example:"recovery-token-123"`
	NewPassword string `json:"newPassword" validate:"required,min=12,letters,number,specialchar" example:"NewPa$$w0rd123"`
}
