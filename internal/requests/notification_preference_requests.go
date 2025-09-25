package requests

import (
	"errors"

	"github.com/go-playground/validator/v10"
)

// NotificationPreferenceRequest represents a request for notification preferences
type NotificationPreferenceRequest struct {
	NotificationType string `json:"notificationType" validate:"required,oneof=message invite system alert document esign payment" example:"message"`
	Enabled          *bool  `json:"enabled" example:"true"`
	DeliveryMethod   string `json:"deliveryMethod" validate:"required,oneof=system email all" example:"system"`
}

// UpdateNotificationPreferencesRequest represents a request to update multiple preferences
type UpdateNotificationPreferencesRequest struct {
	Preferences []NotificationPreferenceRequest `json:"preferences" validate:"required,dive"`
}

// Validate performs validation on the request
func (r *NotificationPreferenceRequest) Validate() error {
	// First validate the struct tags
	validate := validator.New()
	if err := validate.Struct(r); err != nil {
		return err
	}

	// Then manually validate the Enabled field
	if r.Enabled == nil {
		return errors.New("enabled field is required")
	}

	return nil
}
