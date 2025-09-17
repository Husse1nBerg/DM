package responses

import (
	"time"

	"github.com/google/uuid"
)

// PaymentSessionResponse represents the response for a payment session creation
type PaymentSessionResponse struct {
	ID             string                   `json:"id"`
	SessionData    string                   `json:"sessionData"`
	ExpiresAt      time.Time                `json:"expiresAt"`
	PaymentMethods []map[string]interface{} `json:"paymentMethods"`
	Reference      string                   `json:"reference"`
}

// PaymentResponse represents a payment record
type PaymentResponse struct {
	ID                     uuid.UUID `json:"id"`
	OrganizationID         uuid.UUID `json:"organization_id"`
	MarinaID               uuid.UUID `json:"marina_id"`
	CustomerID             uuid.UUID `json:"customer_id"`
	Amount                 float64   `json:"amount"`
	Currency               string    `json:"currency"`
	Status                 string    `json:"status"`
	PaymentMethod          string    `json:"payment_method,omitempty"`
	PaymentType            string    `json:"payment_type"`
	Reference              string    `json:"reference"`
	Description            string    `json:"description,omitempty"`
	AdyenPaymentID         string    `json:"adyen_payment_id,omitempty"`
	AdyenMerchantReference string    `json:"adyen_merchant_reference,omitempty"`
	AdyenPspReference      string    `json:"adyen_psp_reference,omitempty"`
	CreatedAt              time.Time `json:"created_at"`
	UpdatedAt              time.Time `json:"updated_at,omitempty"`
}

// PaymentListResponse represents a paginated list of payments
type PaymentListResponse struct {
	Payments []PaymentResponse `json:"payments"`
	Total    int64             `json:"total"`
	Page     int               `json:"page"`
	Limit    int               `json:"limit"`
}

// WebhookResponse represents the response to a webhook notification
type WebhookResponse struct {
	NotificationResponse string `json:"notificationResponse"`
}
