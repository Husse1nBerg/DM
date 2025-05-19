package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	"github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/telgorithm"
	"github.com/labstack/echo/v4"
)

// SMSHandler handles SMS-related requests
type SMSHandler struct {
	server *server.Server
}

// NewSMSHandler creates a new SMSHandler
func NewSMSHandler(srv *server.Server, cfg *config.Config) *SMSHandler {
	return &SMSHandler{
		server: srv,
	}
}

// SendSMS sends an SMS message
// @Summary Send SMS
// @Description Send SMS message to one or more recipients
// @Tags SMS
// @Accept json
// @Produce json
// @Param params body requests.SendSMSRequest true "SMS details"
// @Success 200 {object} responses.SMSSendResponse "SMS accepted for delivery"
// @Failure 400 {object} responses.Error "Validation error"
// @Failure 500 {object} responses.Error "Server error"
// @Router /sms/send [post]
func (h *SMSHandler) SendSMS(c echo.Context) error {
	logger := h.server.Logger

	// Bind and validate request
	req := new(requests.SendSMSRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Create SMS data
	sms := &telgorithm.SMSData{
		To:        req.To,
		Message:   req.Message,
		MediaURLs: req.MediaURLs,
		Priority:  req.Priority,
		ExpiresOn: req.ExpiresOn,
	}

	// Only set FromNumber if specified in the request
	if req.FromNumber != "" {
		sms.FromNumber = req.FromNumber
	}

	// Send SMS asynchronously
	taskID, resultChan := h.server.Telgorithm.SendSMS(sms)

	// Log the task
	logger.Zap.Infow("SMS queued", "task_id", taskID.String(), "to", req.To)

	// Process the result asynchronously to log success/failure
	go func() {
		result := <-resultChan
		if result.Status == telgorithm.StatusSent {
			logger.Zap.Infow("SMS sent successfully",
				"to", req.To,
				"task_id", result.ID.String(),
				"message_id", result.MessageID,
				"segment_count", result.SegmentCount)
		} else {
			logger.Zap.Errorw("Failed to send SMS",
				"to", req.To,
				"task_id", result.ID.String(),
				"error", result.Error)
		}
	}()

	// Return response to client
	return responses.NewSMSSendResponse(
		taskID,
		"SMS accepted for delivery",
		true,
	).JSON(c)
}

// SendBatchSMS sends multiple SMS messages in a batch
// @Summary Send batch SMS
// @Description Send multiple SMS messages in a batch
// @Tags SMS
// @Accept json
// @Produce json
// @Param params body requests.BatchSMSRequest true "Batch SMS details"
// @Success 200 {object} responses.BatchSMSSendResponse "Batch SMS accepted for delivery"
// @Failure 400 {object} responses.Error "Validation error"
// @Failure 500 {object} responses.Error "Server error"
// @Router /sms/send-batch [post]
func (h *SMSHandler) SendBatchSMS(c echo.Context) error {
	logger := h.server.Logger

	// Bind and validate request
	req := new(requests.BatchSMSRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Prepare the batch data
	batch := &telgorithm.BatchSMSData{
		Messages:   make([]telgorithm.SMSData, 0, len(req.Messages)),
		FromNumber: req.FromNumber,
	}

	// Convert each message
	for _, msg := range req.Messages {
		smsData := telgorithm.SMSData{
			To:        msg.To,
			Message:   msg.Message,
			MediaURLs: msg.MediaURLs,
			Priority:  msg.Priority,
			ExpiresOn: msg.ExpiresOn,
		}
		batch.Messages = append(batch.Messages, smsData)
	}

	// Send batch asynchronously
	batchID, resultChan := h.server.Telgorithm.SendBatchSMS(batch)

	// Log the batch task
	logger.Zap.Infow("Batch SMS queued",
		"batch_id", batchID.String(),
		"message_count", len(req.Messages))

	// Process the result asynchronously
	go func() {
		result := <-resultChan

		// Convert to status details
		details := make([]responses.SMSStatusDetails, 0, len(result.Results))

		for _, status := range result.Results {
			details = append(details, responses.SMSStatusDetails{
				ID:           status.ID,
				Status:       status.Status,
				MessageID:    status.MessageID,
				Error:        status.Error,
				SegmentCount: status.SegmentCount,
				Recipient:    status.Recipient,
			})
		}

		// Log final status
		if result.FailureCount > 0 {
			if result.SuccessCount > 0 {
				logger.Zap.Warnw("Batch SMS completed with partial success",
					"batch_id", result.TaskID.String(),
					"success_count", result.SuccessCount,
					"failure_count", result.FailureCount,
					"total_count", len(result.Results))
			} else {
				logger.Zap.Errorw("Batch SMS failed completely",
					"batch_id", result.TaskID.String(),
					"failure_count", result.FailureCount,
					"total_count", len(result.Results))
			}
		} else {
			logger.Zap.Infow("Batch SMS completed successfully",
				"batch_id", result.TaskID.String(),
				"success_count", result.SuccessCount,
				"total_count", len(result.Results))
		}
	}()

	// Return response to client with initial status
	return responses.NewBatchSMSSendResponse(
		batchID,
		"Batch SMS accepted for delivery",
		true,
		0, // Success count will be updated asynchronously
		0, // Failure count will be updated asynchronously
		len(req.Messages),
		nil, // Details will be available in the async result
	).JSON(c)
}
