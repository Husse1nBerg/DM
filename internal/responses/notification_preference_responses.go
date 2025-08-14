package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/google/uuid"
)

// NotificationPreferenceResponse represents a notification preference in API responses
// @Description Notification preference data including type, enabled status, and delivery method
type NotificationPreferenceResponse struct {
	ID               uuid.UUID `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	UserID           uuid.UUID `json:"userId" example:"550e8400-e29b-41d4-a716-446655440001"`
	NotificationType string    `json:"notificationType" example:"message"`
	Enabled          bool      `json:"enabled" example:"true"`
	DeliveryMethod   string    `json:"deliveryMethod" example:"push"`
	CreatedAt        time.Time `json:"createdAt"`
	UpdatedAt        time.Time `json:"updatedAt"`
}

// NotificationPreferencesResponse represents a list of notification preferences
// @Description List of notification preferences
type NotificationPreferencesResponse struct {
	Data []NotificationPreferenceResponse `json:"data"`
}

// Helper functions to convert from database models

// NotificationPreferenceDBToResponse creates a NotificationPreferenceResponse from database model
func NotificationPreferenceDBToResponse(pref db.NotificationPreference) NotificationPreferenceResponse {
	response := NotificationPreferenceResponse{
		ID:               pref.ID,
		UserID:           pref.UserID,
		NotificationType: pref.NotificationType,
		Enabled:          true,   // Default to true
		DeliveryMethod:   "push", // Default delivery method
		CreatedAt:        pref.CreatedAt.Time,
		UpdatedAt:        pref.UpdatedAt.Time,
	}

	// Handle nullable Enabled field
	if pref.Enabled != nil {
		response.Enabled = *pref.Enabled
	}

	// Handle nullable DeliveryMethod field
	if pref.DeliveryMethod != nil {
		response.DeliveryMethod = *pref.DeliveryMethod
	}

	return response
}

// NewNotificationPreferencesResponse creates a notification preferences response
func NewNotificationPreferencesResponse(preferences []db.NotificationPreference) BaseResponse {
	preferenceResponses := make([]NotificationPreferenceResponse, len(preferences))
	for i, pref := range preferences {
		preferenceResponses[i] = NotificationPreferenceDBToResponse(pref)
	}

	return NewSuccessResponse(NotificationPreferencesResponse{Data: preferenceResponses})
}

func NotificationPreferenceDBToResponseList(preferences []db.NotificationPreference) BaseResponse {
	preferenceResponses := make([]NotificationPreferenceResponse, len(preferences))
	for i, pref := range preferences {
		preferenceResponses[i] = NotificationPreferenceDBToResponse(pref)
	}

	return NewSuccessResponse(preferenceResponses)
}

// NewNotificationPreferenceResponseSuccess creates a successful single preference response
func NewNotificationPreferenceResponseSuccess(preference db.NotificationPreference) BaseResponse {
	return NewSuccessResponse(NotificationPreferenceDBToResponse(preference))
}
