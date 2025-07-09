package requests

import (
	"github.com/google/uuid"
)

// CreateEsignTemplateRequest represents the required parameters to create a new e-signature template
type CreateEsignTemplateRequest struct {
	OrganizationID uuid.UUID  `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	MarinaID       *uuid.UUID `json:"marinaId,omitempty" example:"550e8400-e29b-41d4-a716-446655440001"`
	Name           string     `json:"name" validate:"required" example:"Customer Agreement Template"`
	Description    *string    `json:"description,omitempty" example:"Standard customer agreement template for marina services"`
	Type           string     `json:"type" validate:"required" example:"agreement"`
	Status         string     `json:"status" validate:"required,oneof=draft active archived" example:"draft"`
	// BlobURL        string     `json:"blobUrl" validate:"required" example:"https://s3.amazonaws.com/bucket/templates/agreement.pdf"`
	// BlobMetadata   json.RawMessage `json:"blobMetadata,omitempty" example:"{\"size\": 1024, \"contentType\": \"application/pdf\"}"`
}

// UpdateEsignTemplateRequest represents the parameters that can be updated for an e-signature template
type UpdateEsignTemplateRequest struct {
	Name        string  `json:"name" validate:"required" example:"Updated Customer Agreement Template"`
	Description *string `json:"description,omitempty" example:"Updated description for the template"`
	Type        string  `json:"type" validate:"required" example:"agreement"`
	Status      string  `json:"status" validate:"required,oneof=draft active archived" example:"active"`
	// BlobURL     string  `json:"blobUrl" validate:"required" example:"https://s3.amazonaws.com/bucket/templates/agreement-v2.pdf"`
	// BlobMetadata json.RawMessage `json:"blobMetadata,omitempty" example:"{\"size\": 2048, \"contentType\": \"application/pdf\"}"`
}

// UpdateEsignTemplateStatusRequest represents the parameters to update only the status of a template
type UpdateEsignTemplateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft active archived" example:"active"`
}

// ListEsignTemplatesByMarinaRequest represents the parameters to list templates for a marina with pagination
type ListEsignTemplatesByMarinaRequest struct {
	OrganizationID uuid.UUID `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	MarinaID       uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Limit          int32     `json:"limit" validate:"min=1,max=100" example:"10"`
	Offset         int32     `json:"offset" validate:"min=0" example:"0"`
}

// CreateEsignDocumentRequest represents the required parameters to create a new e-signature document
type CreateEsignDocumentRequest struct {
	TemplateID     *uuid.UUID `json:"templateId,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	OrganizationID uuid.UUID  `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	MarinaID       uuid.UUID  `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Type           string     `json:"type" validate:"required" example:"agreement"`
	Status         string     `json:"status" validate:"required,oneof=draft signed questions sent" example:"draft"`
	// BlobURL        string     `json:"blobUrl" validate:"required" example:"https://s3.amazonaws.com/bucket/documents/agreement-001.pdf"`
	// BlobMetadata   json.RawMessage `json:"blobMetadata,omitempty" example:"{\"size\": 1024, \"contentType\": \"application/pdf\"}"`
}

// UpdateEsignDocumentRequest represents the parameters that can be updated for an e-signature document
type UpdateEsignDocumentRequest struct {
	Type   string `json:"type" validate:"required" example:"agreement"`
	Status string `json:"status" validate:"required,oneof=draft signed questions sent" example:"sent"`
	// BlobURL      string          `json:"blobUrl" validate:"required" example:"https://s3.amazonaws.com/bucket/documents/agreement-001-updated.pdf"`
	// BlobMetadata json.RawMessage `json:"blobMetadata,omitempty" example:"{\"size\": 2048, \"contentType\": \"application/pdf\"}"`
}

// UpdateEsignDocumentStatusRequest represents the parameters to update only the status of a document
type UpdateEsignDocumentStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=draft signed questions sent" example:"sent"`
}

// ListEsignDocumentsByMarinaRequest represents the parameters to list documents for a marina with pagination
type ListEsignDocumentsByMarinaRequest struct {
	OrganizationID uuid.UUID `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	MarinaID       uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Limit          int32     `json:"limit" validate:"min=1,max=100" example:"10"`
	Offset         int32     `json:"offset" validate:"min=0" example:"0"`
}

// ListEsignDocumentsByTemplateRequest represents the parameters to list documents created from a template
type ListEsignDocumentsByTemplateRequest struct {
	TemplateID uuid.UUID `json:"templateId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440002"`
	Limit      int32     `json:"limit" validate:"min=1,max=100" example:"10"`
	Offset     int32     `json:"offset" validate:"min=0" example:"0"`
}

// CreateEsignSubmissionRequest represents the required parameters to create a new e-signature submission
type CreateEsignSubmissionRequest struct {
	DocumentID uuid.UUID `json:"documentId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440002"`
	CustomerID *string   `json:"customerId,omitempty" example:"CUST123"`
	Email      string    `json:"email" validate:"required,email" example:"customer@example.com"`
}

// UpdateEsignSubmissionRequest represents the parameters that can be updated for an e-signature submission
type UpdateEsignSubmissionRequest struct {
	Status     string  `json:"status" validate:"required,oneof=pending signed questions sent" example:"signed"`
	CustomerID *string `json:"customerId,omitempty" example:"CUST123"`
	Email      string  `json:"email" validate:"required,email" example:"customer@example.com"`
	// BlobURL      string          `json:"blobUrl" validate:"required" example:"https://s3.amazonaws.com/bucket/submissions/submission-001-updated.pdf"`
	// BlobMetadata json.RawMessage `json:"blobMetadata,omitempty" example:"{\"size\": 2048, \"contentType\": \"application/pdf\"}"`
}

// ListEsignSubmissionsByMarinaRequest represents the parameters to list submissions for a marina with pagination
type ListEsignSubmissionsByMarinaRequest struct {
	OrganizationID uuid.UUID `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	MarinaID       uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Limit          int32     `json:"limit" validate:"min=1,max=100" example:"10"`
	Offset         int32     `json:"offset" validate:"min=0" example:"0"`
}

// ListEsignSubmissionsByDocumentRequest represents the parameters to list submissions for a document with pagination
type ListEsignSubmissionsByDocumentRequest struct {
	DocumentID uuid.UUID `json:"documentId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440002"`
	Limit      int32     `json:"limit" validate:"min=1,max=100" example:"10"`
	Offset     int32     `json:"offset" validate:"min=0" example:"0"`
}

// ListEsignSubmissionsByStatusRequest represents the parameters to list submissions by status with pagination
type ListEsignSubmissionsByStatusRequest struct {
	OrganizationID uuid.UUID `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	MarinaID       uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	Status         string    `json:"status" validate:"required,oneof=pending signed questions sent" example:"pending"`
	Limit          int32     `json:"limit" validate:"min=1,max=100" example:"10"`
	Offset         int32     `json:"offset" validate:"min=0" example:"0"`
}

// PublicUpdateEsignSubmissionRequest represents the parameters for public update of an e-signature submission
// Only allows updating status and file
type PublicUpdateEsignSubmissionRequest struct {
	Status string `form:"status" validate:"required,oneof=pending signed questions sent" example:"signed"`
}
