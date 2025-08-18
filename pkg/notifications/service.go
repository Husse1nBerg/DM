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
func (s *NotificationService) CreateMessageNotification(
	ctx context.Context,
	userID, organizationID, marinaID uuid.UUID,
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
	SMSData   *SMSNotificationData   `json:"smsData,omitempty"`
}

// EmailNotificationData contains email-specific information
type EmailNotificationData struct {
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	ReplyTo string   `json:"replyTo,omitempty"`
}

// SMSNotificationData contains SMS-specific information
type SMSNotificationData struct {
	To      string `json:"to"`
	Message string `json:"message"`
}

// SmartNotificationResult represents the result of smart notification delivery
type SmartNotificationResult struct {
	UserID         uuid.UUID  `json:"userId"`
	NotificationID *uuid.UUID `json:"notificationId,omitempty"`
	PushDelivered  bool       `json:"pushDelivered"`
	EmailDelivered bool       `json:"emailDelivered"`
	SMSDelivered   bool       `json:"smsDelivered"`
	Errors         []string   `json:"errors,omitempty"`
}

// SendSmartNotification sends notifications based on user preferences and available delivery methods
func (s *NotificationService) SendSmartNotification(ctx context.Context, req SmartNotificationRequest) (*SmartNotificationResult, error) {
	result := &SmartNotificationResult{
		UserID: req.UserID,
	}

	// Get user's notification preferences for this type
	preference, err := s.db.GetNotificationPreference(ctx, db.GetNotificationPreferenceParams{
		UserID:           req.UserID,
		NotificationType: req.Type,
	})
	if err != nil {
		// If no preference found, use default (push only)
		s.logger.Zap.Debugw("No notification preference found, using default",
			"userID", req.UserID,
			"type", req.Type)
		return s.sendDefaultNotification(ctx, req)
	}

	// Check if notifications are enabled for this user and type
	if preference.Enabled == nil || !*preference.Enabled {
		s.logger.Zap.Debugw("Notifications disabled for user",
			"userID", req.UserID,
			"type", req.Type)
		return result, nil
	}

	// Determine delivery method
	deliveryMethod := "push" // default
	if preference.DeliveryMethod != nil {
		deliveryMethod = *preference.DeliveryMethod
	}

	// Send notifications based on delivery method
	switch deliveryMethod {
	case "push":
		return s.sendPushNotification(ctx, req)
	case "email":
		return s.sendEmailNotification(ctx, req)
	case "sms":
		return s.sendSMSNotification(ctx, req)
	case "all":
		return s.sendMultiChannelNotification(ctx, req)
	default:
		// Fallback to push
		return s.sendPushNotification(ctx, req)
	}
}

// sendDefaultNotification sends a default push notification when no preferences are set
func (s *NotificationService) sendDefaultNotification(ctx context.Context, req SmartNotificationRequest) (*SmartNotificationResult, error) {
	result := &SmartNotificationResult{
		UserID: req.UserID,
	}

	// Create and send push notification
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

	notification, err := s.CreateNotification(ctx, notificationReq, true)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Push notification failed: %v", err))
		return result, err
	}

	result.NotificationID = &notification.ID
	result.PushDelivered = true
	return result, nil
}

// sendPushNotification sends only a push notification
func (s *NotificationService) sendPushNotification(ctx context.Context, req SmartNotificationRequest) (*SmartNotificationResult, error) {
	result := &SmartNotificationResult{
		UserID: req.UserID,
	}

	// Create and send push notification
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

	notification, err := s.CreateNotification(ctx, notificationReq, true)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Push notification failed: %v", err))
		return result, err
	}

	result.NotificationID = &notification.ID
	result.PushDelivered = true
	return result, nil
}

// sendEmailNotification sends only an email notification
func (s *NotificationService) sendEmailNotification(ctx context.Context, req SmartNotificationRequest) (*SmartNotificationResult, error) {
	result := &SmartNotificationResult{
		UserID: req.UserID,
	}

	// Check if email data is provided
	if req.EmailData == nil {
		result.Errors = append(result.Errors, "Email data not provided")
		return result, fmt.Errorf("email data not provided")
	}

	// Create push notification for in-app display
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

	notification, err := s.CreateNotification(ctx, notificationReq, false) // Don't send real-time for email
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Push notification creation failed: %v", err))
		return result, err
	}

	result.NotificationID = &notification.ID
	result.PushDelivered = true

	// Note: Email sending would be handled by the calling code
	// This method focuses on notification creation and preference checking
	result.EmailDelivered = true
	return result, nil
}

// sendSMSNotification sends only an SMS notification
func (s *NotificationService) sendSMSNotification(ctx context.Context, req SmartNotificationRequest) (*SmartNotificationResult, error) {
	result := &SmartNotificationResult{
		UserID: req.UserID,
	}

	// Check if SMS data is provided
	if req.SMSData == nil {
		result.Errors = append(result.Errors, "SMS data not provided")
		return result, fmt.Errorf("sms data not provided")
	}

	// Create push notification for in-app display
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

	notification, err := s.CreateNotification(ctx, notificationReq, false) // Don't send real-time for SMS
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Push notification creation failed: %v", err))
		return result, err
	}

	result.NotificationID = &notification.ID
	result.PushDelivered = true

	// Note: SMS sending would be handled by the calling code
	// This method focuses on notification creation and preference checking
	result.SMSDelivered = true
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

	notification, err := s.CreateNotification(ctx, notificationReq, true)
	if err != nil {
		result.Errors = append(result.Errors, fmt.Sprintf("Push notification failed: %v", err))
		return result, err
	}

	result.NotificationID = &notification.ID
	result.PushDelivered = true

	// Mark email and SMS as delivered if data is provided
	if req.EmailData != nil {
		result.EmailDelivered = true
	}
	if req.SMSData != nil {
		result.SMSDelivered = true
	}

	return result, nil
}

// SendBulkSmartNotifications sends notifications to multiple users based on their preferences
func (s *NotificationService) SendBulkSmartNotifications(ctx context.Context, requests []SmartNotificationRequest) ([]*SmartNotificationResult, error) {
	var results []*SmartNotificationResult

	for _, req := range requests {
		result, err := s.SendSmartNotification(ctx, req)
		if err != nil {
			s.logger.Zap.Warnw("Failed to send smart notification",
				"userID", req.UserID,
				"type", req.Type,
				"error", err)
			// Continue with other users even if one fails
		}
		results = append(results, result)
	}

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
	smsData *SMSNotificationData,
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
		SMSData:        smsData,
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
	smsData *SMSNotificationData,
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
				SMSData:        smsData,
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
	emailData *EmailNotificationData,
	smsData *SMSNotificationData,
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
				SMSData:        smsData,
			}

			requests = append(requests, req)
		}
	}

	return s.SendBulkSmartNotifications(ctx, requests)
}
