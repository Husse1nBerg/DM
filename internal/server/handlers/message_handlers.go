package handlers

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"os"

	"bytes"
	"io/ioutil"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/notifications"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
	"github.com/dockworks/dm-web-backend/pkg/telgorithm"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type MessageHandler struct {
	server              *s.Server
	notificationService *notifications.NotificationService
}

func NewMessageHandler(server *s.Server) *MessageHandler {
	// Initialize notification service
	notificationService := notifications.NewNotificationService(
		server.DB.Queries(),
		server.Redis,
		server.Logger,
		server.SendGrid,
		server.Config,
	)

	return &MessageHandler{
		server:              server,
		notificationService: notificationService,
	}
}

// checkMessageLimit checks if the marina has reached its message usage limit
func (h *MessageHandler) checkMessageLimit(ctx echo.Context, marinaID uuid.UUID, messageType string) error {
	// Get marina to check current message usage
	marina, err := h.server.DB.Queries().GetMarinaByID(ctx.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina", err)
		return fmt.Errorf("marina not found")
	}

	// Get marina's notes messages plan
	notesMessagesPlan, err := h.server.DB.Queries().GetMarinaNotesMessagesPlan(ctx.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching notes messages plan", err)
		return fmt.Errorf("error fetching notes messages plan")
	}

	// Check if we've reached the limit
	currentUsage := int32(0)
	maxLimit := notesMessagesPlan.TextLimit

	if messageType == "sms" {
		if marina.TextUsage != nil {
			currentUsage = int32(*marina.TextUsage)
		}
	} else if messageType == "email" {
		if marina.EmailUsage != nil {
			currentUsage = int32(*marina.EmailUsage)
		}
	}

	// If maxLimit is nil, it means unlimited
	if maxLimit == nil {
		return nil
	}

	if currentUsage >= *maxLimit {
		h.server.Logger.Zap.Info("Limit would be exceeded",
			"currentUsage", currentUsage,
			"maxLimit", *maxLimit)

		return fmt.Errorf("limit exceeded: current usage %d has reached the %s limit of %d",
			currentUsage, messageType, *maxLimit)
	}

	return nil
}

// updateMessageUsage updates the marina's message usage count
func (h *MessageHandler) updateUsage(ctx context.Context, marinaID uuid.UUID, messageType string) error {
	increment := int16(1)

	// Increment the usage count
	if messageType == "sms" {
		_, err := h.server.DB.Queries().IncrementMarinaTextUsage(ctx, db.IncrementMarinaTextUsageParams{
			ID:      marinaID,
			Column2: increment,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error updating marina text usage", err)
			return err
		}
	} else if messageType == "email" {
		_, err := h.server.DB.Queries().IncrementMarinaEmailUsage(ctx, db.IncrementMarinaEmailUsageParams{
			ID:      marinaID,
			Column2: increment,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error updating marina email usage", err)
			return err
		}
	}

	return nil
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
	cfg := h.server.Config

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

	// Check message limit before sending
	// if err := h.checkMessageLimit(c, req.MarinaID, "email"); err != nil {
	// 	logger.Zap.Error("Message limit check failed", err)
	// 	return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	// }

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
		HomeURL:   cfg.App.HomeURL(),
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

	// Create notification for marina staff about new customer message
	// Find marina users to notify
	marinaUsers, err := queries.GetUsersByMarina(c.Request().Context(), db.GetUsersByMarinaParams{
		MarinaID:   req.MarinaID,
		IsCustomer: utils.Pointer(false), // Get marina staff, not customers
	})
	if err != nil {
		logger.Zap.Warnw("Failed to get marina users for notification", "marina_id", req.MarinaID, "error", err)
	} else {
		// Create notifications for marina staff using smart notification system
		emailData := &notifications.EmailNotificationData{
			To:      []string{req.Contact},
			Subject: "New Customer Message",
		}

		results, err := h.notificationService.CreateBulkMessageNotifications(
			c.Request().Context(),
			marinaUsers,
			req.MarinaID, // Using marina ID as organization ID for now
			req.MarinaID,
			req.Body,       // Message content preview
			req.Sender,     // Customer name
			req.CustomerID, // Customer ID
			emailData,
		)
		if err != nil {
			logger.Zap.Warnw("Failed to create bulk message notifications", "error", err)
		} else {
			// Log notification results
			for _, result := range results {
				if len(result.Errors) > 0 {
					logger.Zap.Warnw("Notification delivery had errors",
						"user_id", result.UserID,
						"errors", result.Errors)
				} else {
					logger.Zap.Infow("Notification delivered successfully",
						"user_id", result.UserID,
						"system", result.SystemDelivered,
						"email", result.EmailDelivered)
				}
			}
		}
	}

	// Process the result asynchronously to log success/failure and update message usage
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
			// Increment message usage count
			if err := h.updateUsage(context.Background(), req.MarinaID, "email"); err != nil {
				logger.Zap.Errorw("Failed to update message usage",
					"marina_id", req.MarinaID,
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

// CreateMessageMarinaHandler creates a new message
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
	cfg := h.server.Config
	req := new(requests.CreateMessageRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	// Check message limit before sending
	// if req.Type != "email" {
	// 	if err := h.checkMessageLimit(c, req.MarinaID, req.Type); err != nil {
	// 		logger.Zap.Error("Message limit check failed", err)
	// 		return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	// 	}
	// }

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

		// Process the result asynchronously to log success/failure and update message usage
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
				// Increment message usage count
				if err := h.updateUsage(context.Background(), req.MarinaID, "sms"); err != nil {
					logger.Zap.Errorw("Failed to update message usage",
						"marina_id", req.MarinaID,
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
		var taskID uuid.UUID
		var resultChan <-chan sendgrid.EmailStatus

		// Create email data
		email := sendgrid.MessageTemplateData{
			Content:   req.Body,
			Recipient: req.Recipient,
			Sender:    req.Sender,
			HomeURL:   cfg.App.HomeURL(),
		}
		to := []string{req.Contact}
		subject := "Message from " + req.Sender

		// Send email asynchronously
		taskID, resultChan, err = h.server.SendGrid.SendMessageEmail(to, subject, email)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}

		// Log the task
		logger.Zap.Infow("Email queued", "task_id", taskID.String(), "to", req.Recipient)

		// Process the result asynchronously to log success/failure and update message usage
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
				// Increment message usage count
				if err := h.updateUsage(context.Background(), req.MarinaID, "email"); err != nil {
					logger.Zap.Errorw("Failed to update message usage",
						"marina_id", req.MarinaID,
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
		// For internal messages, just mark as sent
		if err := queries.UpdateMessageStatus(c.Request().Context(), db.UpdateMessageStatusParams{
			ID:     message.ID,
			Status: "sent",
		}); err != nil {
			logger.Zap.Errorw("Failed to update message status",
				"message_id", message.ID,
				"error", err)
		}
	}
	// Get the user id of the customer
	customers, err := queries.GetMarinaCustomerUsersByCustomerID(c.Request().Context(), db.GetMarinaCustomerUsersByCustomerIDParams{
		MarinaID:   req.MarinaID,
		CustomerID: &req.CustomerID,
	})
	if err != nil {
		logger.Zap.Errorw("Failed to get customer users", "error", err)
	}
	if len(customers) > 0 {
		// Create notification for customer users about new marina message using smart notification system
		var emailData *notifications.EmailNotificationData

		// Prepare delivery data based on message type
		if req.Type == "email" {
			emailData = &notifications.EmailNotificationData{
				To:      []string{req.Contact},
				Subject: "Message from " + req.Sender,
			}
		}

		results, err := h.notificationService.CreateBulkMessageNotificationsForCustomers(
			c.Request().Context(),
			customers,
			req.MarinaID, // Using marina ID as organization ID for now
			req.MarinaID,
			req.Body,
			req.Sender,
			req.CustomerID,
			emailData,
		)
		if err != nil {
			logger.Zap.Errorw("Failed to create bulk message notifications for customers", "error", err)
		} else {
			// Log notification results
			for _, result := range results {
				if len(result.Errors) > 0 {
					logger.Zap.Warnw("Customer notification delivery had errors",
						"user_id", result.UserID,
						"errors", result.Errors)
				} else {
					logger.Zap.Infow("Customer notification delivered successfully",
						"user_id", result.UserID,
						"system", result.SystemDelivered,
						"email", result.EmailDelivered)
				}
			}
		}
	}
	response := responses.NewMessageResponseSuccess(message)
	return c.JSON(http.StatusCreated, response)
}

// RewriteMessageHandler handles rewriting customer-facing messages using Bedrock API
// @Summary Rewrite customer-facing message
// @Description Rewrite a draft message to be more professional using Bedrock API
// @Tags Message
// @Accept json
// @Produce json
// @Param rewrite body requests.BedrockRewriteRequest true "Rewrite request"
// @Success 200 {object} responses.BedrockRewriteResponse "Rewritten message"
// @Failure 400 {object} responses.Error "Bad request"
// @Failure 500 {object} responses.Error "Server error"
// @Router /message/compose-message [post]
func (h *MessageHandler) RewriteMessageHandler(c echo.Context) error {
	logger := h.server.Logger

	// Parse and validate request
	req := new(requests.BedrockRewriteRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Prepare Bedrock API request
	bedrockRequest := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{"text": req.Draft},
					{"text": "The marina name is " + req.MarinaName + " and the staff sender is " + req.UserName + ". Use a " + req.Tone + " tone."},
					{"text": req.ExtraInstructions},
					{"text": "You rewrite customer-facing messages for a marina. Be concise, clear, and polite. Focus on the main message without additional formalities. Insert the marina name and the staff sender name in the signature if not present. Preserve facts; do not invent details. Output text only."},
				},
			},
		},
	}

	// Prepare HTTP request to Bedrock API
	client := &http.Client{}
	reqBody, err := json.Marshal(bedrockRequest)
	if err != nil {
		logger.Zap.Errorw("Failed to marshal request body", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to prepare request").JSON(c)
	}

	bedrockReq, err := http.NewRequest("POST", "https://bedrock-runtime.us-east-1.amazonaws.com/model/amazon.nova-lite-v1:0/invoke", bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Zap.Errorw("Failed to create new HTTP request", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to create request").JSON(c)
	}

	bedrockReq.Header.Set("Authorization", "Bearer "+os.Getenv("AWS_BEARER_TOKEN_BEDROCK"))
	bedrockReq.Header.Set("Content-Type", "application/json")
	bedrockReq.Header.Set("Accept", "application/json")

	// Send request to Bedrock API
	resp, err := client.Do(bedrockReq)
	if err != nil {
		logger.Zap.Errorw("Failed to send request to Bedrock API", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to send request").JSON(c)
	}
	defer resp.Body.Close()

	// Parse response from Bedrock API
	var apiResponse struct {
		Output struct {
			Message struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
				Role string `json:"role"`
			} `json:"message"`
		} `json:"output"`
		StopReason string `json:"stopReason"`
		Usage      struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
			TotalTokens  int `json:"totalTokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		logger.Zap.Errorw("Failed to decode response from Bedrock API", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to decode response").JSON(c)
	}

	// Map the parsed response to your response structure
	bedrockResponse := responses.BedrockRewriteResponse{
		Message: apiResponse.Output.Message.Content[0].Text,
		Tokens: struct {
			Input  int `json:"input"`
			Output int `json:"output"`
		}{
			Input:  apiResponse.Usage.InputTokens,
			Output: apiResponse.Usage.OutputTokens,
		},
	}

	// Return response to client
	return c.JSON(http.StatusOK, bedrockResponse)
}

// DetectFormFieldsHandler detects form fields in a PDF page image
//
// @Summary Detect form fields
// @Description Detect form fields in a PDF page image and return structured JSON
// @Tags Message
// @Accept json
// @Produce json
// @Param request body requests.BedrockDetectFormFieldsRequest true "PDF page image"
// @Success 200 {object} responses.BedrockDetectFormFieldsResponse "Detected form fields"
// @Failure 400 {object} responses.Error "Bad request"
// @Failure 500 {object} responses.Error "Server error"
// @Router /message/detect-form-fields [post]
func (h *MessageHandler) DetectFormFieldsHandler(c echo.Context) error {
	logger := h.server.Logger

	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		h.server.Logger.Zap.Error("Error getting file", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "File is required").JSON(c)
	}
	defer file.Close()

	// Read file content into a byte slice
	fileBytes, err := ioutil.ReadAll(file)
	if err != nil {
		logger.Zap.Errorw("Failed to read file content", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to read file content").JSON(c)
	}

	// Prepare Bedrock API request
	bedrockRequest := map[string]interface{}{
		"messages": []map[string]interface{}{
			{
				"role": "user",
				"content": []map[string]interface{}{
					{
						"document": map[string]interface{}{
							"name":   header.Filename,
							"format": "pdf",
							"source": map[string]interface{}{
								"bytes": base64.StdEncoding.EncodeToString(fileBytes),
							},
						},
					},
					{
						"text": "Detect form fields in the PDF and return a valid JSON object. Follow these strict rules:\n\n1. Schema:\n{\n  \"pages\": [{\n    \"pageNumber\": integer,\n    \"pageSize\": {\n      \"widthPx\": integer,\n      \"heightPx\": integer\n    },\n    \"fields\": [{\n      \"id\": string,\n      \"type\": string (one of: text|textarea|checkbox|radio|signature|date|initials|email|phone|number),\n      \"label\": string or null,\n      \"required\": boolean,\n      \"confidence\": number between 0-1,\n      \"bbox\": {\n        \"x\": integer,\n        \"y\": integer,\n        \"w\": integer,\n        \"h\": integer\n      },\n      \"bboxNorm\": {\n        \"x\": number between 0-1,\n        \"y\": number between 0-1,\n        \"w\": number between 0-1,\n        \"h\": number between 0-1\n      }\n    }]\n  }]\n}\n\n2. Rules:\n- Use double quotes for all JSON keys and string values\n- All numbers must be valid JSON numbers (no spaces, valid decimals)\n- Coordinates origin is top-left\n- Normalized coordinates must be calculated by dividing by width/height\n- Set required=true only for fields with asterisk or explicit required indicator\n- Use radio type only for mutually exclusive option groups\n- Exclude decorative elements\n- Return only the JSON object, no markdown or commentary",
					},
				},
			},
		},
	}

	// Prepare HTTP request to Bedrock API
	client := &http.Client{}
	reqBody, err := json.Marshal(bedrockRequest)
	if err != nil {
		logger.Zap.Errorw("Failed to marshal request body", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to prepare request").JSON(c)
	}

	bedrockReq, err := http.NewRequest("POST", "https://bedrock-runtime.us-east-1.amazonaws.com/model/amazon.nova-lite-v1:0/invoke", bytes.NewBuffer(reqBody))
	if err != nil {
		logger.Zap.Errorw("Failed to create new HTTP request", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to create request").JSON(c)
	}

	bedrockReq.Header.Set("Authorization", "Bearer "+os.Getenv("AWS_BEARER_TOKEN_BEDROCK"))
	bedrockReq.Header.Set("Content-Type", "application/json")
	bedrockReq.Header.Set("Accept", "application/json")

	// Send request to Bedrock API
	resp, err := client.Do(bedrockReq)
	if err != nil {
		logger.Zap.Errorw("Failed to send request to Bedrock API", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to send request").JSON(c)
	}
	defer resp.Body.Close()

	// Parse response from Bedrock API
	var apiResponse struct {
		Output struct {
			Message struct {
				Content []struct {
					Text string `json:"text"`
				} `json:"content"`
				Role string `json:"role"`
			} `json:"message"`
		} `json:"output"`
		StopReason string `json:"stopReason"`
		Usage      struct {
			InputTokens  int `json:"inputTokens"`
			OutputTokens int `json:"outputTokens"`
			TotalTokens  int `json:"totalTokens"`
		} `json:"usage"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&apiResponse); err != nil {
		logger.Zap.Errorw("Failed to decode response from Bedrock API", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to decode response").JSON(c)
	}

	bedrockResponse := responses.BedrockRewriteResponse{
		Message: apiResponse.Output.Message.Content[0].Text,
		Tokens: struct {
			Input  int `json:"input"`
			Output int `json:"output"`
		}{
			Input:  apiResponse.Usage.InputTokens,
			Output: apiResponse.Usage.OutputTokens,
		},
	}

	// Return the parsed JSON
	return c.JSON(http.StatusOK, bedrockResponse)

	// return c.JSON(http.StatusOK, apiResponse)
}

// GetMessageByIDHandler gets a specific message by ID
//
//	@Summary		Get message
//	@Description	Get a specific message by ID
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			id	query		string	true	"Message ID"
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
		ID:         req.ID,
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

// UpdateCustomerMessageHandler updates a customer message
//
//	@Summary		Update customer message
//	@Description	Update an existing customer message
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			message	body		requests.UpdateCustomerMessageRequest	true	"Message information"
//	@Success		200		{object}	responses.MessageResponseWrapper "Updated message"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		404		{object}	responses.Error "Message not found"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/message/customer [put]
func (h *MessageHandler) UpdateCustomerMessageHandler(c echo.Context) error {
	// Parse and validate request
	logger := h.server.Logger

	req := new(requests.UpdateCustomerMessageRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	// Update the message
	params := db.UpdateMessageParams{
		Body:       req.Body,
		Pinned:     req.Pinned,
		ID:         req.ID,
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
	}

	message, err := queries.UpdateMessage(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Message not found").JSON(c)
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

	response := responses.NewMessageResponseSuccess(message)
	return response.JSON(c)
}

// UpdateMarinaMessageHandler updates a marina message
//
//	@Summary		Update marina message
//	@Description	Update an existing marina message
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			message	body		requests.UpdateMarinaMessageRequest	true	"Message information"
//	@Success		200		{object}	responses.MessageResponseWrapper "Updated message"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		404		{object}	responses.Error "Message not found"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/message/marina [put]
func (h *MessageHandler) UpdateMarinaMessageHandler(c echo.Context) error {
	// Parse and validate request
	logger := h.server.Logger

	req := new(requests.UpdateMarinaMessageRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	// Update the message
	params := db.UpdateMessageParams{
		Body:       req.Body,
		Pinned:     req.Pinned,
		ID:         req.ID,
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
	}

	message, err := queries.UpdateMessage(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Message not found").JSON(c)
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

	response := responses.NewMessageResponseSuccess(message)
	return response.JSON(c)
}

// DeleteCustomerMessageHandler deletes a customer message
//
//	@Summary		Delete customer message
//	@Description	Delete an existing customer message
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			id	query		string	true	"Message ID"
//	@Param			marinaId	query		string	true	"Marina ID"
//	@Param			customerId	query		string	true	"Customer ID"
//	@Success		200		{object}	responses.MessageResponseWrapper "Message deleted"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		404		{object}	responses.Error "Message not found"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/message/customer [delete]
func (h *MessageHandler) DeleteCustomerMessageHandler(c echo.Context) error {
	// Parse and validate request
	req := new(requests.DeleteCustomerMessageRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	// First check if message exists
	params := db.GetMessageByIDParams{
		ID:         req.ID,
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
	}

	_, err := queries.GetMessageByID(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Message not found").JSON(c)
	}

	// Delete the message
	deleteParams := db.DeleteMessageParams{
		ID:         req.ID,
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
	}

	err = queries.DeleteMessage(c.Request().Context(), deleteParams)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return c.NoContent(http.StatusOK)
}

// DeleteMarinaMessageHandler deletes a marina message
//
//	@Summary		Delete marina message
//	@Description	Delete an existing marina message
//	@Tags			Message
//	@Accept			json
//	@Produce		json
//	@Param			id	query		string	true	"Message ID"
//	@Param			marinaId	query		string	true	"Marina ID"
//	@Param			customerId	query		string	true	"Customer ID"
//	@Success		200		{object}	responses.MessageResponseWrapper "Message deleted"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		404		{object}	responses.Error "Message not found"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/message/marina [delete]
func (h *MessageHandler) DeleteMarinaMessageHandler(c echo.Context) error {
	// Parse and validate request
	req := new(requests.DeleteMarinaMessageRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	// First check if message exists
	params := db.GetMessageByIDParams{
		ID:         req.ID,
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
	}

	_, err := queries.GetMessageByID(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Message not found").JSON(c)
	}

	// Delete the message
	deleteParams := db.DeleteMessageParams{
		ID:         req.ID,
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
	}

	err = queries.DeleteMessage(c.Request().Context(), deleteParams)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return c.NoContent(http.StatusOK)
}
