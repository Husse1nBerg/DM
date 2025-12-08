package requests

import (
	"database/sql/driver"
	"encoding/json"
	"errors"

	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
)

// NotificationData represents the JSONB data field for notifications
type NotificationData map[string]interface{}

// Value implements the driver.Valuer interface for JSONB storage
func (nd NotificationData) Value() (driver.Value, error) {
	if nd == nil {
		return nil, nil
	}
	return json.Marshal(nd)
}

// Scan implements the sql.Scanner interface for JSONB retrieval
func (nd *NotificationData) Scan(value interface{}) error {
	if value == nil {
		*nd = nil
		return nil
	}

	bytes, ok := value.([]byte)
	if !ok {
		return errors.New("cannot scan non-[]byte value into NotificationData")
	}

	return json.Unmarshal(bytes, nd)
}

// CreateNotificationRequest represents a request to create a notification
type CreateNotificationRequest struct {
	UserID         uuid.UUID        `json:"userId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440000"`
	OrganizationID uuid.UUID        `json:"organizationId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	MarinaID       uuid.UUID        `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440002"`
	Type           string           `json:"type" validate:"required,oneof=message invite system alert esign document payment service" example:"message"`
	Title          string           `json:"title" validate:"required,max=255" example:"New Message Received"`
	Content        string           `json:"content" validate:"required" example:"You have received a new message from the marina."`
	Data           NotificationData `json:"data,omitempty" example:"{}"`
	Priority       *string          `json:"priority,omitempty" validate:"omitempty,oneof=low normal high urgent" example:"normal"`
}

// ListNotificationsRequest represents a request to list notifications with pagination
type ListNotificationsRequest struct {
	PaginationQuery
	MarinaID *uuid.UUID `query:"marinaId,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
	// Filtros adicionales similares a esign_submissions
	Search     string `query:"search,omitempty" example:"payment"` // Búsqueda global en title y content
	ReadFilter *bool  `query:"read,omitempty" example:"false"`     // Filtro por estado leído
	TypeFilter string `query:"type,omitempty" example:"message"`   // Filtro por tipo de notificación
	// Sorting
	SortBy    string `query:"sortBy,omitempty" example:"created_at"` // type, read, created_at, title
	SortOrder string `query:"sortOrder,omitempty" example:"desc"`    // asc, desc
}

// ListUnreadNotificationsRequest represents a request to list unread notifications
type ListUnreadNotificationsRequest struct {
	UserID         uuid.UUID  `json:"user_id" validate:"required"`
	OrganizationID uuid.UUID  `json:"organization_id" validate:"required"`
	MarinaID       *uuid.UUID `json:"marina_id,omitempty"`
	Limit          int32      `json:"limit" validate:"min=1,max=100"`
	Offset         int32      `json:"offset" validate:"min=0"`
}

// GetUnreadCountRequest represents a request to get unread notification count
type GetUnreadCountRequest struct {
	MarinaID *uuid.UUID `query:"marinaId,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
}

// NotificationIDParam represents a URL parameter for notification ID
type NotificationIDParam struct {
	NotificationID uuid.UUID `param:"notificationId" validate:"required"`
}

// MarkNotificationAsReadRequest represents a request to mark a notification as read
type MarkNotificationAsReadRequest struct {
	NotificationID uuid.UUID `json:"notification_id" validate:"required"`
	UserID         uuid.UUID `json:"user_id" validate:"required"`
}

// MarkAllNotificationsAsReadRequest represents a request to mark all notifications as read
type MarkAllNotificationsAsReadRequest struct {
	MarinaID *uuid.UUID `json:"marinaId,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
}

// GetNotificationsByTypeRequest represents a request to get notifications by type
type GetNotificationsByTypeRequest struct {
	PaginationQuery
	Type     string     `query:"type" validate:"required,oneof=message invite system alert esign document payment service" example:"message"`
	MarinaID *uuid.UUID `query:"marinaId,omitempty" example:"550e8400-e29b-41d4-a716-446655440002"`
}

// Validate performs custom validation on the request
func (r *CreateNotificationRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the request
func (r *ListNotificationsRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the request
func (r *GetUnreadCountRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the request
func (r *MarkAllNotificationsAsReadRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the request
func (r *GetNotificationsByTypeRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}
