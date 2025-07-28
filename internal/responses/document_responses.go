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
	Public     *bool      `json:"public,omitempty" example:"true"`
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
		Public:     &doc.Public,
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
	documentResponses := make([]DocumentResponse, len(docs))
	for i, doc := range docs {
		documentResponses[i] = ConvertDocumentToResponse(doc)
	}
	return NewSuccessResponse(documentResponses)
}
