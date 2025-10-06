package requests

import "github.com/google/uuid"

// ClerkRequest represents the required parameters to retrieve a specific clerk
type ClerkRequest struct {
	ClerkID        string    `json:"clerkId" validate:"required" example:"CLERK001"`
	OrganizationID uuid.UUID `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	SystemID       string    `json:"systemId" validate:"required" example:"SYS001"`
}

// ClerkListRequest represents the required parameters to list all clerks
type ClerkListRequest struct {
	OrganizationID  uuid.UUID `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	SystemID        string    `json:"systemId" validate:"required" example:"SYS001"`
	IncludeInactive *bool     `json:"includeInactive" query:"includeInactive" example:"false"`
}

// LocationListRequest represents the required parameters to list all locations
type LocationListRequest struct {
	OrganizationID uuid.UUID `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	SystemID       string    `json:"systemId" validate:"required" example:"SYS001"`
}
