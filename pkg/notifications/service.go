package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/dockworks/dm-web-backend/pkg/redis"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// NotificationService provides notification management functionality
type NotificationService struct {
	db       *db.Queries
	redis    *redis.Client
	logger   *logger.Logger
	sendgrid *sendgrid.Client
	Config   *config.Config
}

// NotificationEvent represents different types of notification events
type NotificationEvent struct {
	Type           string                    `json:"type"`
	UserID         uuid.UUID                 `json:"userId"`
	OrganizationID uuid.UUID                 `json:"organizationId"`
	MarinaID       uuid.UUID                 `json:"marinaId"`
	Title          string                    `json:"title"`
	Content        string                    `json:"content"`
	Data           requests.NotificationData `json:"data,omitempty"`
	Priority       string                    `json:"priority"`
}

// NewNotificationService creates a new notification service
func NewNotificationService(database *db.Queries, redisClient *redis.Client, logger *logger.Logger, sendgrid *sendgrid.Client, config *config.Config) *NotificationService {
	return &NotificationService{
		db:       database,
		redis:    redisClient,
		logger:   logger,
		sendgrid: sendgrid,
		Config:   config,
	}
}

// CreateNotification creates a new notification and optionally sends it in real-time
func (s *NotificationService) CreateNotificationService(ctx context.Context, req requests.CreateNotificationRequest, sendRealTime bool) (*db.Notification, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Set default priority if not provided
	priority := "normal"
	if req.Priority != nil {
		priority = *req.Priority
	}

	// Convert data to JSONB
	var dataBytes []byte
	if req.Data != nil {
		var err error
		dataBytes, err = json.Marshal(req.Data)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal notification data: %w", err)
		}
	}

	// Create notification in database
	params := db.CreateNotificationParams{
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		MarinaID:       req.MarinaID,
		Type:           req.Type,
		Title:          req.Title,
		Content:        req.Content,
		Data:           dataBytes,
		Priority:       &priority,
	}

	notification, err := s.db.CreateNotification(ctx, params)
	if err != nil {
		// Check if it's a context timeout error
		if ctx.Err() == context.DeadlineExceeded {
			s.logger.Zap.Errorw("Notification creation timed out",
				"userID", req.UserID,
				"type", req.Type,
				"error", err)
			return nil, fmt.Errorf("notification creation timed out: %w", err)
		}

		s.logger.Zap.Errorw("Failed to create notification",
			"error", err,
			"userID", req.UserID,
			"type", req.Type)
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.Zap.Infow("Notification created",
		"notificationID", notification.ID,
		"userID", req.UserID,
		"type", req.Type)

	// Verify the notification was actually saved by trying to retrieve it
	// This helps identify if there are transaction rollback issues
	// Skip verification under time pressure to improve performance
	if deadline, ok := ctx.Deadline(); ok {
		remaining := time.Until(deadline)
		if remaining < 30*time.Second {
			s.logger.Zap.Debugw("Skipping notification verification due to time pressure",
				"notificationID", notification.ID,
				"userID", req.UserID,
				"remaining_time", remaining)
		} else {
			verificationCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
			defer cancel()

			_, verifyErr := s.db.GetNotificationByID(verificationCtx, db.GetNotificationByIDParams{
				ID:     notification.ID,
				UserID: req.UserID,
			})
			if verifyErr != nil {
				s.logger.Zap.Errorw("Notification verification failed - notification may not have been committed",
					"notificationID", notification.ID,
					"userID", req.UserID,
					"type", req.Type,
					"error", verifyErr)
				// Don't fail the operation, but log the warning
			} else {
				s.logger.Zap.Debugw("Notification verified in database",
					"notificationID", notification.ID,
					"userID", req.UserID,
					"type", req.Type)
			}
		}
	} else {
		// No deadline, always verify
		verificationCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()

		_, verifyErr := s.db.GetNotificationByID(verificationCtx, db.GetNotificationByIDParams{
			ID:     notification.ID,
			UserID: req.UserID,
		})
		if verifyErr != nil {
			s.logger.Zap.Errorw("Notification verification failed - notification may not have been committed",
				"notificationID", notification.ID,
				"userID", req.UserID,
				"type", req.Type,
				"error", verifyErr)
			// Don't fail the operation, but log the warning
		} else {
			s.logger.Zap.Debugw("Notification verified in database",
				"notificationID", notification.ID,
				"userID", req.UserID,
				"type", req.Type)
		}
	}

	// Send real-time notification if requested
	// if sendRealTime {
	// 	if err := s.SendRealTimeNotification(ctx, notification); err != nil {
	// 		s.logger.Zap.Warnw("Failed to send real-time notification",
	// 			"error", err,
	// 			"notificationID", notification.ID)
	// 		// Don't fail the entire operation if real-time sending fails
	// 	}
	// }

	return &notification, nil
}

// SendRealTimeNotification sends a notification via Redis Pub/Sub for real-time delivery
func (s *NotificationService) SendRealTimeNotification(ctx context.Context, notification db.Notification) error {
	// Get unread count for the user
	unreadCount, err := s.GetUnreadCount(ctx, notification.UserID, notification.OrganizationID, &notification.MarinaID)
	if err != nil {
		s.logger.Zap.Warnw("Failed to get unread count", "error", err, "userID", notification.UserID)
		unreadCount = 0 // Continue with 0 count
	}

	// Create real-time notification payload
	event := map[string]interface{}{
		"event":        "notification",
		"notification": notification,
		"unreadCount":  unreadCount,
		"timestamp":    time.Now(),
	}

	// Publish to user-specific channel
	userChannel := fmt.Sprintf("notifications:user:%s", notification.UserID.String())
	if err := s.redis.Publish(ctx, userChannel, event); err != nil {
		return fmt.Errorf("failed to publish to user channel: %w", err)
	}

	// Publish to marina-wide channel
	marinaChannel := fmt.Sprintf("notifications:marina:%s", notification.MarinaID.String())
	if err := s.redis.Publish(ctx, marinaChannel, event); err != nil {
		s.logger.Zap.Warnw("Failed to publish to marina channel", "error", err, "marinaID", notification.MarinaID)
		// Don't fail if marina channel fails
	}

	s.logger.Zap.Debugw("Real-time notification sent", "notificationID", notification.ID, "userChannel", userChannel)
	return nil
}

// GetUserNotifications retrieves paginated notifications for a user
func (s *NotificationService) GetUserNotifications(ctx context.Context, userID, organizationID uuid.UUID, marinaID *uuid.UUID, limit, offset int32) ([]db.Notification, error) {
	params := db.ListNotificationsParams{
		UserID:         userID,
		OrganizationID: organizationID,
		Limit:          limit,
		Offset:         offset,
	}

	// Handle optional marina ID - we need to pass a valid UUID even if it's nullable
	if marinaID != nil {
		params.Column3 = *marinaID
	} else {
		// Pass a nil UUID - the SQL query will handle this with the NULL check
		params.Column3 = uuid.Nil
	}

	notifications, err := s.db.ListNotifications(ctx, params)
	if err != nil {
		s.logger.Zap.Errorw("Failed to get notifications", "error", err, "userID", userID)
		return nil, fmt.Errorf("failed to get notifications: %w", err)
	}

	return notifications, nil
}

// GetUnreadNotifications retrieves unread notifications for a user
func (s *NotificationService) GetUnreadNotifications(ctx context.Context, userID, organizationID uuid.UUID, marinaID *uuid.UUID, limit, offset int32) ([]db.Notification, error) {
	params := db.ListUnreadNotificationsParams{
		UserID:         userID,
		OrganizationID: organizationID,
		Limit:          limit,
		Offset:         offset,
	}

	// Handle optional marina ID - we need to pass a valid UUID even if it's nullable
	if marinaID != nil {
		params.Column3 = *marinaID
	} else {
		// Pass a nil UUID - the SQL query will handle this with the NULL check
		params.Column3 = uuid.Nil
	}

	notifications, err := s.db.ListUnreadNotifications(ctx, params)
	if err != nil {
		s.logger.Zap.Errorw("Failed to get unread notifications", "error", err, "userID", userID)
		return nil, fmt.Errorf("failed to get unread notifications: %w", err)
	}

	return notifications, nil
}

// GetUnreadCount returns the count of unread notifications for a user
func (s *NotificationService) GetUnreadCount(ctx context.Context, userID, organizationID uuid.UUID, marinaID *uuid.UUID) (int64, error) {
	params := db.GetUnreadNotificationCountParams{
		UserID:         userID,
		OrganizationID: organizationID,
	}

	// Handle optional marina ID - we need to pass a valid UUID even if it's nullable
	if marinaID != nil {
		params.Column3 = *marinaID
	} else {
		// Pass a nil UUID - the SQL query will handle this with the NULL check
		params.Column3 = uuid.Nil
	}

	count, err := s.db.GetUnreadNotificationCount(ctx, params)
	if err != nil {
		s.logger.Zap.Errorw("Failed to get unread count", "error", err, "userID", userID)
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}

	return count, nil
}

// MarkNotificationAsRead marks a specific notification as read
func (s *NotificationService) MarkNotificationAsRead(ctx context.Context, notificationID, userID uuid.UUID, sendRealTime bool) (*db.Notification, error) {
	params := db.MarkNotificationAsReadParams{
		ID:     notificationID,
		UserID: userID,
	}

	notification, err := s.db.MarkNotificationAsRead(ctx, params)
	if err != nil {
		s.logger.Zap.Errorw("Failed to mark notification as read", "error", err, "notificationID", notificationID, "userID", userID)
		return nil, fmt.Errorf("failed to mark notification as read: %w", err)
	}

	s.logger.Zap.Infow("Notification marked as read", "notificationID", notificationID, "userID", userID)

	// Send real-time update if requested
	if sendRealTime {
		if err := s.SendReadStatusUpdate(ctx, notification); err != nil {
			s.logger.Zap.Warnw("Failed to send read status update", "error", err, "notificationID", notificationID)
		}
	}

	return &notification, nil
}

func (s *NotificationService) MarkNotificationAsUnread(ctx context.Context, notificationID, userID uuid.UUID, sendRealTime bool) (*db.Notification, error) {
	params := db.MarkNotificationAsUnreadParams{
		ID:     notificationID,
		UserID: userID,
	}

	notification, err := s.db.MarkNotificationAsUnread(ctx, params)
	if err != nil {
		s.logger.Zap.Errorw("Failed to mark notification as unread", "error", err, "notificationID", notificationID, "userID", userID)
		return nil, fmt.Errorf("failed to mark notification as unread: %w", err)
	}

	s.logger.Zap.Infow("Notification marked as unread", "notificationID", notificationID, "userID", userID)

	// Send real-time update if requested
	if sendRealTime {
		if err := s.SendUnreadCountUpdate(ctx, userID, notification.OrganizationID, &notification.MarinaID); err != nil {
			s.logger.Zap.Warnw("Failed to send unread count update", "error", err, "notificationID", notificationID)
		}
	}

	return &notification, nil
}

// MarkAllNotificationsAsRead marks all notifications as read for a user
func (s *NotificationService) MarkAllNotificationsAsRead(ctx context.Context, userID, organizationID uuid.UUID, marinaID *uuid.UUID, sendRealTime bool) error {
	params := db.MarkAllNotificationsAsReadParams{
		UserID:         userID,
		OrganizationID: organizationID,
	}

	// Handle optional marina ID - we need to pass a valid UUID even if it's nullable
	if marinaID != nil {
		params.Column3 = *marinaID
	} else {
		// Pass a nil UUID - the SQL query will handle this with the NULL check
		params.Column3 = uuid.Nil
	}

	err := s.db.MarkAllNotificationsAsRead(ctx, params)
	if err != nil {
		s.logger.Zap.Errorw("Failed to mark all notifications as read", "error", err, "userID", userID)
		return fmt.Errorf("failed to mark all notifications as read: %w", err)
	}

	s.logger.Zap.Infow("All notifications marked as read", "userID", userID)

	// Send real-time update if requested
	if sendRealTime {
		if err := s.SendUnreadCountUpdate(ctx, userID, organizationID, marinaID); err != nil {
			s.logger.Zap.Warnw("Failed to send unread count update", "error", err, "userID", userID)
		}
	}

	return nil
}

// SendReadStatusUpdate sends a real-time update when a notification is read
func (s *NotificationService) SendReadStatusUpdate(ctx context.Context, notification db.Notification) error {
	// Get updated unread count
	unreadCount, err := s.GetUnreadCount(ctx, notification.UserID, notification.OrganizationID, &notification.MarinaID)
	if err != nil {
		s.logger.Zap.Warnw("Failed to get unread count for read status update", "error", err, "userID", notification.UserID)
		unreadCount = 0
	}

	// Create read status update payload
	event := map[string]interface{}{
		"event":        "read_status_update",
		"notification": notification,
		"unreadCount":  unreadCount,
		"timestamp":    time.Now(),
	}

	// Publish to user-specific channel
	userChannel := fmt.Sprintf("notifications:user:%s", notification.UserID.String())
	if err := s.redis.Publish(ctx, userChannel, event); err != nil {
		return fmt.Errorf("failed to publish read status update: %w", err)
	}

	return nil
}

// SendUnreadCountUpdate sends a real-time unread count update
func (s *NotificationService) SendUnreadCountUpdate(ctx context.Context, userID, organizationID uuid.UUID, marinaID *uuid.UUID) error {
	// Get current unread count
	unreadCount, err := s.GetUnreadCount(ctx, userID, organizationID, marinaID)
	if err != nil {
		return fmt.Errorf("failed to get unread count: %w", err)
	}

	// Create unread count update payload
	event := map[string]interface{}{
		"event":       "unread_count_update",
		"unreadCount": unreadCount,
		"timestamp":   time.Now(),
	}

	// Publish to user-specific channel
	userChannel := fmt.Sprintf("notifications:user:%s", userID.String())
	if err := s.redis.Publish(ctx, userChannel, event); err != nil {
		return fmt.Errorf("failed to publish unread count update: %w", err)
	}

	return nil
}

// GetNotificationsByType retrieves notifications of a specific type
func (s *NotificationService) GetNotificationsByType(ctx context.Context, userID, organizationID uuid.UUID, notificationType string, marinaID *uuid.UUID, limit, offset int32) ([]db.Notification, error) {
	params := db.GetNotificationsByTypeParams{
		UserID:         userID,
		OrganizationID: organizationID,
		Type:           notificationType,
		Limit:          limit,
		Offset:         offset,
	}

	// Handle optional marina ID - we need to pass a valid UUID even if it's nullable
	if marinaID != nil {
		params.Column4 = *marinaID
	} else {
		// Pass a nil UUID - the SQL query will handle this with the NULL check
		params.Column4 = uuid.Nil
	}

	notifications, err := s.db.GetNotificationsByType(ctx, params)
	if err != nil {
		s.logger.Zap.Errorw("Failed to get notifications by type", "error", err, "userID", userID, "type", notificationType)
		return nil, fmt.Errorf("failed to get notifications by type: %w", err)
	}

	return notifications, nil
}

// Helper methods for common notification creation patterns

// CreateMessageNotification creates a notification for new messages
func (s *NotificationService) CreateMessageNotification(
	ctx context.Context,
	userID uuid.UUID,
	organizationID uuid.UUID,
	marinaID uuid.UUID,
	messageContent string,
	sender string,
	customerID string,
) error {
	title := "New Message Received"
	content := fmt.Sprintf("You have received a new message from %s", sender)

	// Initialize base data
	data := requests.NotificationData{
		"messagePreview": messageContent,
		"sender":         sender,
		"customerID":     customerID,
	}

	req := requests.CreateNotificationRequest{
		UserID:         userID,
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Type:           "message",
		Title:          title,
		Content:        content,
		Data:           data,
		Priority:       nil, // Use default priority
	}

	_, err := s.CreateNotificationService(ctx, req, true) // Send real-time
	return err
}

// CreateESignNotification creates a notification for e-sign submissions
func (s *NotificationService) CreateESignNotification(ctx context.Context, userID, organizationID, marinaID uuid.UUID, documentURL, submissionID, email string) error {
	title := "New E-Sign Submission Update"
	content := fmt.Sprintf("You have received a new e-sign submission update from %s", email)

	data := requests.NotificationData{
		"document_url":  documentURL,
		"submission_id": submissionID,
		"email":         email,
	}

	req := requests.CreateNotificationRequest{
		UserID:         userID,
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Type:           "esign",
		Title:          title,
		Content:        content,
		Data:           data,
		Priority:       nil, // Use default priority
	}

	_, err := s.CreateNotificationService(ctx, req, true) // Send real-time
	return err
}

// CreateESignSmartNotification creates a smart notification for e-sign submissions that respects user preferences
func (s *NotificationService) CreateESignSmartNotification(ctx context.Context, userID, organizationID, marinaID uuid.UUID, documentURL, submissionID, email string) error {
	title := "New E-Sign Submission Update"
	content := fmt.Sprintf("You have received a new e-sign submission update from %s", email)

	data := map[string]interface{}{
		"document_url":  documentURL,
		"submission_id": submissionID,
		"email":         email,
	}

	// Create smart notification request
	smartReq := SmartNotificationRequest{
		UserID:         userID,
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Type:           "esign",
		Title:          title,
		Content:        content,
		Data:           data,
		Priority:       nil, // Use default priority
	}

	// Send smart notification (respects user preferences)
	result, err := s.SendSmartNotification(ctx, smartReq)
	if err != nil {
		s.logger.Zap.Errorw("Failed to send smart e-sign notification",
			"userID", userID,
			"submissionID", submissionID,
			"email", email,
			"error", err)
		return err
	}

	// Log the result
	s.logger.Zap.Infow("Smart e-sign notification sent",
		"userID", userID,
		"submissionID", submissionID,
		"email", email,
		"systemDelivered", result.SystemDelivered,
		"emailDelivered", result.EmailDelivered,
		"errors", result.Errors)

	return nil
}

// CreateInviteNotification creates a notification for invitations
func (s *NotificationService) CreateInviteNotification(ctx context.Context, userID, organizationID, marinaID uuid.UUID, inviterName string) error {
	title := "Invitation Received"
	content := fmt.Sprintf("You have been invited by %s", inviterName)

	data := requests.NotificationData{
		"inviter": inviterName,
	}

	req := requests.CreateNotificationRequest{
		UserID:         userID,
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Type:           "invite",
		Title:          title,
		Content:        content,
		Data:           data,
		Priority:       utils.Pointer("high"), // Invites are high priority
	}

	_, err := s.CreateNotificationService(ctx, req, true) // Send real-time
	return err
}

// CreateSystemNotification creates a system-level notification
func (s *NotificationService) CreateSystemNotification(ctx context.Context, userID, organizationID, marinaID uuid.UUID, title, content string, priority string) error {
	req := requests.CreateNotificationRequest{
		UserID:         userID,
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Type:           "system",
		Title:          title,
		Content:        content,
		Priority:       &priority,
	}

	_, err := s.CreateNotificationService(ctx, req, true) // Send real-time
	return err
}

// SmartNotificationRequest represents a request for smart notification delivery
type SmartNotificationRequest struct {
	UserID         uuid.UUID              `json:"userId"`
	OrganizationID uuid.UUID              `json:"organizationId"`
	MarinaID       uuid.UUID              `json:"marinaId"`
	Type           string                 `json:"type"` // notification type (message, invite, system, alert, esign)
	Title          string                 `json:"title"`
	Content        string                 `json:"content"`
	Data           map[string]interface{} `json:"data,omitempty"`
	Priority       *string                `json:"priority,omitempty"`
	// Additional fields for multi-channel delivery
	EmailData *EmailNotificationData `json:"emailData,omitempty"`
}

// EmailNotificationData contains email-specific information
type EmailNotificationData struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	ReplyTo string   `json:"replyTo,omitempty"`
}

// SmartNotificationResult represents the result of smart notification delivery
type SmartNotificationResult struct {
	UserID          uuid.UUID  `json:"userId"`
	NotificationID  *uuid.UUID `json:"notificationId,omitempty"`
	SystemDelivered bool       `json:"systemDelivered"`
	EmailDelivered  bool       `json:"emailDelivered"`
	Errors          []string   `json:"errors,omitempty"`
}

// SendSmartNotification sends notifications based on user preferences and available delivery methods
func (s *NotificationService) SendSmartNotification(ctx context.Context, req SmartNotificationRequest) (*SmartNotificationResult, error) {
	result := &SmartNotificationResult{
		UserID: req.UserID,
	}

	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled: %w", ctx.Err())
	}

	// Get user's notification preferences for this type
	preference, err := s.db.GetNotificationPreference(ctx, db.GetNotificationPreferenceParams{
		UserID:           req.UserID,
		NotificationType: req.Type,
	})
	if err != nil {
		// Check if it's a context timeout error
		if ctx.Err() == context.DeadlineExceeded {
			s.logger.Zap.Warnw("Notification preference lookup timed out, notifications disabled by default",
				"userID", req.UserID,
				"type", req.Type,
				"error", err)
			return result, nil
		}

		// If no preference found (no rows), respect user's choice to not receive notifications
		// This implements an opt-in approach where users must explicitly enable notifications
		s.logger.Zap.Infow("DEBUG: No notification preference found, notifications disabled by default",
			"userID", req.UserID,
			"type", req.Type,
			"error", err)
		return result, nil
	}

	// Debug: Log the preference details
	s.logger.Zap.Infow("DEBUG: Notification preference found",
		"userID", req.UserID,
		"type", req.Type,
		"preference_id", preference.ID,
		"enabled", preference.Enabled,
		"delivery_method", preference.DeliveryMethod)

	// Check if notifications are enabled for this user and type
	if preference.Enabled == nil || !*preference.Enabled {
		s.logger.Zap.Infow("DEBUG: Notifications disabled for user",
			"userID", req.UserID,
			"type", req.Type,
			"enabled", preference.Enabled)
		return result, nil
	}

	// Determine delivery method
	deliveryMethod := "system" // default
	if preference.DeliveryMethod != nil {
		deliveryMethod = *preference.DeliveryMethod
	}

	// Send notifications based on delivery method
	switch deliveryMethod {
	case "system":
		return s.sendSystemNotification(ctx, req)
	case "email":
		return s.sendEmailNotification(ctx, req)
	case "all":
		return s.sendMultiChannelNotification(ctx, req)
	default:
		// Fallback to system
		return s.sendSystemNotification(ctx, req)
	}
}

// sendDefaultNotification sends a default system notification when no preferences are set
func (s *NotificationService) sendDefaultNotification(ctx context.Context, req SmartNotificationRequest) (*SmartNotificationResult, error) {
	result := &SmartNotificationResult{
		UserID: req.UserID,
	}

	s.logger.Zap.Debugw("Creating default system notification",
		"userID", req.UserID,
		"type", req.Type)

	// Create and send system notification
	notificationReq := requests.CreateNotificationRequest{
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		MarinaID:       req.MarinaID,
		Type:           req.Type,
		Title:          req.Title,
		Content:        req.Content,
		Data:           req.Data,
		Priority:       req.Priority,
	}

	notification, err := s.CreateNotificationService(ctx, notificationReq, true)
	if err != nil {
		s.logger.Zap.Warnw("Failed to create default system notification",
			"userID", req.UserID,
			"type", req.Type,
			"error", err)

		// Add error to result but don't fail completely
		result.Errors = append(result.Errors, fmt.Sprintf("System notification failed: %v", err))
		// Don't return here - continue to try email if EmailData is provided
	} else {
		s.logger.Zap.Debugw("Default system notification created successfully",
			"userID", req.UserID,
			"type", req.Type,
			"notificationID", notification.ID)

		result.NotificationID = &notification.ID
		result.SystemDelivered = true
	}

	// If EmailData is provided, also send an email notification
	if req.EmailData != nil {
		s.logger.Zap.Debugw("Sending default email notification",
			"userID", req.UserID,
			"type", req.Type)

		// Get user information to construct email data
		user, err := s.db.GetUserByID(ctx, req.UserID)
		if err != nil {
			s.logger.Zap.Warnw("Failed to get user information for default email notification",
				"userID", req.UserID,
				"type", req.Type,
				"error", err)
			result.Errors = append(result.Errors, fmt.Sprintf("Failed to get user information: %v", err))

			// Even if user lookup fails, try to send email using provided EmailData
			if req.EmailData != nil && len(req.EmailData.To) > 0 {
				s.logger.Zap.Debugw("Attempting to send email using provided EmailData",
					"userID", req.UserID,
					"type", req.Type,
					"to_emails", req.EmailData.To)

				// Send email using SendGrid with fallback data
				emailData := sendgrid.NotificationTemplateData{
					Recipient:    "User", // Fallback recipient name
					Type:         req.Type,
					CustomerName: "DockMaster", // Default fallback
					HomeURL:      s.Config.App.FrontendBaseURL,
				}

				subject := req.Title
				if req.EmailData.Subject != "" {
					subject = req.EmailData.Subject
				}

				taskID, resultChan, err := s.sendgrid.SendNotificationEmail(req.EmailData.To, subject, emailData)
				if err != nil {
					s.logger.Zap.Warnw("Failed to send default email notification with fallback data",
						"userID", req.UserID,
						"type", req.Type,
						"error", err)
					result.Errors = append(result.Errors, fmt.Sprintf("Email sending failed: %v", err))
				} else {
					// Monitor email delivery status
					go func() {
						for status := range resultChan {
							if status.Error != "" {
								s.logger.Zap.Errorw("Default email notification delivery failed",
									"userID", req.UserID,
									"taskID", taskID,
									"error", status.Error)
							} else {
								s.logger.Zap.Infow("Default email notification delivered successfully",
									"userID", req.UserID,
									"taskID", taskID,
									"status", status.Status)
							}
						}
					}()

					result.EmailDelivered = true
				}
			}
		} else {
			// Get organization information for customer name
			organization, err := s.db.GetOrganizationByID(ctx, req.OrganizationID)
			if err != nil {
				s.logger.Zap.Warnw("Failed to get organization information for default email notification",
					"userID", req.UserID,
					"organizationID", req.OrganizationID,
					"type", req.Type,
					"error", err)
				// Continue without organization info
			}

			// Send email using SendGrid
			emailData := sendgrid.NotificationTemplateData{
				Recipient: fmt.Sprintf("%s %s", user.FirstName, user.LastName),
				Type:      req.Type,
				CustomerName: func() string {
					if err == nil && organization.Name != "" {
						return organization.Name
					}
					return "DockMaster" // Default fallback
				}(),
				HomeURL: s.Config.App.FrontendBaseURL,
			}

			// Use provided email data if available, otherwise use user's email
			toEmails := []string{user.Email}
			if req.EmailData != nil && len(req.EmailData.To) > 0 {
				toEmails = req.EmailData.To
			}

			subject := req.Title
			if req.EmailData != nil && req.EmailData.Subject != "" {
				subject = req.EmailData.Subject
			}

			taskID, resultChan, err := s.sendgrid.SendNotificationEmail(toEmails, subject, emailData)
			if err != nil {
				s.logger.Zap.Warnw("Failed to send default email notification",
					"userID", req.UserID,
					"type", req.Type,
					"error", err)
				result.Errors = append(result.Errors, fmt.Sprintf("Email sending failed: %v", err))
			} else {
				// Monitor email delivery status
				go func() {
					for status := range resultChan {
						if status.Error != "" {
							s.logger.Zap.Errorw("Default email notification delivery failed",
								"userID", req.UserID,
								"taskID", taskID,
								"error", status.Error)
						} else {
							s.logger.Zap.Infow("Default email notification delivered successfully",
								"userID", req.UserID,
								"taskID", taskID,
								"status", status.Status)
						}
					}
				}()

				result.EmailDelivered = true
			}
		}
	}

	return result, nil
}

// sendSystemNotification sends only a system notification
func (s *NotificationService) sendSystemNotification(ctx context.Context, req SmartNotificationRequest) (*SmartNotificationResult, error) {
	result := &SmartNotificationResult{
		UserID: req.UserID,
	}

	s.logger.Zap.Debugw("Creating system notification",
		"userID", req.UserID,
		"type", req.Type)

	// Create and send system notification
	notificationReq := requests.CreateNotificationRequest{
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		MarinaID:       req.MarinaID,
		Type:           req.Type,
		Title:          req.Title,
		Content:        req.Content,
		Data:           req.Data,
		Priority:       req.Priority,
	}

	notification, err := s.CreateNotificationService(ctx, notificationReq, true)
	if err != nil {
		s.logger.Zap.Warnw("Failed to create system notification",
			"userID", req.UserID,
			"type", req.Type,
			"error", err)

		// Add error to result but don't fail completely
		result.Errors = append(result.Errors, fmt.Sprintf("System notification failed: %v", err))
		return result, nil // Return result with error instead of failing
	}

	s.logger.Zap.Debugw("System notification created successfully",
		"userID", req.UserID,
		"type", req.Type,
		"notificationID", notification.ID)

	result.NotificationID = &notification.ID
	result.SystemDelivered = true
	return result, nil
}

// sendEmailNotification sends only an email notification
func (s *NotificationService) sendEmailNotification(ctx context.Context, req SmartNotificationRequest) (*SmartNotificationResult, error) {
	result := &SmartNotificationResult{
		UserID: req.UserID,
	}

	// Get user information to construct email data
	user, err := s.db.GetUserByID(ctx, req.UserID)
	if err != nil {
		s.logger.Zap.Warnw("Failed to get user information for email notification",
			"userID", req.UserID,
			"type", req.Type,
			"error", err)
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get user information: %v", err))
		return result, nil
	}

	// Get organization information for customer name
	organization, err := s.db.GetOrganizationByID(ctx, req.OrganizationID)
	if err != nil {
		s.logger.Zap.Warnw("Failed to get organization information for email notification",
			"userID", req.UserID,
			"organizationID", req.OrganizationID,
			"type", req.Type,
			"error", err)
		// Continue without organization info - use default fallback
	}

	// Send email using SendGrid
	emailData := sendgrid.NotificationTemplateData{
		Recipient: fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		Type:      req.Type,
		CustomerName: func() string {
			if err == nil && organization.Name != "" {
				return organization.Name
			}
			return "DockMaster" // Default fallback
		}(),
		HomeURL: s.Config.App.FrontendBaseURL,
	}

	// Use provided email data if available, otherwise use user's email
	toEmails := []string{user.Email}
	if req.EmailData != nil && len(req.EmailData.To) > 0 {
		toEmails = req.EmailData.To
	}

	subject := req.Title
	if req.EmailData != nil && req.EmailData.Subject != "" {
		subject = req.EmailData.Subject
	}

	taskID, resultChan, err := s.sendgrid.SendNotificationEmail(toEmails, subject, emailData)
	if err != nil {
		s.logger.Zap.Warnw("Failed to send email notification",
			"userID", req.UserID,
			"type", req.Type,
			"error", err)
		result.Errors = append(result.Errors, fmt.Sprintf("Email sending failed: %v", err))
		return result, nil
	}

	// Monitor email delivery status
	go func() {
		for status := range resultChan {
			if status.Error != "" {
				s.logger.Zap.Errorw("Email notification delivery failed",
					"userID", req.UserID,
					"taskID", taskID,
					"error", status.Error)
			} else {
				s.logger.Zap.Infow("Email notification delivered successfully",
					"userID", req.UserID,
					"taskID", taskID,
					"status", status.Status)
			}
		}
	}()

	result.EmailDelivered = true
	return result, nil
}

// sendMultiChannelNotification sends notifications via all available channels
func (s *NotificationService) sendMultiChannelNotification(ctx context.Context, req SmartNotificationRequest) (*SmartNotificationResult, error) {
	result := &SmartNotificationResult{
		UserID: req.UserID,
	}

	// Create push notification
	notificationReq := requests.CreateNotificationRequest{
		UserID:         req.UserID,
		OrganizationID: req.OrganizationID,
		MarinaID:       req.MarinaID,
		Type:           req.Type,
		Title:          req.Title,
		Content:        req.Content,
		Data:           req.Data,
		Priority:       req.Priority,
	}

	notification, err := s.CreateNotificationService(ctx, notificationReq, true)
	if err != nil {
		s.logger.Zap.Warnw("Failed to create multi-channel notification",
			"userID", req.UserID,
			"type", req.Type,
			"error", err)

		// Add error to result but don't fail completely
		result.Errors = append(result.Errors, fmt.Sprintf("Multi-channel notification failed: %v", err))
		return result, nil // Return result with error instead of failing
	}

	result.NotificationID = &notification.ID
	result.SystemDelivered = true

	// Send email notification as well
	// Get user information to construct email data
	user, err := s.db.GetUserByID(ctx, req.UserID)
	if err != nil {
		s.logger.Zap.Warnw("Failed to get user information for multi-channel email notification",
			"userID", req.UserID,
			"type", req.Type,
			"error", err)
		result.Errors = append(result.Errors, fmt.Sprintf("Failed to get user information: %v", err))
		return result, nil
	}

	// Get organization information for customer name
	organization, err := s.db.GetOrganizationByID(ctx, req.OrganizationID)
	if err != nil {
		s.logger.Zap.Warnw("Failed to get organization information for multi-channel email notification",
			"userID", req.UserID,
			"organizationID", req.OrganizationID,
			"type", req.Type,
			"error", err)
		// Continue without organization info - use empty string for customer name
	}

	// Send email using SendGrid
	emailData := sendgrid.NotificationTemplateData{
		Recipient: fmt.Sprintf("%s %s", user.FirstName, user.LastName),
		Type:      req.Type,
		CustomerName: func() string {
			if err == nil && organization.Name != "" {
				return organization.Name
			}
			return "DockMaster" // Default fallback
		}(),
		HomeURL: s.Config.App.FrontendBaseURL,
	}

	// Use provided email data if available, otherwise use user's email
	toEmails := []string{user.Email}
	if req.EmailData != nil && len(req.EmailData.To) > 0 {
		toEmails = req.EmailData.To
	}

	subject := req.Title
	if req.EmailData != nil && req.EmailData.Subject != "" {
		subject = req.EmailData.Subject
	}

	taskID, resultChan, err := s.sendgrid.SendNotificationEmail(toEmails, subject, emailData)
	if err != nil {
		s.logger.Zap.Warnw("Failed to send multi-channel email notification",
			"userID", req.UserID,
			"type", req.Type,
			"error", err)
		result.Errors = append(result.Errors, fmt.Sprintf("Email sending failed: %v", err))
		return result, nil
	}

	// Monitor email delivery status
	go func() {
		for status := range resultChan {
			if status.Error != "" {
				s.logger.Zap.Errorw("Multi-channel email notification delivery failed",
					"userID", req.UserID,
					"taskID", taskID,
					"error", status.Error)
			} else {
				s.logger.Zap.Infow("Multi-channel email notification delivered successfully",
					"userID", req.UserID,
					"taskID", taskID,
					"status", status.Status)
			}
		}
	}()

	result.EmailDelivered = true
	return result, nil
}

// SendBulkSmartNotifications sends notifications to multiple users based on their preferences
func (s *NotificationService) SendBulkSmartNotifications(ctx context.Context, requests []SmartNotificationRequest) ([]*SmartNotificationResult, error) {
	var results []*SmartNotificationResult

	// Check if context is already cancelled
	if ctx.Err() != nil {
		return nil, fmt.Errorf("context cancelled before starting bulk notifications: %w", ctx.Err())
	}

	s.logger.Zap.Infow("Starting bulk smart notifications",
		"total_requests", len(requests))

	successCount := 0
	errorCount := 0
	startTime := time.Now()

	// Process notifications in batches to avoid overwhelming the database
	// Increased batch size for better efficiency with larger user counts
	const batchSize = 15
	totalBatches := (len(requests) + batchSize - 1) / batchSize

	for batchIndex := 0; batchIndex < totalBatches; batchIndex++ {
		// Check if we're approaching timeout (warn when 80% of time is used)
		elapsed := time.Since(startTime)
		if deadline, ok := ctx.Deadline(); ok {
			remaining := time.Until(deadline)
			progressPercentage := float64(len(results)) / float64(len(requests)) * 100

			// More aggressive timeout warnings
			if remaining < elapsed*4 { // Less than 20% time remaining
				s.logger.Zap.Warnw("Approaching timeout, processing remaining notifications quickly",
					"elapsed", elapsed,
					"remaining", remaining,
					"processed_count", len(results),
					"total_count", len(requests),
					"progress_percentage", fmt.Sprintf("%.1f%%", progressPercentage))
			} else if remaining < elapsed*2 { // Less than 50% time remaining
				s.logger.Zap.Infow("Moderate time pressure, monitoring progress",
					"elapsed", elapsed,
					"remaining", remaining,
					"processed_count", len(results),
					"total_count", len(requests),
					"progress_percentage", fmt.Sprintf("%.1f%%", progressPercentage))
			}
		}

		startIdx := batchIndex * batchSize
		endIdx := startIdx + batchSize
		if endIdx > len(requests) {
			endIdx = len(requests)
		}

		batchRequests := requests[startIdx:endIdx]
		s.logger.Zap.Infow("Processing notification batch",
			"batch_index", batchIndex+1,
			"total_batches", totalBatches,
			"batch_start", startIdx+1,
			"batch_end", endIdx,
			"batch_size", len(batchRequests),
			"elapsed_time", elapsed,
			"progress_percentage", fmt.Sprintf("%.1f%%", float64(len(results))/float64(len(requests))*100))

		// Process each request in the current batch
		for i, req := range batchRequests {
			// Check context before each notification
			if ctx.Err() != nil {
				s.logger.Zap.Warnw("Context cancelled during bulk notifications, stopping",
					"processed_count", startIdx+i,
					"total_count", len(requests),
					"elapsed_time", time.Since(startTime),
					"error", ctx.Err())
				goto endProcessing
			}

			s.logger.Zap.Debugw("Processing notification request",
				"batch_index", batchIndex+1,
				"request_in_batch", i+1,
				"batch_size", len(batchRequests),
				"overall_progress", fmt.Sprintf("%d/%d", startIdx+i+1, len(requests)),
				"userID", req.UserID,
				"type", req.Type)

			result, err := s.SendSmartNotification(ctx, req)
			if err != nil {
				errorCount++
				s.logger.Zap.Warnw("Failed to send smart notification",
					"batch_index", batchIndex+1,
					"request_in_batch", i+1,
					"overall_progress", fmt.Sprintf("%d/%d", startIdx+i+1, len(requests)),
					"userID", req.UserID,
					"type", req.Type,
					"error", err)

				// Create a failed result instead of skipping the user
				failedResult := &SmartNotificationResult{
					UserID: req.UserID,
					Errors: []string{fmt.Sprintf("Notification failed: %v", err)},
				}
				results = append(results, failedResult)
			} else {
				successCount++
				s.logger.Zap.Debugw("Successfully processed notification request",
					"batch_index", batchIndex+1,
					"request_in_batch", i+1,
					"overall_progress", fmt.Sprintf("%d/%d", startIdx+i+1, len(requests)),
					"userID", req.UserID,
					"type", req.Type,
					"system_delivered", result.SystemDelivered,
					"email_delivered", result.EmailDelivered)
				results = append(results, result)
			}
		}

		// Small delay between batches to avoid overwhelming the database
		// Reduced delay for better performance with larger user counts
		// Skip delay if we're approaching timeout
		if batchIndex < totalBatches-1 {
			if deadline, ok := ctx.Deadline(); ok {
				remaining := time.Until(deadline)
				if remaining < 30*time.Second {
					s.logger.Zap.Warnw("Skipping batch delay due to approaching timeout",
						"remaining_time", remaining,
						"batch_index", batchIndex+1)
				} else {
					time.Sleep(25 * time.Millisecond)
				}
			} else {
				time.Sleep(25 * time.Millisecond)
			}
		}
	}

endProcessing:
	totalTime := time.Since(startTime)
	s.logger.Zap.Infow("Bulk notifications completed",
		"total_requests", len(requests),
		"processed_results", len(results),
		"success_count", successCount,
		"error_count", errorCount,
		"total_batches", totalBatches,
		"total_time", totalTime,
		"average_time_per_notification", totalTime/time.Duration(len(results)))

	return results, nil
}

// CreateSmartMessageNotification creates a message notification based on user preferences
// This is a convenience method specifically for message notifications
func (s *NotificationService) CreateSmartMessageNotification(
	ctx context.Context,
	userID, organizationID, marinaID uuid.UUID,
	messageContent string,
	sender string,
	customerID string,
	emailData *EmailNotificationData,
) (*SmartNotificationResult, error) {
	title := "New Message Received"
	content := fmt.Sprintf("You have received a new message from %s", sender)

	// Initialize base data
	data := map[string]interface{}{
		"messagePreview": messageContent,
		"sender":         sender,
		"customerID":     customerID,
	}

	req := SmartNotificationRequest{
		UserID:         userID,
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Type:           "message",
		Title:          title,
		Content:        content,
		Data:           data,
		Priority:       nil, // Use default priority
		EmailData:      emailData,
	}

	return s.SendSmartNotification(ctx, req)
}

// CreateBulkMessageNotifications creates message notifications for multiple users based on their preferences
func (s *NotificationService) CreateBulkMessageNotifications(
	ctx context.Context,
	users []db.GetUsersByMarinaRow,
	organizationID, marinaID uuid.UUID,
	messageContent string,
	sender string,
	customerID string,
	emailData *EmailNotificationData,
) ([]*SmartNotificationResult, error) {
	var requests []SmartNotificationRequest

	for _, user := range users {
		// Only notify active users
		if user.IsActive != nil && *user.IsActive {
			title := "New Message Received"
			content := fmt.Sprintf("You have received a new message from %s", sender)

			data := map[string]interface{}{
				"messagePreview": messageContent,
				"sender":         sender,
				"customerID":     customerID,
			}

			req := SmartNotificationRequest{
				UserID:         user.ID,
				OrganizationID: organizationID,
				MarinaID:       marinaID,
				Type:           "message",
				Title:          title,
				Content:        content,
				Data:           data,
				Priority:       nil,
				EmailData:      emailData,
			}

			requests = append(requests, req)
		}
	}

	return s.SendBulkSmartNotifications(ctx, requests)
}

// CreateBulkMessageNotificationsForCustomers creates message notifications for customer users based on their preferences
func (s *NotificationService) CreateBulkMessageNotificationsForCustomers(
	ctx context.Context,
	users []db.User,
	organizationID, marinaID uuid.UUID,
	messageContent string,
	sender string,
	customerID string,
	messageType string,
) ([]*SmartNotificationResult, error) {
	var requests []SmartNotificationRequest

	for _, user := range users {
		// Only notify active users
		if user.IsActive != nil && *user.IsActive {
			title := "New Message Received"
			content := fmt.Sprintf("You have received a new message from %s", sender)

			data := map[string]interface{}{
				"messagePreview": messageContent,
				"sender":         sender,
				"customerID":     customerID,
			}

			var req SmartNotificationRequest

			if messageType == "email" {
				emailData := &EmailNotificationData{
					To:      []string{user.Email},
					Subject: "Message from " + sender,
				}

				req = SmartNotificationRequest{
					UserID:         user.ID,
					OrganizationID: organizationID,
					MarinaID:       marinaID,
					Type:           "message",
					Title:          title,
					Content:        content,
					Data:           data,
					Priority:       nil,
					EmailData:      emailData,
				}
			} else {
				req = SmartNotificationRequest{
					UserID:         user.ID,
					OrganizationID: organizationID,
					MarinaID:       marinaID,
					Type:           "message",
					Title:          title,
					Content:        content,
					Data:           data,
					Priority:       nil,
					EmailData:      nil,
				}
			}

			requests = append(requests, req)
		}
	}

	return s.SendBulkSmartNotifications(ctx, requests)
}

// CreateBulkDocumentNotifications creates document notifications for multiple marina users based on their preferences
func (s *NotificationService) CreateBulkDocumentNotifications(
	ctx context.Context,
	users []db.GetUsersByMarinaRow,
	organizationID, marinaID uuid.UUID,
	fileName string,
	customerID string,
	emailData *EmailNotificationData,
) ([]*SmartNotificationResult, error) {
	var requests []SmartNotificationRequest

	s.logger.Zap.Infow("Creating bulk document notifications",
		"total_users", len(users),
		"fileName", fileName,
		"customerID", customerID,
		"marinaID", marinaID)

	activeUserCount := 0
	inactiveUserCount := 0

	for _, user := range users {
		// Only notify active users
		if user.IsActive != nil && *user.IsActive {
			activeUserCount++
			title := "New Document Uploaded"
			content := fmt.Sprintf("A new document '%s' has been uploaded by customer", fileName)

			data := map[string]interface{}{
				"fileName":   fileName,
				"customerID": customerID,
			}

			req := SmartNotificationRequest{
				UserID:         user.ID,
				OrganizationID: organizationID,
				MarinaID:       marinaID,
				Type:           "document",
				Title:          title,
				Content:        content,
				Data:           data,
				Priority:       nil,
				EmailData:      emailData,
			}

			requests = append(requests, req)
			s.logger.Zap.Debugw("Added document notification request",
				"userID", user.ID,
				"fileName", fileName)
		} else {
			inactiveUserCount++
			s.logger.Zap.Debugw("Skipping inactive user for document notification",
				"userID", user.ID,
				"isActive", user.IsActive)
		}
	}

	s.logger.Zap.Infow("Document notification requests prepared",
		"total_users", len(users),
		"active_users", activeUserCount,
		"inactive_users", inactiveUserCount,
		"notification_requests", len(requests),
		"fileName", fileName)

	if len(requests) == 0 {
		s.logger.Zap.Warnw("No active users found for document notifications",
			"marinaID", marinaID,
			"fileName", fileName)
		return []*SmartNotificationResult{}, nil
	}

	s.logger.Zap.Infow("Sending bulk document notifications",
		"active_users", len(requests),
		"fileName", fileName)

	results, err := s.SendBulkSmartNotifications(ctx, requests)
	if err != nil {
		s.logger.Zap.Errorw("Failed to send bulk document notifications",
			"error", err,
			"requested_count", len(requests),
			"fileName", fileName)
		return nil, err
	}

	s.logger.Zap.Infow("Bulk document notifications completed",
		"requested_count", len(requests),
		"results_count", len(results),
		"fileName", fileName)

	return results, nil
}
