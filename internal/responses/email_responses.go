package responses

import (
	"net/http"

	"github.com/google/uuid"
)

// EmailSendResponse defines the response from email sending endpoints
// @Description Email send response
// @Schema responses.EmailSendResponse
type EmailSendResponse struct {
	BaseResponse
	Data EmailSendResponseData `json:"data"`
}

// EmailSendResponseData contains the data for the email send response
type EmailSendResponseData struct {
	TaskID   uuid.UUID `json:"taskId"`   // Task ID for tracking
	Message  string    `json:"message"`  // Status message
	Accepted bool      `json:"accepted"` // Whether the email was accepted for delivery
}

// NewEmailSendResponse creates a new email send response
func NewEmailSendResponse(taskID uuid.UUID, message string, accepted bool) *EmailSendResponse {
	return &EmailSendResponse{
		BaseResponse: BaseResponse{
			Code:    http.StatusOK,
			Message: message,
		},
		Data: EmailSendResponseData{
			TaskID:   taskID,
			Message:  message,
			Accepted: accepted,
		},
	}
}
