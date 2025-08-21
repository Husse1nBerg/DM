package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/notifications"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/redis/go-redis/v9"
)

type NotificationHandler struct {
	server              *s.Server
	notificationService *notifications.NotificationService
}

func NewNotificationHandler(server *s.Server) *NotificationHandler {
	// Initialize notification service
	notificationService := notifications.NewNotificationService(
		server.DB.Queries(),
		server.Redis,
		server.Logger,
	)

	return &NotificationHandler{
		server:              server,
		notificationService: notificationService,
	}
}

// CreateNotificationHandler creates a new notification
//
//	@Summary		Create notification
//	@Description	Create a new notification for a user
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			notification	body		requests.CreateNotificationRequest	true	"Notification information"
//	@Success		201				{object}	responses.NotificationResponseWrapper "Created notification"
//	@Failure		400				{object}	responses.Error "Bad request"
//	@Failure		401				{object}	responses.Error "Unauthorized"
//	@Failure		500				{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/notification [post]
func (h *NotificationHandler) CreateNotificationHandler(c echo.Context) error {
	// Parse and validate the request body
	req := new(requests.CreateNotificationRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Create the notification with real-time sending enabled
	notification, err := h.notificationService.CreateNotification(c.Request().Context(), *req, true)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.NewNotificationResponseSuccess(*notification)
	return c.JSON(http.StatusCreated, response)
}

// ListNotificationsHandler lists notifications for the current user
//
//	@Summary		List notifications
//	@Description	Get paginated notifications for the current user with filtering, search, and sorting
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Param			unreadOnly	query		bool	false	"Show only unread notifications (legacy)"	default(false)
//	@Param			search		query		string	false	"Global search across title and content"
//	@Param			read		query		bool	false	"Filter by read status (true/false)"
//	@Param			type		query		string	false	"Filter by notification type" Enums(message, invite, system, alert, esign, document, payment)
//	@Param			sortBy		query		string	false	"Sort field" Enums(type, read, created_at, title) default(priority+created_at)
//	@Param			sortOrder	query		string	false	"Sort direction" Enums(asc, desc) default(desc)
//	@Success		200			{object}	responses.NotificationListResponse "Paginated list of notifications"
//	@Failure		401			{object}	responses.Error "Unauthorized"
//	@Failure		500			{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/notification/list [get]
func (h *NotificationHandler) ListNotificationsHandler(c echo.Context) error {
	// Get user info from JWT token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims) //

	// Get current user data from database to get the current marina ID
	queries := h.server.DB.Queries()
	currentUser, err := queries.GetUserByID(c.Request().Context(), claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "User not found").JSON(c)
	}

	// Parse pagination and filter params
	req := new(requests.ListNotificationsRequest)
	if err := c.Bind(req); err != nil {
		req.PaginationQuery.Page = 1
		req.PaginationQuery.PageSize = 10
	}

	// Ensure minimum values to prevent negative offset
	if req.PaginationQuery.Page < 1 {
		req.PaginationQuery.Page = 1
	}
	if req.PaginationQuery.PageSize < 1 {
		req.PaginationQuery.PageSize = 10
	}

	// Extract new filter parameters (similar to esign submissions pattern)
	search := c.QueryParam("search")
	typeFilter := c.QueryParam("type")
	sortBy := c.QueryParam("sortBy")
	sortOrder := c.QueryParam("sortOrder")
	unreadOnly := c.QueryParam("unreadOnly") == "true"

	// Handle read filter
	var readFilter *bool
	if unreadOnly {
		// Maintain compatibility with unreadOnly parameter
		readFilter = &[]bool{false}[0]
	} else if readParam := c.QueryParam("read"); readParam != "" {
		if parsed, err := strconv.ParseBool(readParam); err == nil {
			readFilter = &parsed
		}
	}

	// Set defaults for sorting
	if sortOrder == "" {
		sortOrder = "desc"
	}

	// Calculate offset (ensure it's not negative)
	offset := (req.PaginationQuery.Page - 1) * req.PaginationQuery.PageSize
	if offset < 0 {
		offset = 0
	}

	var notifications []db.Notification

	// Detect if we should use new filtered query or maintain current behavior
	useNewQuery := search != "" || typeFilter != "" || sortBy != "" || (readFilter != nil && !unreadOnly)

	if !useNewQuery {
		// Use current behavior to maintain compatibility
		if unreadOnly {
			notifications, err = h.notificationService.GetUnreadNotifications(
				c.Request().Context(),
				claims.ID,
				claims.OrgId,
				&currentUser.MarinaID,
				req.PaginationQuery.PageSize,
				offset,
			)
		} else {
			notifications, err = h.notificationService.GetUserNotifications(
				c.Request().Context(),
				claims.ID,
				claims.OrgId,
				&currentUser.MarinaID,
				req.PaginationQuery.PageSize,
				offset,
			)
		}

		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}

		// Get actual total count for pagination
		var total int64
		if unreadOnly {
			// Count unread notifications
			marinaID := currentUser.MarinaID
			if marinaID == uuid.Nil {
				marinaID = uuid.UUID{} // Zero UUID for NULL case
			}
			total, err = h.server.DB.Queries().GetUnreadNotificationCount(c.Request().Context(), db.GetUnreadNotificationCountParams{
				UserID:         claims.ID,
				OrganizationID: claims.OrgId,
				Column3:        marinaID,
			})
		} else {
			// Count all notifications
			marinaID := currentUser.MarinaID
			if marinaID == uuid.Nil {
				marinaID = uuid.UUID{} // Zero UUID for NULL case
			}
			total, err = h.server.DB.Queries().CountAllNotifications(c.Request().Context(), db.CountAllNotificationsParams{
				UserID:         claims.ID,
				OrganizationID: claims.OrgId,
				Column3:        marinaID,
			})
		}
		if err != nil {
			h.server.Logger.Zap.Error("Error counting notifications", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error counting notifications").JSON(c)
		}

		return responses.NewNotificationsPaginatedResponse(notifications, total, req.PaginationQuery.PageSize, req.PaginationQuery.Page).JSON(c)
	} else {
		// Use new filtered query
		ctx := c.Request().Context()

		// Handle nullable marina_id for SQLC - use zero UUID for NULL case
		marinaID := currentUser.MarinaID
		if marinaID == uuid.Nil {
			marinaID = uuid.UUID{} // Zero UUID to match '00000000-0000-0000-0000-000000000000'
		}

		// Handle nullable read filter for SQLC - use empty string for NULL case
		readFilterStr := ""
		if readFilter != nil {
			if *readFilter {
				readFilterStr = "true"
			} else {
				readFilterStr = "false"
			}
		}

		notifications, err = h.server.DB.Queries().ListNotificationsWithFilters(ctx, db.ListNotificationsWithFiltersParams{
			UserID:         claims.ID,
			OrganizationID: claims.OrgId,
			Column3:        marinaID,
			Column4:        search,
			Column5:        readFilterStr,
			Column6:        typeFilter,
			Column7:        sortBy,
			Column8:        sortOrder,
			Limit:          req.PaginationQuery.PageSize,
			Offset:         offset,
		})
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}

		total, err := h.server.DB.Queries().CountNotificationsWithFilters(ctx, db.CountNotificationsWithFiltersParams{
			UserID:         claims.ID,
			OrganizationID: claims.OrgId,
			Column3:        marinaID,
			Column4:        search,
			Column5:        readFilterStr,
			Column6:        typeFilter,
		})
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}

		return responses.NewNotificationsPaginatedResponse(notifications, total, req.PaginationQuery.PageSize, req.PaginationQuery.Page).JSON(c)
	}
}

// GetUnreadCountHandler returns the count of unread notifications
//
//	@Summary		Get unread count
//	@Description	Get the count of unread notifications for the current user
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.UnreadCountResponse "Unread notification count"
//	@Failure		401	{object}	responses.Error "Unauthorized"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/notification/unread-count [get]
func (h *NotificationHandler) GetUnreadCountHandler(c echo.Context) error {
	// Get user info from JWT token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	// Get current user data from database to get the current marina ID
	queries := h.server.DB.Queries()
	currentUser, err := queries.GetUserByID(c.Request().Context(), claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "User not found").JSON(c)
	}

	count, err := h.notificationService.GetUnreadCount(
		c.Request().Context(),
		claims.ID,
		claims.OrgId,
		&currentUser.MarinaID,
	)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewUnreadCountResponse(count).JSON(c)
}

// MarkAsReadHandler marks a notification as read
//
//	@Summary		Mark notification as read
//	@Description	Mark a specific notification as read
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Notification ID"
//	@Success		200	{object}	responses.NotificationResponseWrapper "Updated notification"
//	@Failure		400	{object}	responses.Error "Bad request"
//	@Failure		401	{object}	responses.Error "Unauthorized"
//	@Failure		404	{object}	responses.Error "Notification not found"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/notification/{id}/read [put]
func (h *NotificationHandler) MarkAsReadHandler(c echo.Context) error {
	// Get user info from JWT token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	// Parse notification ID from path
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid notification ID").JSON(c)
	}

	// Mark notification as read with real-time update
	notification, err := h.notificationService.MarkNotificationAsRead(
		c.Request().Context(),
		notificationID,
		claims.ID,
		true, // Send real-time update
	)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.NewNotificationResponseSuccess(*notification)
	return response.JSON(c)
}

// MarkAllAsReadHandler marks all notifications as read
//
//	@Summary		Mark all notifications as read
//	@Description	Mark all notifications as read for the current user
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.BaseResponse "Success message"
//	@Failure		401	{object}	responses.Error "Unauthorized"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/notification/mark-all-read [put]
func (h *NotificationHandler) MarkAllAsReadHandler(c echo.Context) error {
	// Get user info from JWT token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	// Get current user data from database to get the current marina ID
	queries := h.server.DB.Queries()
	currentUser, err := queries.GetUserByID(c.Request().Context(), claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "User not found").JSON(c)
	}

	// Mark all notifications as read with real-time update
	err = h.notificationService.MarkAllNotificationsAsRead(
		c.Request().Context(),
		claims.ID,
		claims.OrgId,
		&currentUser.MarinaID,
		true, // Send real-time update
	)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMessageResponse(http.StatusOK, "All notifications marked as read").JSON(c)
}

// GetNotificationsByTypeHandler retrieves notifications by type
//
//	@Summary		Get notifications by type
//	@Description	Get paginated notifications of a specific type for the current user
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			type		path		string	true	"Notification type (message, invite, system, alert)"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Success		200			{object}	responses.NotificationListResponse "Paginated list of notifications"
//	@Failure		400			{object}	responses.Error "Bad request"
//	@Failure		401			{object}	responses.Error "Unauthorized"
//	@Failure		500			{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/notification/type/{type} [get]
func (h *NotificationHandler) GetNotificationsByTypeHandler(c echo.Context) error {
	// Get user info from JWT token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	// Get current user data from database to get the current marina ID
	queries := h.server.DB.Queries()
	currentUser, err := queries.GetUserByID(c.Request().Context(), claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "User not found").JSON(c)
	}

	// Get notification type from path
	notificationType := c.Param("type")
	if notificationType == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Notification type is required").JSON(c)
	}

	// Parse pagination params
	req := new(requests.GetNotificationsByTypeRequest)
	if err := c.Bind(req); err != nil {
		req.PaginationQuery.Page = 1
		req.PaginationQuery.PageSize = 10
	}
	req.Type = notificationType

	// Ensure minimum values to prevent negative offset
	if req.PaginationQuery.Page < 1 {
		req.PaginationQuery.Page = 1
	}
	if req.PaginationQuery.PageSize < 1 {
		req.PaginationQuery.PageSize = 10
	}

	// Calculate offset (ensure it's not negative)
	offset := (req.PaginationQuery.Page - 1) * req.PaginationQuery.PageSize
	if offset < 0 {
		offset = 0
	}

	notifications, err := h.notificationService.GetNotificationsByType(
		c.Request().Context(),
		claims.ID,
		claims.OrgId,
		notificationType,
		&currentUser.MarinaID,
		req.PaginationQuery.PageSize,
		offset,
	)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Get total count for pagination (simplified)
	total := int64(len(notifications))
	if len(notifications) == int(req.PaginationQuery.PageSize) {
		total = int64(req.PaginationQuery.Page * req.PaginationQuery.PageSize)
	}

	return responses.NewNotificationsPaginatedResponse(notifications, total, req.PaginationQuery.PageSize, req.PaginationQuery.Page).JSON(c)
}

// GetNotificationHandler gets a specific notification by ID
//
//	@Summary		Get notification
//	@Description	Get a specific notification by ID
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Notification ID"
//	@Success		200	{object}	responses.NotificationResponseWrapper "Notification details"
//	@Failure		400	{object}	responses.Error "Bad request"
//	@Failure		401	{object}	responses.Error "Unauthorized"
//	@Failure		404	{object}	responses.Error "Notification not found"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/notification/{id} [get]
func (h *NotificationHandler) GetNotificationHandler(c echo.Context) error {
	// Get user info from JWT token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	// Parse notification ID from path
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid notification ID").JSON(c)
	}

	// Get notification from database using the generated query
	queries := h.server.DB.Queries()
	params := db.GetNotificationByIDParams{
		ID:     notificationID,
		UserID: claims.ID,
	}

	notification, err := queries.GetNotificationByID(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Notification not found").JSON(c)
	}

	response := responses.NewNotificationResponseSuccess(notification)
	return response.JSON(c)
}

// DeleteNotificationHandler deletes a notification
//
//	@Summary		Delete notification
//	@Description	Delete a specific notification
//	@Tags			Notification
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Notification ID"
//	@Success		200	{object}	responses.BaseResponse "Success message"
//	@Failure		400	{object}	responses.Error "Bad request"
//	@Failure		401	{object}	responses.Error "Unauthorized"
//	@Failure		404	{object}	responses.Error "Notification not found"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/notification/{id} [delete]
func (h *NotificationHandler) DeleteNotificationHandler(c echo.Context) error {
	// Get user info from JWT token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	// Parse notification ID from path
	notificationIDStr := c.Param("id")
	notificationID, err := uuid.Parse(notificationIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid notification ID").JSON(c)
	}

	// Delete notification from database using the generated query
	queries := h.server.DB.Queries()
	params := db.DeleteNotificationParams{
		ID:     notificationID,
		UserID: claims.ID,
	}

	err = queries.DeleteNotification(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMessageResponse(http.StatusOK, "Notification deleted successfully").JSON(c)
}

// NotificationStreamHandler provides Server-Sent Events stream for real-time notifications
//
//	@Summary		Real-time notification stream
//	@Description	Subscribe to real-time notifications via Server-Sent Events
//	@Tags			Notification
//	@Accept			json
//	@Produce		text/event-stream
//	@Success		200	{string}	string "Server-Sent Events stream"
//	@Failure		401	{object}	responses.Error "Unauthorized"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/notification/stream [get]
func (h *NotificationHandler) NotificationStreamHandler(c echo.Context) error {
	// Get user info from JWT token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	// Get current user data from database to get the current marina ID
	queries := h.server.DB.Queries()
	currentUser, err := queries.GetUserByID(c.Request().Context(), claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "User not found").JSON(c)
	}

	// Set SSE headers
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Cache-Control", "no-cache")
	c.Response().Header().Set("Connection", "keep-alive")
	c.Response().Header().Set("Access-Control-Allow-Origin", "*")
	c.Response().Header().Set("Access-Control-Allow-Headers", "Cache-Control")

	// Create context for the connection
	ctx, cancel := context.WithCancel(c.Request().Context())
	defer cancel()

	// Subscribe to user-specific and marina-wide notification channels
	userChannel := fmt.Sprintf("notifications:user:%s", claims.ID.String())
	marinaChannel := fmt.Sprintf("notifications:marina:%s", currentUser.MarinaID.String())

	pubsub := h.server.Redis.Subscribe(ctx, userChannel, marinaChannel)
	defer pubsub.Close()

	// Send initial connection message
	initialMsg := map[string]interface{}{
		"event":     "connected",
		"message":   "Connected to notification stream",
		"timestamp": time.Now(),
		"channels":  []string{userChannel, marinaChannel},
	}
	if err := h.writeSSEMessage(c, "connected", initialMsg); err != nil {
		h.server.Logger.Zap.Errorw("Failed to send initial SSE message", "error", err, "userID", claims.ID)
		return err
	}

	// Send heartbeat every 30 seconds to keep connection alive
	heartbeatTicker := time.NewTicker(15 * time.Second)
	defer heartbeatTicker.Stop()

	h.server.Logger.Zap.Infow("SSE connection established", "userID", claims.ID, "userChannel", userChannel, "marinaChannel", marinaChannel)

	for {
		select {
		case <-ctx.Done():
			h.server.Logger.Zap.Infow("SSE connection closed", "userID", claims.ID)
			return nil

		case <-heartbeatTicker.C:
			// Send heartbeat to keep connection alive
			heartbeat := map[string]interface{}{
				"event":     "heartbeat",
				"timestamp": time.Now(),
			}
			if err := h.writeSSEMessage(c, "heartbeat", heartbeat); err != nil {
				h.server.Logger.Zap.Warnw("Failed to send heartbeat", "error", err, "userID", claims.ID)
				return nil
			}

		default:
			// Check for new messages from Redis with timeout
			msg, err := pubsub.ReceiveTimeout(ctx, 1*time.Second)
			if err != nil {
				// Timeout is expected, continue to next iteration
				continue
			}

			switch m := msg.(type) {
			case *redis.Message:
				// Parse the notification data
				var notificationData map[string]interface{}
				if err := json.Unmarshal([]byte(m.Payload), &notificationData); err != nil {
					h.server.Logger.Zap.Warnw("Failed to parse notification data", "error", err, "payload", m.Payload)
					continue
				}

				// Determine event type
				eventType := "notification"
				if event, ok := notificationData["event"].(string); ok {
					eventType = event
				}

				// Send the notification via SSE
				if err := h.writeSSEMessage(c, eventType, notificationData); err != nil {
					h.server.Logger.Zap.Warnw("Failed to send SSE notification", "error", err, "userID", claims.ID)
					return nil
				}

				h.server.Logger.Zap.Debugw("SSE notification sent", "userID", claims.ID, "eventType", eventType, "channel", m.Channel)

			case *redis.Subscription:
				// Log subscription events
				h.server.Logger.Zap.Debugw("Redis subscription event", "kind", m.Kind, "channel", m.Channel, "count", m.Count)
			}
		}
	}
}

// writeSSEMessage writes a Server-Sent Event message to the response
func (h *NotificationHandler) writeSSEMessage(c echo.Context, eventType string, data interface{}) error {
	// Convert data to JSON
	jsonData, err := json.Marshal(data)
	if err != nil {
		return fmt.Errorf("failed to marshal SSE data: %w", err)
	}

	// Write SSE format: event: eventType\ndata: jsonData\n\n
	response := c.Response()
	if _, err := fmt.Fprintf(response, "event: %s\n", eventType); err != nil {
		return err
	}
	if _, err := fmt.Fprintf(response, "data: %s\n\n", string(jsonData)); err != nil {
		return err
	}

	// Flush the response to send immediately
	response.Flush()
	return nil
}
