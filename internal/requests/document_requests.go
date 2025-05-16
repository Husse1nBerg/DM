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
}

// UpdateDocumentRequest represents the parameters that can be updated for a document
type UpdateDocumentRequest struct {
	FileName string `json:"fileName" validate:"required" example:"updated-document.pdf"`
	FileType string `json:"fileType" validate:"required" example:"application/pdf"`
	FileSize int64  `json:"fileSize" validate:"required" example:"2048"`
}

// // CreateCustomerDocumentRequest represents the required parameters to create a new customer document
// type CreateCustomerDocumentRequest struct {
// 	MarinaID   uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
// 	EntityType string    `json:"entityType" validate:"required,oneof=customer vessel user" example:"customer"`
// 	EntityID   string    `json:"entityId" validate:"required" example:"0000000917"`
// 	FileName   string    `json:"fileName" validate:"required" example:"document.pdf"`
// 	FileType   string    `json:"fileType" validate:"required" example:"application/pdf"`
// 	FileSize   int64     `json:"fileSize" validate:"required" example:"1024"`
// }

// // ListCustomerDocumentsRequest represents the parameters to list documents for a customer
// type ListCustomerDocumentsRequest struct {
// 	MarinaID   uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
// 	EntityType string    `json:"entityType" validate:"required,oneof=customer vessel user" example:"customer"`
// 	EntityID   string    `json:"entityId" validate:"required" example:"0000000917"`
// }

// // UpdateCustomerDocumentRequest represents the parameters that can be updated for a customer document
// type UpdateCustomerDocumentRequest struct {
// 	FileName string `json:"fileName" validate:"required" example:"updated-document.pdf"`
// 	FileType string `json:"fileType" validate:"required" example:"application/pdf"`
// 	FileSize int64  `json:"fileSize" validate:"required" example:"2048"`
// }

// // CreateVesselDocumentRequest represents the required parameters to create a new vessel document
// type CreateVesselDocumentRequest struct {
// 	MarinaID   uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
// 	EntityType string    `json:"entityType" validate:"required,oneof=customer vessel user" example:"vessel"`
// 	EntityID   string    `json:"entityId" validate:"required" example:"VESSEL789012"`
// 	FileName   string    `json:"fileName" validate:"required" example:"document.pdf"`
// 	FileType   string    `json:"fileType" validate:"required" example:"application/pdf"`
// 	FileSize   int64     `json:"fileSize" validate:"required" example:"1024"`
// }

// // ListVesselDocumentsRequest represents the parameters to list documents for a vessel
// type ListVesselDocumentsRequest struct {
// 	MarinaID   uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
// 	EntityType string    `json:"entityType" validate:"required,oneof=customer vessel user" example:"vessel"`
// 	EntityID   string    `json:"entityId" validate:"required" example:"VESSEL789012"`
// }

// // UpdateVesselDocumentRequest represents the parameters that can be updated for a vessel document
// type UpdateVesselDocumentRequest struct {
// 	FileName string `json:"fileName" validate:"required" example:"updated-document.pdf"`
// 	FileType string `json:"fileType" validate:"required" example:"application/pdf"`
// 	FileSize int64  `json:"fileSize" validate:"required" example:"2048"`
// }

// // CreateUserDocumentRequest represents the required parameters to create a new user document
// type CreateUserDocumentRequest struct {
// 	MarinaID   uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
// 	EntityType string    `json:"entityType" validate:"required,oneof=customer vessel user" example:"user"`
// 	EntityID   string    `json:"entityId" validate:"required" example:"USER123456"`
// 	FileName   string    `json:"fileName" validate:"required" example:"document.pdf"`
// 	FileType   string    `json:"fileType" validate:"required" example:"application/pdf"`
// 	FileSize   int64     `json:"fileSize" validate:"required" example:"1024"`
// }

// // ListUserDocumentsRequest represents the parameters to list documents for a user
// type ListUserDocumentsRequest struct {
// 	MarinaID   uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
// 	EntityType string    `json:"entityType" validate:"required,oneof=customer vessel user" example:"user"`
// 	EntityID   string    `json:"entityId" validate:"required" example:"USER123456"`
// }

// // UpdateUserDocumentRequest represents the parameters that can be updated for a user document
// type UpdateUserDocumentRequest struct {
// 	FileName string `json:"fileName" validate:"required" example:"updated-document.pdf"`
// 	FileType string `json:"fileType" validate:"required" example:"application/pdf"`
// 	FileSize int64  `json:"fileSize" validate:"required" example:"2048"`
// }
