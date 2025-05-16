package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// DocumentResponse represents a document in the system
// @Description Document data including file path, file type, and size
type DocumentResponse struct {
	ID         uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MarinaID   uuid.UUID  `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440001"`
	EntityType string     `json:"entityType" example:"customer"`
	EntityID   string     `json:"entityId" example:"0000000000"`
	FileName   string     `json:"fileName" example:"contract.pdf"`
	FileType   string     `json:"fileType" example:"application/pdf"`
	FilePath   string     `json:"filePath" example:"/documents/customers/550e8400-e29b-41d4-a716-446655440002/contract.pdf"`
	FileSize   int64      `json:"fileSize" example:"1024"`
	FileURL    string     `json:"fileUrl" example:"https://example.com/documents/customers/550e8400-e29b-41d4-a716-446655440002/contract.pdf"`
	CreatedAt  *time.Time `json:"createdAt,omitempty"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
}

// Convert a database Document to a response model
func ConvertDocumentToResponse(doc db.Document) DocumentResponse {
	fileURL := utils.GetFullDocumentURL(&doc.FilePath)
	return DocumentResponse{
		ID:         doc.ID,
		MarinaID:   doc.MarinaID,
		EntityType: doc.EntityType,
		EntityID:   doc.EntityID,
		FileName:   doc.FileName,
		FileType:   doc.FileType,
		FilePath:   doc.FilePath,
		FileSize:   doc.FileSize,
		FileURL:    *fileURL,
		CreatedAt:  utils.PgTimeToTimePtr(doc.CreatedAt),
		UpdatedAt:  utils.PgTimeToTimePtr(doc.UpdatedAt),
	}
}

// NewDocumentResponseSuccess creates a successful response with a document
func NewDocumentResponseSuccess(doc db.Document) BaseResponse {
	return NewSuccessResponse(ConvertDocumentToResponse(doc))
}

// NewDocumentsResponseSuccess creates a successful response with a list of documents
func NewDocumentsResponseSuccess(docs []db.Document) BaseResponse {
	var response []DocumentResponse
	for _, doc := range docs {
		response = append(response, ConvertDocumentToResponse(doc))
	}
	return NewSuccessResponse(response)
}

// // DocumentResponse represents a document in the system
// // @Description Document data including file path, file type, and size
// type CustomerDocumentResponse struct {
// 	ID         uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
// 	MarinaID   uuid.UUID  `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440001"`
// 	EntityType string     `json:"entityType" example:"customer"`
// 	EntityID   string     `json:"entityId" example:"0000000000"`
// 	FileName   string     `json:"fileName" example:"contract.pdf"`
// 	FileType   string     `json:"fileType" example:"application/pdf"`
// 	FilePath   string     `json:"filePath" example:"/documents/customers/550e8400-e29b-41d4-a716-446655440002/contract.pdf"`
// 	FileSize   int64      `json:"fileSize" example:"1024"`
// 	FileURL    string     `json:"fileUrl" example:"https://example.com/documents/customers/550e8400-e29b-41d4-a716-446655440002/contract.pdf"`
// 	CreatedAt  *time.Time `json:"createdAt,omitempty"`
// 	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
// }

// // Convert a database Document to a response model
// func ConvertCustomerDocumentToResponse(doc db.Document) CustomerDocumentResponse {
// 	fileURL := utils.GetFullDocumentURL(&doc.FilePath)
// 	return CustomerDocumentResponse{
// 		ID:         doc.ID,
// 		MarinaID:   doc.MarinaID,
// 		EntityType: doc.EntityType,
// 		EntityID:   doc.EntityID,
// 		FileName:   doc.FileName,
// 		FileType:   doc.FileType,
// 		FilePath:   doc.FilePath,
// 		FileSize:   doc.FileSize,
// 		FileURL:    *fileURL,
// 		CreatedAt:  utils.PgTimeToTimePtr(doc.CreatedAt),
// 		UpdatedAt:  utils.PgTimeToTimePtr(doc.UpdatedAt),
// 	}
// }

// // NewDocumentResponseSuccess creates a successful response with a document
// func NewCustomerDocumentResponseSuccess(doc db.Document) BaseResponse {
// 	return NewSuccessResponse(ConvertCustomerDocumentToResponse(doc))
// }

// // NewDocumentsResponseSuccess creates a successful response with a list of documents
// func NewCustomerDocumentsResponseSuccess(docs []db.Document) BaseResponse {
// 	var response []CustomerDocumentResponse
// 	for _, doc := range docs {
// 		response = append(response, ConvertCustomerDocumentToResponse(doc))
// 	}
// 	return NewSuccessResponse(response)
// }

// // VesselDocumentResponse represents a vessel document in the system
// // @Description Vessel document data including file path, file type, and size
// type VesselDocumentResponse struct {
// 	ID         uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
// 	MarinaID   uuid.UUID  `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440001"`
// 	EntityType string     `json:"entityType" example:"vessel"`
// 	EntityID   string     `json:"entityId" example:"VESSEL123456"`
// 	FileName   string     `json:"fileName" example:"contract.pdf"`
// 	FileType   string     `json:"fileType" example:"application/pdf"`
// 	FilePath   string     `json:"filePath" example:"/documents/vessels/550e8400-e29b-41d4-a716-446655440002/contract.pdf"`
// 	FileSize   int64      `json:"fileSize" example:"1024"`
// 	FileURL    string     `json:"fileUrl" example:"https://example.com/documents/vessels/550e8400-e29b-41d4-a716-446655440002/contract.pdf"`
// 	CreatedAt  *time.Time `json:"createdAt,omitempty"`
// 	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
// }

// // Convert a database VesselDocument to a response model
// func ConvertVesselDocumentToResponse(doc db.Document) VesselDocumentResponse {
// 	fileURL := utils.GetFullDocumentURL(&doc.FilePath)
// 	return VesselDocumentResponse{
// 		ID:         doc.ID,
// 		MarinaID:   doc.MarinaID,
// 		EntityType: doc.EntityType,
// 		EntityID:   doc.EntityID,
// 		FileName:   doc.FileName,
// 		FileType:   doc.FileType,
// 		FilePath:   doc.FilePath,
// 		FileSize:   doc.FileSize,
// 		FileURL:    *fileURL,
// 		CreatedAt:  utils.PgTimeToTimePtr(doc.CreatedAt),
// 		UpdatedAt:  utils.PgTimeToTimePtr(doc.UpdatedAt),
// 	}
// }

// // NewVesselDocumentResponseSuccess creates a successful response with a vessel document
// func NewVesselDocumentResponseSuccess(doc db.Document) BaseResponse {
// 	return NewSuccessResponse(ConvertVesselDocumentToResponse(doc))
// }

// // NewVesselDocumentsResponseSuccess creates a successful response with a list of vessel documents
// func NewVesselDocumentsResponseSuccess(docs []db.Document) BaseResponse {
// 	var response []VesselDocumentResponse
// 	for _, doc := range docs {
// 		response = append(response, ConvertVesselDocumentToResponse(doc))
// 	}
// 	return NewSuccessResponse(response)
// }

// // UserDocumentResponse represents a user document in the system
// // @Description User document data including file path, file type, and size
// type UserDocumentResponse struct {
// 	ID         uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
// 	MarinaID   uuid.UUID  `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440001"`
// 	EntityType string     `json:"entityType" example:"user"`
// 	EntityID   string     `json:"entityId" example:"USER123456"`
// 	FileName   string     `json:"fileName" example:"contract.pdf"`
// 	FileType   string     `json:"fileType" example:"application/pdf"`
// 	FilePath   string     `json:"filePath" example:"/documents/users/550e8400-e29b-41d4-a716-446655440002/contract.pdf"`
// 	FileSize   int64      `json:"fileSize" example:"1024"`
// 	FileURL    string     `json:"fileUrl" example:"https://example.com/documents/users/550e8400-e29b-41d4-a716-446655440002/contract.pdf"`
// 	CreatedAt  *time.Time `json:"createdAt,omitempty"`
// 	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
// }

// // Convert a database UserDocument to a response model
// func ConvertUserDocumentToResponse(doc db.Document) UserDocumentResponse {
// 	fileURL := utils.GetFullDocumentURL(&doc.FilePath)
// 	return UserDocumentResponse{
// 		ID:         doc.ID,
// 		MarinaID:   doc.MarinaID,
// 		EntityType: doc.EntityType,
// 		EntityID:   doc.EntityID,
// 		FileName:   doc.FileName,
// 		FileType:   doc.FileType,
// 		FilePath:   doc.FilePath,
// 		FileSize:   doc.FileSize,
// 		FileURL:    *fileURL,
// 		CreatedAt:  utils.PgTimeToTimePtr(doc.CreatedAt),
// 		UpdatedAt:  utils.PgTimeToTimePtr(doc.UpdatedAt),
// 	}
// }

// // NewUserDocumentResponseSuccess creates a successful response with a user document
// func NewUserDocumentResponseSuccess(doc db.Document) BaseResponse {
// 	return NewSuccessResponse(ConvertUserDocumentToResponse(doc))
// }

// // NewUserDocumentsResponseSuccess creates a successful response with a list of user documents
// func NewUserDocumentsResponseSuccess(docs []db.Document) BaseResponse {
// 	var response []UserDocumentResponse
// 	for _, doc := range docs {
// 		response = append(response, ConvertUserDocumentToResponse(doc))
// 	}
// 	return NewSuccessResponse(response)
// }
