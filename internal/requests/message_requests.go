package requests

import (
	"github.com/google/uuid"
)

// MessageIDParam is used to validate message ID in URL parameters
type MessageIDParam struct {
	MessageID uuid.UUID `query:"messageId" validate:"required"`
}

// ListMessagesRequest is used to validate list messages request parameters
type ListMessagesRequest struct {
	PaginationQuery
	MarinaID   uuid.UUID `query:"marinaId" validate:"required"`
	CustomerID string    `query:"customerId" validate:"required"`
}

// ListMessagesByCustomerRequest is used to validate list messages by customer request parameters
type ListMessagesByCustomerRequest struct {
	PaginationQuery
	MarinaID   uuid.UUID `query:"marinaId" validate:"required"`
	CustomerID string    `query:"customerId" validate:"required"`
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
}

// DeleteMessageRequest is used to validate delete message request parameters
type DeleteMessageRequest struct {
	MessageID  uuid.UUID `query:"messageId" validate:"required"`
	MarinaID   uuid.UUID `query:"marinaId" validate:"required"`
	CustomerID string    `query:"customerId" validate:"required"`
}
