package handlers

import (
	"context"
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
	"github.com/dockworks/dm-web-backend/pkg/telgorithm"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MessageHandler struct {
	server *s.Server
}

func NewMessageHandler(server *s.Server) *MessageHandler {
	return &MessageHandler{server: server}
}

// ListMessagesHandler lists messages for a marina and customer
//
//	@Summary		List messages
//	@Description	Get messages for a marina and customer
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			marinaId		query		string	true	"Marina ID"
//	@Param			customerId		query		string	true	"Customer ID"
//	@Param			page			query		int		false	"Page number"	default(1)
//	@Param			pageSize		query		int		false	"Page size"		default(10)
//	@Success		200	{object}	responses.MessageListResponse "Paginated list of messages"
//	@Failure		400	{object}	responses.Error "Bad request"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/message/marina [get]
func (h *MessageHandler) ListMessagesMarinaHandler(c echo.Context) error {
	// Parse and validate request
	req := new(requests.ListMessagesRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Set default pagination if not provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	queries := h.server.DB.Queries()

	// Get messages
	params := db.ListMessagesParams{
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
		Limit:      int32(req.PageSize),
		Offset:     int32((req.Page - 1) * req.PageSize),
	}

	messages, err := queries.ListMessages(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	allParams := db.ListMessagesAllParams{
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
	}
	allMessages, err := queries.ListMessagesAll(c.Request().Context(), allParams)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	var total int64
	if allMessages == nil {
		total = 0
	} else {
		total = int64(len(allMessages))
	}

	return responses.NewMessagesPaginatedResponse(messages, total, int32(params.Limit), int32(req.Page)).JSON(c)
}

// ListMessagesByCustomerHandler lists messages by type (email, sms) for a marina and customer
//
//	@Summary		List customer messages
//	@Description	Get messages of type email or sms for a marina and customer
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			marinaId		query		string	true	"Marina ID"
//	@Param			customerId		query		string	true	"Customer ID"
//	@Param			page			query		int		false	"Page number"	default(1)
//	@Param			pageSize		query		int		false	"Page size"		default(10)
//	@Success		200	{object}	responses.MessageListResponse "Paginated list of customer messages"
//	@Failure		400	{object}	responses.Error "Bad request"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/message/customer [get]
func (h *MessageHandler) ListMessagesCustomerHandler(c echo.Context) error {
	// Parse and validate request
	req := new(requests.ListMessagesByCustomerRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Set default pagination if not provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	queries := h.server.DB.Queries()

	// Get messages
	params := db.ListMessagesByCustomerParams{
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
		Limit:      int32(req.PageSize),
		Offset:     int32((req.Page - 1) * req.PageSize),
	}

	messages, err := queries.ListMessagesByCustomer(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Count total messages for pagination
	// In a real application, you might want to implement a count query
	allParams := db.ListMessagesByCustomerAllParams{
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
	}
	allMessages, err := queries.ListMessagesByCustomerAll(c.Request().Context(), allParams)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	// if allMessages is nil, set total to 0
	var total int64
	if allMessages == nil {
		total = 0
	} else {
		total = int64(len(allMessages))
	}

	return responses.NewMessagesPaginatedResponse(messages, total, int32(params.Limit), int32(req.Page)).JSON(c)
}

// CreateMessageHandler creates a new message
//
//	@Summary		Create message
//	@Description	Create a new message
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			message	body		requests.CreateMessageRequest	true	"Message information"
//	@Success		201		{object}	responses.MessageResponseWrapper "Created message"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/message/customer [post]
func (h *MessageHandler) CreateMessageHandler(c echo.Context) error {
	// Parse and validate request
	logger := h.server.Logger

	req := new(requests.CreateMessageRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	if req.Type == "internal" || req.Type == "sms" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Internal and SMS messages are not allowed for customers").JSON(c)
	}

	// Create the message
	params := db.CreateMessageParams{
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
		Type:       "email",
		Direction:  "to_marina",
		Body:       req.Body,
		Sender:     req.Sender,
		Recipient:  req.Recipient,
		Contact:    req.Contact,
		Status:     "pending",
		Pinned:     req.Pinned,
	}

	message, err := queries.CreateMessage(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// If the message is pinned, update other messages' pinned status
	if req.Pinned {
		err = queries.UpdateOtherMessagesPinnedStatus(c.Request().Context(), db.UpdateOtherMessagesPinnedStatusParams{
			MarinaID:   req.MarinaID,
			CustomerID: req.CustomerID,
			ID:         message.ID,
		})
		if err != nil {
			logger.Zap.Errorw("Failed to update other messages' pinned status",
				"message_id", message.ID,
				"error", err)
		}
	}

	// Create email data
	email := sendgrid.MessageTemplateData{
		Content:   req.Body,
		Recipient: req.Recipient,
		Sender:    req.Sender,
	}
	to := []string{req.Contact}
	subject := "Message from " + req.Sender

	// Send email asynchronously
	taskID, resultChan, err := h.server.SendGrid.SendMessageEmail(to, subject, email)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Log the task
	logger.Zap.Infow("Email queued", "task_id", taskID.String(), "to", req.Recipient)

	// Process the result asynchronously to log success/failure
	go func() {
		result := <-resultChan
		if result.Status == telgorithm.StatusSent {
			logger.Zap.Infow("Email sent successfully",
				"to", req.Contact,
				"task_id", result.ID.String(),
				"message_id", result.ID,
				"status", result.Status)
			// Use background context for DB update
			if err := queries.UpdateMessageStatus(context.Background(), db.UpdateMessageStatusParams{
				ID:     message.ID,
				Status: "sent",
			}); err != nil {
				logger.Zap.Errorw("Failed to update message status",
					"message_id", message.ID,
					"error", err)
			}
		} else {
			logger.Zap.Errorw("Failed to send email",
				"to", req.Contact,
				"task_id", result.ID.String(),
				"error", result.Error)
			// Use background context for DB update
			if err := queries.UpdateMessageStatus(context.Background(), db.UpdateMessageStatusParams{
				ID:     message.ID,
				Status: "failed",
			}); err != nil {
				logger.Zap.Errorw("Failed to update message status",
					"message_id", message.ID,
					"error", err)
			}
		}
	}()

	response := responses.NewMessageResponseSuccess(message)
	return c.JSON(http.StatusCreated, response)
}

// CreateMessageHandler creates a new message
//
//	@Summary		Create message
//	@Description	Create a new message
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			message	body		requests.CreateMessageRequest	true	"Message information"
//	@Success		201		{object}	responses.MessageResponseWrapper "Created message"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/message/marina [post]
func (h *MessageHandler) CreateMessageMarinaHandler(c echo.Context) error {
	// Parse and validate request
	logger := h.server.Logger

	req := new(requests.CreateMessageRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	var direction string
	if req.Type == "internal" {
		direction = "internal"
	} else if req.Type == "sms" {
		direction = "to_customer"
	} else {
		direction = "to_customer"
	}
	// Create the message
	params := db.CreateMessageParams{
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
		Type:       req.Type,
		Direction:  direction,
		Body:       req.Body,
		Sender:     req.Sender,
		Recipient:  req.Recipient,
		Contact:    req.Contact,
		Status:     "pending",
		Pinned:     req.Pinned,
	}

	message, err := queries.CreateMessage(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// If the message is pinned, update other messages' pinned status
	if req.Pinned {
		err = queries.UpdateOtherMessagesPinnedStatus(c.Request().Context(), db.UpdateOtherMessagesPinnedStatusParams{
			MarinaID:   req.MarinaID,
			CustomerID: req.CustomerID,
			ID:         message.ID,
		})
		if err != nil {
			logger.Zap.Errorw("Failed to update other messages' pinned status",
				"message_id", message.ID,
				"error", err)
		}
	}

	if req.Type == "sms" {
		to := []string{req.Contact}
		sms := &telgorithm.SMSData{
			To:      to,
			Message: req.Body,
		}
		// Send SMS asynchronously
		taskID, resultChan := h.server.Telgorithm.SendSMS(sms)

		// Log the task
		logger.Zap.Infow("SMS queued", "task_id", taskID.String(), "to", req.Contact)

		// Process the result asynchronously to log success/failure
		go func() {
			result := <-resultChan
			if result.Status == telgorithm.StatusSent {
				logger.Zap.Infow("SMS sent successfully",
					"to", req.Contact,
					"task_id", result.ID.String(),
					"message_id", result.MessageID,
					"segment_count", result.SegmentCount)
				// Use background context for DB update
				if err := queries.UpdateMessageStatus(context.Background(), db.UpdateMessageStatusParams{
					ID:     message.ID,
					Status: "sent",
				}); err != nil {
					logger.Zap.Errorw("Failed to update message status",
						"message_id", message.ID,
						"error", err)
				}
			} else {
				logger.Zap.Errorw("Failed to send SMS",
					"to", req.Contact,
					"task_id", result.ID.String(),
					"error", result.Error)
				// Use background context for DB update
				if err := queries.UpdateMessageStatus(context.Background(), db.UpdateMessageStatusParams{
					ID:     message.ID,
					Status: "failed",
				}); err != nil {
					logger.Zap.Errorw("Failed to update message status",
						"message_id", message.ID,
						"error", err)
				}
			}
		}()

	} else if req.Type == "email" {

		// Create email data
		email := sendgrid.MessageTemplateData{
			Content:   req.Body,
			Recipient: req.Recipient,
			Sender:    req.Sender,
		}
		to := []string{req.Contact}
		subject := "Message from " + req.Sender

		// Send email asynchronously
		taskID, resultChan, err := h.server.SendGrid.SendMessageEmail(to, subject, email)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}

		// Log the task
		logger.Zap.Infow("Email queued", "task_id", taskID.String(), "to", req.Recipient)

		// Process the result asynchronously to log success/failure
		go func() {
			result := <-resultChan
			if result.Status == telgorithm.StatusSent {
				logger.Zap.Infow("Email sent successfully",
					"to", req.Contact,
					"task_id", result.ID.String(),
					"message_id", result.ID,
					"status", result.Status)
				// Use background context for DB update
				if err := queries.UpdateMessageStatus(context.Background(), db.UpdateMessageStatusParams{
					ID:     message.ID,
					Status: "sent",
				}); err != nil {
					logger.Zap.Errorw("Failed to update message status",
						"message_id", message.ID,
						"error", err)
				}
			} else {
				logger.Zap.Errorw("Failed to send email",
					"to", req.Contact,
					"task_id", result.ID.String(),
					"error", result.Error)
				// Use background context for DB update
				if err := queries.UpdateMessageStatus(context.Background(), db.UpdateMessageStatusParams{
					ID:     message.ID,
					Status: "failed",
				}); err != nil {
					logger.Zap.Errorw("Failed to update message status",
						"message_id", message.ID,
						"error", err)
				}
			}
		}()

	} else {
		queries.UpdateMessageStatus(c.Request().Context(), db.UpdateMessageStatusParams{ID: message.ID, Status: "sent"})
	}

	response := responses.NewMessageResponseSuccess(message)
	return c.JSON(http.StatusCreated, response)
}

// GetMessageByIDHandler gets a specific message by ID
//
//	@Summary		Get message
//	@Description	Get a specific message by ID
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			messageId	query		string	true	"Message ID"
//	@Param			marinaId	query		string	true	"Marina ID"
//	@Param			customerId	query		string	true	"Customer ID"
//	@Success		200		{object}	responses.MessageResponseWrapper "Message details"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		404		{object}	responses.Error "Message not found"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/message/get [get]
func (h *MessageHandler) GetMessageByIDHandler(c echo.Context) error {
	// Parse and validate request
	req := new(struct {
		requests.MessageIDParam
		MarinaID   string `query:"marinaId" validate:"required,uuid4"`
		CustomerID string `query:"customerId" validate:"required,uuid4"`
	})
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	marinaID, err := uuid.Parse(req.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID").JSON(c)
	}

	// Get the message
	params := db.GetMessageByIDParams{
		ID:         req.MessageID,
		MarinaID:   marinaID,
		CustomerID: req.CustomerID,
	}

	message, err := queries.GetMessageByID(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Message not found").JSON(c)
	}

	response := responses.NewMessageResponseSuccess(message)
	return response.JSON(c)
}
