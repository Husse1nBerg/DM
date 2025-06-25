package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// EsignTemplateResponse represents an e-signature template in the system
// @Description E-signature template data including blob URL and metadata
type EsignTemplateResponse struct {
	ID             uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	OrganizationID uuid.UUID  `json:"organizationId" example:"550e8400-e29b-41d4-a716-446655440001"`
	MarinaID       *uuid.UUID `json:"marinaId,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	Name           string     `json:"name" example:"Customer Agreement Template"`
	Description    *string    `json:"description,omitempty" example:"Standard customer agreement template for marina services"`
	Type           string     `json:"type" example:"agreement"`
	Status         string     `json:"status" example:"active"`
	BlobURL        string     `json:"blobUrl" example:"https://s3.amazonaws.com/bucket/templates/agreement.pdf"`
	// BlobMetadata   *json.RawMessage `json:"blobMetadata,omitempty" example:"{\"size\": 1024, \"contentType\": \"application/pdf\"}"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// EsignDocumentResponse represents an e-signature document in the system
// @Description E-signature document data including blob URL, metadata, and signature status
type EsignDocumentResponse struct {
	ID             uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	TemplateID     *uuid.UUID `json:"templateId,omitempty" example:"550e8400-e29b-41d4-a716-446655440003"`
	OrganizationID uuid.UUID  `json:"organizationId" example:"550e8400-e29b-41d4-a716-446655440001"`
	MarinaID       uuid.UUID  `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440002"`
	Type           string     `json:"type" example:"agreement"`
	Status         string     `json:"status" example:"draft"`
	BlobURL        string     `json:"blobUrl" example:"https://s3.amazonaws.com/bucket/documents/agreement-001.pdf"`
	// BlobMetadata   *json.RawMessage `json:"blobMetadata,omitempty" example:"{\"size\": 1024, \"contentType\": \"application/pdf\"}"`
	CreatedAt *time.Time `json:"createdAt,omitempty"`
	UpdatedAt *time.Time `json:"updatedAt,omitempty"`
}

// Convert a database EsignTemplate to a response model
func ConvertEsignTemplateToResponse(template db.EsignTemplate) EsignTemplateResponse {
	blobURL := utils.GetFullESignURL(&template.BlobUrl)
	// var blobMetadata *json.RawMessage
	// if len(template.BlobMetadata) > 0 {
	// 	raw := json.RawMessage(template.BlobMetadata)
	// 	blobMetadata = &raw
	// }

	var marinaID *uuid.UUID
	if template.MarinaID != uuid.Nil {
		marinaID = &template.MarinaID
	}

	return EsignTemplateResponse{
		ID:             template.ID,
		OrganizationID: template.OrganizationID,
		MarinaID:       marinaID,
		Name:           template.Name,
		Description:    template.Description,
		Type:           template.Type,
		Status:         template.Status,
		BlobURL:        *blobURL,
		// BlobMetadata:   blobMetadata,
		CreatedAt: utils.PgTimeToTimePtr(template.CreatedAt),
		UpdatedAt: utils.PgTimeToTimePtr(template.UpdatedAt),
	}
}

// Convert a database EsignDocument to a response model
func ConvertEsignDocumentToResponse(document db.EsignDocument) EsignDocumentResponse {
	blobURL := utils.GetFullESignURL(&document.BlobUrl)
	// var blobMetadata *json.RawMessage
	// if len(document.BlobMetadata) > 0 {
	// 	raw := json.RawMessage(document.BlobMetadata)
	// 	blobMetadata = &raw
	// }

	var templateID *uuid.UUID
	if document.TemplateID != uuid.Nil {
		templateID = &document.TemplateID
	}

	return EsignDocumentResponse{
		ID:             document.ID,
		TemplateID:     templateID,
		OrganizationID: document.OrganizationID,
		MarinaID:       document.MarinaID,
		Type:           document.Type,
		Status:         document.Status,
		BlobURL:        *blobURL,
		// BlobMetadata:   blobMetadata,
		CreatedAt: utils.PgTimeToTimePtr(document.CreatedAt),
		UpdatedAt: utils.PgTimeToTimePtr(document.UpdatedAt),
	}
}

// NewEsignTemplateResponseSuccess creates a successful response with an e-signature template
func NewEsignTemplateResponseSuccess(template db.EsignTemplate) BaseResponse {
	return NewSuccessResponse(ConvertEsignTemplateToResponse(template))
}

// NewEsignTemplatesResponseSuccess creates a successful response with a list of e-signature templates
func NewEsignTemplatesResponseSuccess(templates []db.EsignTemplate) BaseResponse {
	templateResponses := make([]EsignTemplateResponse, len(templates))
	for i, template := range templates {
		templateResponses[i] = ConvertEsignTemplateToResponse(template)
	}
	return NewSuccessResponse(templateResponses)
}

// NewEsignDocumentResponseSuccess creates a successful response with an e-signature document
func NewEsignDocumentResponseSuccess(document db.EsignDocument) BaseResponse {
	return NewSuccessResponse(ConvertEsignDocumentToResponse(document))
}

// NewEsignDocumentsResponseSuccess creates a successful response with a list of e-signature documents
func NewEsignDocumentsResponseSuccess(documents []db.EsignDocument) BaseResponse {
	documentResponses := make([]EsignDocumentResponse, len(documents))
	for i, document := range documents {
		documentResponses[i] = ConvertEsignDocumentToResponse(document)
	}
	return NewSuccessResponse(documentResponses)
}

// NewEsignTemplatesPaginatedResponse creates a paginated response with e-signature templates
func NewEsignTemplatesPaginatedResponse(templates []db.EsignTemplate, total int64, perPage, currentPage int32) BaseResponse {
	templateResponses := make([]EsignTemplateResponse, len(templates))
	for i, template := range templates {
		templateResponses[i] = ConvertEsignTemplateToResponse(template)
	}
	return NewPaginatedResponse(templateResponses, total, perPage, currentPage)
}

// NewEsignDocumentsPaginatedResponse creates a paginated response with e-signature documents
func NewEsignDocumentsPaginatedResponse(documents []db.EsignDocument, total int64, perPage, currentPage int32) BaseResponse {
	documentResponses := make([]EsignDocumentResponse, len(documents))
	for i, document := range documents {
		documentResponses[i] = ConvertEsignDocumentToResponse(document)
	}
	return NewPaginatedResponse(documentResponses, total, perPage, currentPage)
}

// EsignTemplateListResponse is purely for Swagger documentation
type EsignTemplateListResponse struct {
	Data        []EsignTemplateResponse `json:"data"`
	Total       int64                   `json:"total" example:"42"`
	PerPage     int32                   `json:"perPage" example:"10"`
	CurrentPage int32                   `json:"currentPage" example:"1"`
	LastPage    int32                   `json:"lastPage" example:"5"`
}

// EsignDocumentListResponse is purely for Swagger documentation
type EsignDocumentListResponse struct {
	Data        []EsignDocumentResponse `json:"data"`
	Total       int64                   `json:"total" example:"42"`
	PerPage     int32                   `json:"perPage" example:"10"`
	CurrentPage int32                   `json:"currentPage" example:"1"`
	LastPage    int32                   `json:"lastPage" example:"5"`
}
