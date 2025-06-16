package requests

import (
	"github.com/google/uuid"
)

// CreateMarinaGalleryItemRequest represents the required parameters to create a new marina gallery item
type CreateMarinaGalleryItemRequest struct {
	MarinaID    uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Description *string   `json:"description,omitempty" example:"Beautiful view of the marina at sunset"`
}

// UpdateMarinaGalleryItemRequest represents the parameters that can be updated for a marina gallery item
type UpdateMarinaGalleryItemRequest struct {
	Description *string `json:"description,omitempty" example:"Updated description of the marina view"`
}

// CreateVesselGalleryItemRequest represents the required parameters to create a new vessel gallery item
type CreateVesselGalleryItemRequest struct {
	MarinaID    uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	CustomerID  string    `json:"customerId" validate:"required" example:"CUST123456"`
	VesselID    string    `json:"vesselId" validate:"required" example:"VESSEL789012"`
	Description *string   `json:"description,omitempty" example:"Port side view of the yacht"`
	Main        *bool     `json:"main,omitempty" example:"true"`
}

// UpdateVesselGalleryItemRequest represents the parameters that can be updated for a vessel gallery item
type UpdateVesselGalleryItemRequest struct {
	Description *string `json:"description,omitempty" example:"Updated description of the vessel image"`
	Main        *bool   `json:"main,omitempty" example:"true"`
}
