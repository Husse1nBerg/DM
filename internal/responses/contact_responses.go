package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// ContactResponse represents a single contact response
// @Description Contact response model
// @Schema responses.ContactResponse
type ContactResponse struct {
	ID          uuid.UUID  `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	MarinaID    uuid.UUID  `json:"marinaId" example:"123e4567-e89b-12d3-a456-426614174000"`
	Type        string     `json:"type" example:"EMERGENCY"`
	Name        string     `json:"name" example:"John Doe"`
	Description *string    `json:"description,omitempty" example:"Primary emergency contact"`
	Email       *string    `json:"email,omitempty" example:"john@example.com"`
	Phone       *string    `json:"phone,omitempty" example:"+1-555-123-4567"`
	IsCPContact bool       `json:"is_cp_contact" example:"false"`
	CreatedAt   *time.Time `json:"createdAt,omitempty" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   *time.Time `json:"updatedAt,omitempty" example:"2023-01-01T00:00:00Z"`
}

// ContactListResponse represents a list of contacts response
// @Description Contact list response model
// @Schema responses.ContactListResponse
type ContactListResponse struct {
	BaseResponse
	Data []ContactResponse `json:"data"`
}

// ContactListPaginatedResponse is for Swagger and paginated responses
// @Description Contact list paginated response model
// @Schema responses.ContactListPaginatedResponse
type ContactListPaginatedResponse struct {
	Data        []ContactResponse `json:"data"`
	Total       int64             `json:"total" example:"42"`
	PerPage     int32             `json:"perPage" example:"10"`
	CurrentPage int32             `json:"currentPage" example:"1"`
	LastPage    int32             `json:"lastPage" example:"5"`
}

// ConvertContactToResponse converts a database contact to a response model
func ConvertContactToResponse(contact db.Contact) ContactResponse {
	return ContactResponse{
		ID:          contact.ID,
		MarinaID:    contact.MarinaID,
		Type:        contact.Type,
		Name:        contact.Name,
		Description: contact.Description,
		Email:       contact.Email,
		Phone:       contact.Phone,
		IsCPContact: contact.IsCpContact != nil && *contact.IsCpContact,
		CreatedAt:   utils.PgTimeToTimePtr(contact.CreatedAt),
		UpdatedAt:   utils.PgTimeToTimePtr(contact.UpdatedAt),
	}
}

// NewContactResponse creates a new contact response
func NewContactResponse(contact db.Contact) BaseResponse {
	return NewSuccessResponse(ConvertContactToResponse(contact))
}

// NewContactListResponse creates a new contact list response
func NewContactListResponse(contacts []db.Contact) BaseResponse {
	contactResponses := make([]ContactResponse, len(contacts))
	for i, contact := range contacts {
		contactResponses[i] = ConvertContactToResponse(contact)
	}

	return NewSuccessResponse(contactResponses)
}

// NewContactListPaginatedResponse creates a paginated response for contacts
func NewContactListPaginatedResponse(contacts []db.Contact, total int64, perPage, currentPage int32) BaseResponse {
	contactResponses := make([]ContactResponse, len(contacts))
	for i, contact := range contacts {
		contactResponses[i] = ConvertContactToResponse(contact)
	}
	return NewPaginatedResponse(contactResponses, total, perPage, currentPage)
}
