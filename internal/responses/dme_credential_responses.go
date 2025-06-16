package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// DMECredentialResponse represents a DME credential in the system
// @Description DME API credentials used for integration
type DMECredentialResponse struct {
	ID             uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OrganizationID uuid.UUID  `json:"organizationId" example:"550e8400-e29b-41d4-a716-446655440001"`
	Username       string     `json:"username" example:"api_user"`
	IsOldAPI       *bool      `json:"isOldApi,omitempty" example:"false"`
	AccessToken    *string    `json:"accessToken,omitempty"`
	RefreshToken   *string    `json:"refreshToken,omitempty"`
	ExpiryDate     *time.Time `json:"expiryDate,omitempty"`
	CreatedAt      *time.Time `json:"createdAt,omitempty"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
}

// ConvertDMECredentialToResponse converts a database DME credential to a response model
func ConvertDMECredentialToResponse(cred db.DmeCredential) DMECredentialResponse {
	return DMECredentialResponse{
		ID:             cred.ID,
		OrganizationID: cred.OrganizationID,
		Username:       cred.Username,
		IsOldAPI:       cred.IsOldApi,
		AccessToken:    cred.AccessToken,
		RefreshToken:   cred.RefreshToken,
		ExpiryDate:     utils.PgTimeToTimePtr(cred.ExpiryDate),
		CreatedAt:      utils.PgTimeToTimePtr(cred.CreatedAt),
		UpdatedAt:      utils.PgTimeToTimePtr(cred.UpdatedAt),
	}
}

// NewDMECredentialResponseSuccess creates a new successful DME credential response
func NewDMECredentialResponseSuccess(cred db.DmeCredential) BaseResponse {
	return NewSuccessResponse(ConvertDMECredentialToResponse(cred))
}

// DMECredentialResponseWrapper is purely for Swagger documentation
type DMECredentialResponseWrapper struct {
	Data DMECredentialResponse `json:"data"`
}
