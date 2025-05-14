package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// MarinaGalleryItemResponse represents a marina gallery item in the system
// @Description Marina gallery item data including image URL and description
type MarinaGalleryItemResponse struct {
	ID          uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MarinaID    uuid.UUID  `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440001"`
	ImageURL    string     `json:"imageUrl" example:"/images/marinas/sunset_view.jpg"`
	Description *string    `json:"description,omitempty" example:"Beautiful view of the marina at sunset"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

// VesselGalleryItemResponse represents a vessel gallery item in the system
// @Description Vessel gallery item data including image URL, description, and main image flag
type VesselGalleryItemResponse struct {
	ID          uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MarinaID    uuid.UUID  `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440001"`
	CustomerID  string     `json:"customerId" example:"CUST123456"`
	VesselID    string     `json:"boatId" example:"BOAT123456"`
	ImageURL    string     `json:"imageUrl" example:"/images/vessels/yacht_port_side.jpg"`
	Description *string    `json:"description,omitempty" example:"Port side view of the yacht"`
	Main        *bool      `json:"main,omitempty" example:"true"`
	CreatedAt   *time.Time `json:"createdAt,omitempty"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty"`
}

// Convert a database MarinaGallery to a response model
func ConvertMarinaGalleryItemToResponse(item db.MarinaGallery) MarinaGalleryItemResponse {
	return MarinaGalleryItemResponse{
		ID:          item.ID,
		MarinaID:    item.MarinaID,
		ImageURL:    *utils.GetFullImageURL(&item.ImageUrl),
		Description: item.Description,
		CreatedAt:   utils.PgTimeToTimePtr(item.CreatedAt),
		UpdatedAt:   utils.PgTimeToTimePtr(item.UpdatedAt),
	}
}

// Convert a database VesselGallery to a response model
func ConvertVesselGalleryItemToResponse(item db.VesselGallery) VesselGalleryItemResponse {
	return VesselGalleryItemResponse{
		ID:          item.ID,
		MarinaID:    item.MarinaID,
		CustomerID:  item.CustomerID,
		VesselID:    item.VesselID,
		ImageURL:    *utils.GetFullImageURL(&item.ImageUrl),
		Description: item.Description,
		Main:        item.Main,
		CreatedAt:   utils.PgTimeToTimePtr(item.CreatedAt),
		UpdatedAt:   utils.PgTimeToTimePtr(item.UpdatedAt),
	}
}

// NewMarinaGalleryItemResponseSuccess creates a successful response with a marina gallery item
func NewMarinaGalleryItemResponseSuccess(item db.MarinaGallery) BaseResponse {
	return NewSuccessResponse(ConvertMarinaGalleryItemToResponse(item))
}

// NewVesselGalleryItemResponseSuccess creates a successful response with a vessel gallery item
func NewVesselGalleryItemResponseSuccess(item db.VesselGallery) BaseResponse {
	return NewSuccessResponse(ConvertVesselGalleryItemToResponse(item))
}

// NewMarinaGalleryResponseSuccess creates a successful response with a list of marina gallery items
func NewMarinaGalleryResponseSuccess(items []db.MarinaGallery) BaseResponse {
	var response []MarinaGalleryItemResponse
	for _, item := range items {
		response = append(response, ConvertMarinaGalleryItemToResponse(item))
	}
	return NewSuccessResponse(response)
}

// NewVesselGalleryResponseSuccess creates a successful response with a list of vessel gallery items
func NewVesselGalleryResponseSuccess(items []db.VesselGallery) BaseResponse {
	var response []VesselGalleryItemResponse
	for _, item := range items {
		response = append(response, ConvertVesselGalleryItemToResponse(item))
	}
	return NewSuccessResponse(response)
}
