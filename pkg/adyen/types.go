package adyen

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PaymentEvent represents a payment event in the system
type PaymentEvent struct {
	ID        uuid.UUID       `json:"id"`
	PaymentID uuid.UUID       `json:"payment_id"`
	EventType string          `json:"event_type"`
	EventData json.RawMessage `json:"event_data"`
	CreatedAt time.Time       `json:"created_at"`
}

// PaymentStatus represents the possible states of a payment
type PaymentStatus string

const (
	PaymentStatusPending    PaymentStatus = "pending"
	PaymentStatusAuthorized PaymentStatus = "authorized"
	PaymentStatusCaptured   PaymentStatus = "captured"
	PaymentStatusFailed     PaymentStatus = "failed"
	PaymentStatusCancelled  PaymentStatus = "cancelled"
	PaymentStatusRefunded   PaymentStatus = "refunded"
)

// PaymentType represents the type of payment
type PaymentType string

const (
	PaymentTypeOneTime      PaymentType = "one_time"
	PaymentTypeRecurring    PaymentType = "recurring"
	PaymentTypeSubscription PaymentType = "subscription"
)
