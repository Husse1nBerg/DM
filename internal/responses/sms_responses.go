package responses

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// SMSSendResponse represents a response after an SMS has been queued for delivery
type SMSSendResponse struct {
	TaskID     uuid.UUID `json:"task_id"`
	Message    string    `json:"message"`
	Success    bool      `json:"success"`
	StatusCode int       `json:"-"`
}

// NewSMSSendResponse creates a new SMS send response
func NewSMSSendResponse(taskID uuid.UUID, message string, success bool) *SMSSendResponse {
	return &SMSSendResponse{
		TaskID:     taskID,
		Message:    message,
		Success:    success,
		StatusCode: http.StatusOK,
	}
}

// JSON returns the response as JSON
func (r *SMSSendResponse) JSON(c echo.Context) error {
	return c.JSON(r.StatusCode, r)
}

// SMSStatusDetails represents the status details of an individual SMS message
type SMSStatusDetails struct {
	ID           uuid.UUID `json:"id"`
	Status       string    `json:"status"`
	MessageID    string    `json:"message_id,omitempty"`
	Error        string    `json:"error,omitempty"`
	SegmentCount int       `json:"segment_count,omitempty"`
	Recipient    string    `json:"recipient,omitempty"`
}

// BatchSMSSendResponse represents a response after a batch of SMS messages has been queued
type BatchSMSSendResponse struct {
	BatchID      uuid.UUID          `json:"batch_id"`
	Message      string             `json:"message"`
	Success      bool               `json:"success"`
	SuccessCount int                `json:"success_count"`
	FailureCount int                `json:"failure_count"`
	TotalCount   int                `json:"total_count"`
	Details      []SMSStatusDetails `json:"details,omitempty"`
	StatusCode   int                `json:"-"`
}

// NewBatchSMSSendResponse creates a new batch SMS send response
func NewBatchSMSSendResponse(
	batchID uuid.UUID,
	message string,
	success bool,
	successCount int,
	failureCount int,
	totalCount int,
	details []SMSStatusDetails,
) *BatchSMSSendResponse {
	return &BatchSMSSendResponse{
		BatchID:      batchID,
		Message:      message,
		Success:      success,
		SuccessCount: successCount,
		FailureCount: failureCount,
		TotalCount:   totalCount,
		Details:      details,
		StatusCode:   http.StatusOK,
	}
}

// JSON returns the response as JSON
func (r *BatchSMSSendResponse) JSON(c echo.Context) error {
	return c.JSON(r.StatusCode, r)
}
