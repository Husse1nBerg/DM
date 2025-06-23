package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/google/uuid"
)

// NotificationResponse represents a notification in API responses
// @Description Notification data including type, title, content, and status information
type NotificationResponse struct {
	ID             uuid.UUID                 `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID         uuid.UUID                 `json:"userId" example:"550e8400-e29b-41d4-a716-446655440001"`
	OrganizationID uuid.UUID                 `json:"organizationId" example:"550e8400-e29b-41d4-a716-446655440002"`
	MarinaID       uuid.UUID                 `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440003"`
	Type           string                    `json:"type" example:"message"`
	Title          string                    `json:"title" example:"New Message Received"`
	Content        string                    `json:"content" example:"You have received a new message from the marina."`
	Data           requests.NotificationData `json:"data,omitempty"`
	Read           bool                      `json:"read" example:"false"`
	ReadAt         *time.Time                `json:"readAt,omitempty"`
	Priority       string                    `json:"priority" example:"normal"`
	CreatedAt      time.Time                 `json:"createdAt"`
	UpdatedAt      time.Time                 `json:"updatedAt"`
}

// NotificationResponseWrapper wraps a notification response
// @Description Wrapper for a single notification response
type NotificationResponseWrapper struct {
	Data NotificationResponse `json:"data"`
}

// NotificationListResponse represents a paginated list of notifications
// @Description Paginated list of notifications
type NotificationListResponse struct {
	Data        []NotificationResponse `json:"data"`
	Total       int64                  `json:"total" example:"42"`
	PerPage     int32                  `json:"perPage" example:"10"`
	CurrentPage int32                  `json:"currentPage" example:"1"`
	LastPage    int32                  `json:"lastPage" example:"5"`
}

// UnreadCountResponse represents the count of unread notifications
// @Description Unread notification count
type UnreadCountResponse struct {
	Count int64 `json:"count" example:"5"`
}

// RealTimeNotificationResponse represents a real-time notification for SSE
// @Description Real-time notification event for Server-Sent Events
type RealTimeNotificationResponse struct {
	Event        string               `json:"event" example:"notification"` // "notification" or "read_status_update"
	Notification NotificationResponse `json:"notification"`
	UnreadCount  int64                `json:"unreadCount" example:"4"`
	Timestamp    time.Time            `json:"timestamp"`
}

// Helper functions to convert from database models

// NotificationDBToResponse creates a NotificationResponse from database model
func NotificationDBToResponse(notification db.Notification) NotificationResponse {
	response := NotificationResponse{
		ID:             notification.ID,
		UserID:         notification.UserID,
		OrganizationID: notification.OrganizationID,
		MarinaID:       notification.MarinaID,
		Type:           notification.Type,
		Title:          notification.Title,
		Content:        notification.Content,
		Read:           false,    // Default to false
		Priority:       "normal", // Default priority
		CreatedAt:      notification.CreatedAt.Time,
		UpdatedAt:      notification.UpdatedAt.Time,
	}

	// Handle nullable Read field
	if notification.Read != nil {
		response.Read = *notification.Read
	}

	// Handle nullable Priority field
	if notification.Priority != nil {
		response.Priority = *notification.Priority
	}

	// Handle nullable ReadAt field
	if notification.ReadAt.Valid {
		response.ReadAt = &notification.ReadAt.Time
	}

	// Handle JSONB Data field
	if notification.Data != nil {
		data := make(requests.NotificationData)
		if err := data.Scan(notification.Data); err == nil {
			response.Data = data
		}
	}

	return response
}

// NewNotificationResponseSuccess creates a successful single notification response
func NewNotificationResponseSuccess(notification db.Notification) BaseResponse {
	return NewSuccessResponse(NotificationDBToResponse(notification))
}

// NewNotificationsPaginatedResponse creates a paginated response of notifications
func NewNotificationsPaginatedResponse(notifications []db.Notification, total int64, perPage, currentPage int32) BaseResponse {
	notificationResponses := make([]NotificationResponse, len(notifications))
	for i, notification := range notifications {
		notificationResponses[i] = NotificationDBToResponse(notification)
	}

	return NewPaginatedResponse(notificationResponses, total, perPage, currentPage)
}

// NewUnreadCountResponse creates an unread count response
func NewUnreadCountResponse(count int64) BaseResponse {
	return NewSuccessResponse(UnreadCountResponse{Count: count})
}

// NewRealTimeNotificationResponse creates a RealTimeNotificationResponse
func NewRealTimeNotificationResponse(notification db.Notification, unreadCount int64, event string) *RealTimeNotificationResponse {
	return &RealTimeNotificationResponse{
		Event:        event,
		Notification: NotificationDBToResponse(notification),
		UnreadCount:  unreadCount,
		Timestamp:    time.Now(),
	}
}
