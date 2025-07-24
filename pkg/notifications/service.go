package notifications

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/dockworks/dm-web-backend/pkg/redis"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// NotificationService provides notification management functionality
type NotificationService struct {
	db     *db.Queries
	redis  *redis.Client
	logger *logger.Logger
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
func NewNotificationService(database *db.Queries, redisClient *redis.Client, logger *logger.Logger) *NotificationService {
	return &NotificationService{
		db:     database,
		redis:  redisClient,
		logger: logger,
	}
}

// CreateNotification creates a new notification and optionally sends it in real-time
func (s *NotificationService) CreateNotification(ctx context.Context, req requests.CreateNotificationRequest, sendRealTime bool) (*db.Notification, error) {
	// Validate the request
	if err := req.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
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
		s.logger.Zap.Errorw("Failed to create notification", "error", err, "userID", req.UserID)
		return nil, fmt.Errorf("failed to create notification: %w", err)
	}

	s.logger.Zap.Infow("Notification created", "notificationID", notification.ID, "userID", req.UserID, "type", req.Type)

	// Send real-time notification if requested
	if sendRealTime {
		if err := s.SendRealTimeNotification(ctx, notification); err != nil {
			s.logger.Zap.Warnw("Failed to send real-time notification", "error", err, "notificationID", notification.ID)
			// Don't fail the entire operation if real-time sending fails
		}
	}

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
func (s *NotificationService) CreateMessageNotification(ctx context.Context, userID, organizationID, marinaID uuid.UUID, messageContent string, sender string, customerID uuid.UUID) error {
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

	_, err := s.CreateNotification(ctx, req, true) // Send real-time
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

	_, err := s.CreateNotification(ctx, req, true) // Send real-time
	return err
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

	_, err := s.CreateNotification(ctx, req, true) // Send real-time
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

	_, err := s.CreateNotification(ctx, req, true) // Send real-time
	return err
}
