package responses

import (
	"net/http"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// MessageResponse represents a single message
// @Description Message data including customer, type, direction, and status information
type MessageResponse struct {
	ID         uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	MarinaID   uuid.UUID  `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440001"`
	CustomerID string     `json:"customerId" example:"CUST123"`
	Type       string     `json:"type" example:"email"`
	Direction  string     `json:"direction" example:"to_customer"`
	Body       string     `json:"body" example:"Your reservation has been confirmed"`
	Sender     string     `json:"sender" example:"John Doe"`
	Recipient  string     `json:"recipient" example:"Jane Smith"`
	Contact    string     `json:"contact" example:"jane@example.com"`
	Status     string     `json:"status" example:"sent"`
	Pinned     bool       `json:"pinned" example:"false"`
	CreatedAt  time.Time  `json:"createdAt"`
	UpdatedAt  *time.Time `json:"updatedAt,omitempty"`
}

// MessageResponseWrapper wraps a message response
// @Description Wrapper for a single message response
type MessageResponseWrapper struct {
	Data MessageResponse `json:"data"`
}

// MessageListResponse represents a paginated list of messages
// @Description Paginated list of messages
type MessageListResponse struct {
	Data        []MessageResponse `json:"data"`
	Total       int64             `json:"total" example:"42"`
	PerPage     int32             `json:"perPage" example:"10"`
	CurrentPage int32             `json:"currentPage" example:"1"`
	LastPage    int32             `json:"lastPage" example:"5"`
}

// BedrockRewriteResponse represents the response from the Bedrock API rewrite
// @Description Response from the Bedrock API rewrite
type BedrockRewriteResponse struct {
	Message string `json:"message"`
	Tokens  struct {
		Input  int `json:"input"`
		Output int `json:"output"`
	} `json:"tokens"`
}

// NewMessageResponseSuccess creates a new successful message response
func NewMessageResponseSuccess(message db.Message) BaseResponse {
	return NewSuccessResponse(MessageDBToResponse(message))
}

// MessageDBToResponse converts a db.Message to a MessageResponse
func MessageDBToResponse(message db.Message) MessageResponse {
	var updatedAt *time.Time
	if message.UpdatedAt.Valid {
		t := message.UpdatedAt.Time
		updatedAt = &t
	}

	return MessageResponse{
		ID:         message.ID,
		MarinaID:   message.MarinaID,
		CustomerID: message.CustomerID,
		Type:       message.Type,
		Direction:  message.Direction,
		Body:       message.Body,
		Sender:     message.Sender,
		Recipient:  message.Recipient,
		Contact:    message.Contact,
		Status:     message.Status,
		Pinned:     message.Pinned,
		CreatedAt:  message.CreatedAt.Time,
		UpdatedAt:  updatedAt,
	}
}

// NewMessagesPaginatedResponse creates a paginated response of messages
func NewMessagesPaginatedResponse(messages []db.Message, total int64, perPage, page int32) BaseResponse {
	messageResponses := make([]MessageResponse, len(messages))
	for i, message := range messages {
		messageResponses[i] = MessageDBToResponse(message)
	}

	return NewPaginatedResponse(messageResponses, total, perPage, page)
}

// JSON returns the JSON response
func (r *MessageResponseWrapper) JSON(c echo.Context) error {
	return c.JSON(http.StatusOK, r)
}

// JSON returns the JSON response
func (r *MessageListResponse) JSON(c echo.Context) error {
	return c.JSON(http.StatusOK, r)
}
