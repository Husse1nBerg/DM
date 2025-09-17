package requests

import "github.com/google/uuid"

type CreatePaymentSessionRequest struct {
	OrganizationID uuid.UUID `json:"organization_id" validate:"required"`
	MarinaID       uuid.UUID `json:"marina_id" validate:"required"`
	Amount         float64   `json:"amount" validate:"required,gt=0"`
	Currency       string    `json:"currency" validate:"required,len=3"`
	ReturnURL      string    `json:"return_url" validate:"required,url"`
	CustomerID     uuid.UUID `json:"customer_id" validate:"required"`
	CountryCode    string    `json:"country_code" validate:"required,len=2"`
	ShopperLocale  string    `json:"shopper_locale" validate:"required"`
	PaymentMethods []string  `json:"payment_methods" validate:"required,dive,oneof=visa mc amex discover ach"`
	Description    string    `json:"description" validate:"omitempty"`
	PaymentType    string    `json:"payment_type" validate:"required"`
}

type PaymentWebhookRequest struct {
	Live              bool               `json:"live"`
	NotificationItems []NotificationItem `json:"notificationItems"`
}

type NotificationItem struct {
	NotificationRequestItem NotificationRequestItem `json:"NotificationRequestItem"`
}

type NotificationRequestItem struct {
	Amount struct {
		Value    int64  `json:"value"`
		Currency string `json:"currency"`
	} `json:"amount"`
	EventCode           string            `json:"eventCode"`
	EventDate           string            `json:"eventDate"`
	MerchantAccountCode string            `json:"merchantAccountCode"`
	MerchantReference   string            `json:"merchantReference"`
	OriginalReference   string            `json:"originalReference"`
	PaymentMethod       string            `json:"paymentMethod"`
	PSPReference        string            `json:"pspReference"`
	Reason              string            `json:"reason"`
	Success             bool              `json:"success"`
	AdditionalData      map[string]string `json:"additionalData"`
}
