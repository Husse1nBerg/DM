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
