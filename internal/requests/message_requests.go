package requests

import (
	"github.com/google/uuid"
)

// MessageIDParam is used to validate message ID in URL parameters
type MessageIDParam struct {
	ID uuid.UUID `query:"id" validate:"required"`
}

// ListMessagesRequest is used to validate list messages request parameters
type ListMessagesRequest struct {
	PaginationQuery
	MarinaID   uuid.UUID `query:"marinaId" validate:"required"`
	CustomerID string    `query:"customerId" validate:"required"`
	Pinned     *bool     `query:"pinned"`
}

// ListMessagesByCustomerRequest is used to validate list messages by customer request parameters
type ListMessagesByCustomerRequest struct {
	PaginationQuery
	MarinaID   uuid.UUID `query:"marinaId" validate:"required"`
	CustomerID string    `query:"customerId" validate:"required"`
	Pinned     *bool     `query:"pinned"`
}

// CreateMessageRequest is used to validate create message request body
type CreateMessageRequest struct {
	MarinaID   uuid.UUID `json:"marinaId" validate:"required"`
	CustomerID string    `json:"customerId" validate:"required"`
	Type       string    `json:"type" validate:"required,oneof=email sms internal"`
	Body       string    `json:"body" validate:"required"`
	Sender     string    `json:"sender" validate:"required"`
	Recipient  string    `json:"recipient" validate:"required"`
	Contact    string    `json:"contact" validate:"required"`
	Pinned     bool      `json:"pinned"`
}

// UpdateCustomerMessageRequest is used to validate update customer message request body
type UpdateCustomerMessageRequest struct {
	ID         uuid.UUID `json:"id" validate:"required"`
	MarinaID   uuid.UUID `json:"marinaId" validate:"required"`
	CustomerID string    `json:"customerId" validate:"required"`
	Body       string    `json:"body" validate:"required"`
	Pinned     bool      `json:"pinned"`
}

// UpdateMarinaMessageRequest is used to validate update marina message request body
type UpdateMarinaMessageRequest struct {
	ID         uuid.UUID `json:"id" validate:"required"`
	MarinaID   uuid.UUID `json:"marinaId" validate:"required"`
	CustomerID string    `json:"customerId" validate:"required"`
	Body       string    `json:"body" validate:"required"`
	Pinned     bool      `json:"pinned"`
}

// DeleteCustomerMessageRequest is used to validate delete customer message request parameters
type DeleteCustomerMessageRequest struct {
	ID         uuid.UUID `query:"id" validate:"required"`
	MarinaID   uuid.UUID `query:"marinaId" validate:"required"`
	CustomerID string    `query:"customerId" validate:"required"`
}

// DeleteMarinaMessageRequest is used to validate delete marina message request parameters
type DeleteMarinaMessageRequest struct {
	ID         uuid.UUID `query:"id" validate:"required"`
	MarinaID   uuid.UUID `query:"marinaId" validate:"required"`
	CustomerID string    `query:"customerId" validate:"required"`
}

// BedrockRewriteRequest is used to validate Bedrock API rewrite request body
type BedrockRewriteRequest struct {
	Draft             string `json:"draft" validate:"required"`
	MarinaName        string `json:"marinaName" validate:"required"`
	UserName          string `json:"userName" validate:"required"`
	Tone              string `json:"tone" validate:"required"`
	ExtraInstructions string `json:"extraInstructions"`
}

type BedrockDetectFormFieldsRequest struct {
	Document   string `json:"document" validate:"required"`
	PageNumber int    `json:"pageNumber" validate:"required"`
}
