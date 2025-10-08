package requests

import (
	"github.com/google/uuid"
)

// CreateDocumentRequest represents the required parameters to create a new document
type CreateDocumentRequest struct {
	MarinaID   uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	EntityType string    `json:"entityType" validate:"required,oneof=customer vessel user" example:"customer"`
	EntityID   string    `json:"entityId" validate:"required" example:"0000000917"`
	FileName   string    `json:"fileName" validate:"required" example:"document.pdf"`
	FileType   string    `json:"fileType" validate:"required" example:"application/pdf"`
	FileSize   int64     `json:"fileSize" validate:"required" example:"1024"`
}

// ListDocumentsRequest represents the parameters to list documents for an entity
type ListDocumentsRequest struct {
	MarinaID   uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	EntityType string    `json:"entityType" validate:"required,oneof=customer vessel user" example:"customer"`
	EntityID   string    `json:"entityId" validate:"required" example:"0000000917"`
	Public     *bool     `json:"public,omitempty" example:"true"`
}

// UpdateDocumentRequest represents the parameters that can be updated for a document
type UpdateDocumentRequest struct {
	FileName string `json:"fileName" validate:"required" example:"updated-document.pdf"`
	FileType string `json:"fileType" validate:"required" example:"application/pdf"`
	FileSize int64  `json:"fileSize" validate:"required" example:"2048"`
	Public   *bool  `json:"public,omitempty" example:"true"`
}

// UpdateBoatDocumentPublicRequest represents the request to update only the public field of a boat document
type UpdateBoatDocumentPublicRequest struct {
	Public bool `json:"public" example:"true"`
}
