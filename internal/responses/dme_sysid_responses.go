package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// DMESysIDResponse represents a DME System ID in the system
// @Description DME System IDs used for marina integration
type DMESysIDResponse struct {
	ID             uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OrganizationID uuid.UUID  `json:"organizationId" example:"550e8400-e29b-41d4-a716-446655440001"`
	MarinaID       *uuid.UUID `json:"marinaId,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	Name           string     `json:"name" example:"Production System"`
	Description    *string    `json:"description,omitempty" example:"Primary production environment"`
	SystemID       string     `json:"systemId" example:"SYS123456"`
	IsActive       *bool      `json:"isActive,omitempty" example:"true"`
	CreatedAt      *time.Time `json:"createdAt,omitempty"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty"`
}

// ConvertDMESysIDToResponse converts a database DME System ID to a response model
func ConvertDMESysIDToResponse(sysid db.DmeSysid) DMESysIDResponse {
	return DMESysIDResponse{
		ID:             sysid.ID,
		OrganizationID: sysid.OrganizationID,
		MarinaID:       &sysid.MarinaID,
		Name:           sysid.Name,
		Description:    sysid.Description,
		SystemID:       sysid.SystemID,
		IsActive:       sysid.IsActive,
		CreatedAt:      utils.PgTimeToTimePtr(sysid.CreatedAt),
		UpdatedAt:      utils.PgTimeToTimePtr(sysid.UpdatedAt),
	}
}

// NewDMESysIDResponseSuccess creates a new successful DME System ID response
func NewDMESysIDResponseSuccess(sysid db.DmeSysid) BaseResponse {
	return NewSuccessResponse(ConvertDMESysIDToResponse(sysid))
}

// DMESysIDListResponse creates a list response of DME System IDs
func NewDMESysIDListResponse(sysids []db.DmeSysid) BaseResponse {
	sysidResponses := make([]DMESysIDResponse, len(sysids))
	for i, sysid := range sysids {
		sysidResponses[i] = ConvertDMESysIDToResponse(sysid)
	}
	return NewSuccessResponse(sysidResponses)
}

// DMESysIDResponseWrapper is purely for Swagger documentation
type DMESysIDResponseWrapper struct {
	Data DMESysIDResponse `json:"data"`
}

// DMESysIDListResponseWrapper is purely for Swagger documentation
type DMESysIDListResponseWrapper struct {
	Data []DMESysIDResponse `json:"data"`
}
