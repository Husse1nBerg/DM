package responses

import (
	"github.com/adyen/adyen-go-api-library/v14/src/checkout"
)

// PaymentSessionResponse represents the response for creating a payment session
type PaymentSessionResponse struct {
	SessionData string `json:"session_data"`
	SessionID   string `json:"session_id"`
	ClientKey   string `json:"client_key"`
}

// PaymentResultResponse represents the response for payment completion
type PaymentResultResponse struct {
	PspReference  string  `json:"psp_reference"`
	ResultCode    string  `json:"result_code"`
	RefusalReason *string `json:"refusal_reason,omitempty"`
}

// WebhookResponse represents the response for webhook processing
type WebhookResponse struct {
	Message string `json:"message"`
	Status  string `json:"status"`
}

// NewPaymentSessionResponse creates a new payment session response
func NewPaymentSessionResponse(session *checkout.CreateCheckoutSessionResponse, clientKey string) *PaymentSessionResponse {
	return &PaymentSessionResponse{
		SessionData: *session.SessionData,
		SessionID:   session.Id,
		ClientKey:   clientKey,
	}
}

// NewPaymentResultResponse creates a new payment result response
func NewPaymentResultResponse(result *checkout.PaymentDetailsResponse) *PaymentResultResponse {
	response := &PaymentResultResponse{
		PspReference: *result.PspReference,
		ResultCode:   *result.ResultCode,
	}

	if result.RefusalReason != nil {
		response.RefusalReason = result.RefusalReason
	}

	return response
}
