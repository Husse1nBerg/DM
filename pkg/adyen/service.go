package adyen

import (
	"context"
	"fmt"

	"github.com/adyen/adyen-go-api-library/v14/src/checkout"
	"github.com/adyen/adyen-go-api-library/v14/src/common"
	"github.com/adyen/adyen-go-api-library/v14/src/hmacvalidator"
	"github.com/adyen/adyen-go-api-library/v14/src/webhook"
	"github.com/google/uuid"
)

type PaymentService struct {
	client *Client
}

func NewPaymentService(client *Client) *PaymentService {
	return &PaymentService{
		client: client,
	}
}

// CreateCheckoutSession creates a new checkout session for payment
func (s *PaymentService) CreateCheckoutSession(ctx context.Context, req CreateCheckoutSessionRequest) (*checkout.CreateCheckoutSessionResponse, error) {
	service := s.client.APIClient.Checkout()

	orderRef := uuid.Must(uuid.NewRandom())

	body := checkout.CreateCheckoutSessionRequest{
		Reference: orderRef.String(),
		Amount: checkout.Amount{
			Value:    req.Amount,
			Currency: req.Currency,
		},
		CountryCode:           common.PtrString(req.CountryCode),
		MerchantAccount:       s.client.MerchantAccount,
		Channel:               common.PtrString("Web"),
		ReturnUrl:             req.ReturnURL,
		ShopperIP:             common.PtrString(req.ShopperIP),
		LineItems:             req.LineItems,
		Metadata:              req.Metadata,              // Custom metadata that Adyen will echo back in webhooks
		AllowedPaymentMethods: req.AllowedPaymentMethods, // e.g., ["scheme", "ach"] for cards and ACH
	}

	apiReq := service.PaymentsApi.SessionsInput().CreateCheckoutSessionRequest(body)
	res, httpRes, err := service.PaymentsApi.Sessions(ctx, apiReq)

	if err != nil {
		return nil, fmt.Errorf("failed to create checkout session: %w", err)
	}

	if httpRes.StatusCode >= 300 {
		return nil, fmt.Errorf("adyen API error: status %d", httpRes.StatusCode)
	}

	return &res, nil
}

// HandlePaymentDetails handles payment completion details (for redirects)
func (s *PaymentService) HandlePaymentDetails(ctx context.Context, req PaymentDetailsRequest) (*checkout.PaymentDetailsResponse, error) {
	service := s.client.APIClient.Checkout()

	apiReq := service.PaymentsApi.PaymentsDetailsInput()
	apiReq = apiReq.PaymentDetailsRequest(checkout.PaymentDetailsRequest{
		PaymentData: common.PtrString(req.PaymentData),
		Details: checkout.PaymentCompletionDetails{
			RedirectResult: common.PtrString(req.RedirectResult),
			Payload:        common.PtrString(req.Payload),
		},
	})

	res, httpRes, err := service.PaymentsApi.PaymentsDetails(ctx, apiReq)

	if err != nil {
		return nil, fmt.Errorf("failed to handle payment details: %w", err)
	}

	if httpRes.StatusCode >= 300 {
		return nil, fmt.Errorf("adyen API error: status %d", httpRes.StatusCode)
	}

	return &res, nil
}

// ValidateWebhook validates the HMAC signature of incoming webhooks
func (s *PaymentService) ValidateWebhook(notification webhook.NotificationRequestItem) bool {
	return hmacvalidator.ValidateHmac(notification, s.client.HMACKey)
}

// ProcessWebhook processes incoming webhook notifications
func (s *PaymentService) ProcessWebhook(body string) (*webhook.Webhook, error) {
	notificationRequest, err := webhook.HandleRequest(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse webhook request: %w", err)
	}

	return notificationRequest, nil
}

// CreateCheckoutSessionRequest represents the request to create a checkout session
type CreateCheckoutSessionRequest struct {
	Amount                int64               `json:"amount"`
	Currency              string              `json:"currency"`
	CountryCode           string              `json:"country_code"`
	ReturnURL             string              `json:"return_url"`
	ShopperIP             string              `json:"shopper_ip"`
	LineItems             []checkout.LineItem `json:"line_items,omitempty"`
	Metadata              *map[string]string  `json:"metadata,omitempty"`                // Custom data that will be echoed back in webhooks
	AllowedPaymentMethods []string            `json:"allowed_payment_methods,omitempty"` // e.g., ["scheme", "ach"] for cards and ACH
}

// PaymentDetailsRequest represents the request to handle payment details
type PaymentDetailsRequest struct {
	PaymentData    string `json:"payment_data"`
	RedirectResult string `json:"redirect_result,omitempty"`
	Payload        string `json:"payload,omitempty"`
}
