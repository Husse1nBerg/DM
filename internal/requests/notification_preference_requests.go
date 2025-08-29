package requests

import (
	"github.com/go-playground/validator/v10"
)

// NotificationPreferenceRequest represents a request for notification preferences
type NotificationPreferenceRequest struct {
	NotificationType string `json:"notificationType" validate:"required,oneof=message invite system alert document esign payment" example:"message"`
	Enabled          bool   `json:"enabled" validate:"required" example:"true"`
	DeliveryMethod   string `json:"deliveryMethod" validate:"required,oneof=system email all" example:"system"`
}

// UpdateNotificationPreferencesRequest represents a request to update multiple preferences
type UpdateNotificationPreferencesRequest struct {
	Preferences []NotificationPreferenceRequest `json:"preferences" validate:"required,dive"`
}

// Validate performs custom validation on the request
func (r *NotificationPreferenceRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
