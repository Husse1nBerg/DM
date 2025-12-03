package responses

import (
	"fmt"
	"time"

	"github.com/adyen/adyen-go-api-library/v14/src/checkout"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// PaymentSessionResponse represents the response for creating a payment session
type PaymentSessionResponse struct {
	SessionData string `json:"sessionData"`
	SessionID   string `json:"sessionId"`
	ClientKey   string `json:"clientKey"`
}

// PaymentResultResponse represents the response for payment completion
type PaymentResultResponse struct {
	PspReference  string  `json:"pspReference"`
	ResultCode    string  `json:"resultCode"`
	RefusalReason *string `json:"refusalReason,omitempty"`
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

// PaymentResponse represents a payment record
type PaymentResponse struct {
	ID                   uuid.UUID  `json:"id"`
	MarinaID             uuid.UUID  `json:"marinaId"`
	OrganizationID       uuid.UUID  `json:"organizationId"`
	EntityType           *string    `json:"entityType,omitempty"`
	EntityID             *string    `json:"entityId,omitempty"`
	Amount               string     `json:"amount"`
	Currency             string     `json:"currency"`
	PaymentMethod        *string    `json:"paymentMethod,omitempty"`
	ReferenceNumber      string     `json:"referenceNumber"`
	Status               string     `json:"status"`
	AuthorizationStatus  *string    `json:"authorizationStatus,omitempty"`
	BatchStatus          *string    `json:"batchStatus,omitempty"`
	AdyenPSPReference    *string    `json:"adyenPspReference,omitempty"`
	AdyenSessionID       *string    `json:"adyenSessionId,omitempty"`
	AdyenPaymentPayload  *string    `json:"adyenPaymentPayload,omitempty"`
	AdyenPaymentResponse *string    `json:"adyenPaymentResponse,omitempty"`
	DmeBatchRequest      *string    `json:"dmeBatchRequest,omitempty"`
	DmeBatchResponse     *string    `json:"dmeBatchResponse,omitempty"`
	BatchID              *string    `json:"batchId,omitempty"`
	BatchPaymentID       *uuid.UUID `json:"batchPaymentId,omitempty"`
	PaymentDate          *time.Time `json:"paymentDate,omitempty"`
	AuthorizedAt         *time.Time `json:"authorizedAt,omitempty"`
	CompletedAt          *time.Time `json:"completedAt,omitempty"`
	FailedAt             *time.Time `json:"failedAt,omitempty"`
	CustomerID           *string    `json:"customerId,omitempty"`
	LocationCode         *string    `json:"locationCode,omitempty"`
	TransactionID        *string    `json:"transactionId,omitempty"`
	AuthCode             *string    `json:"authCode,omitempty"`
	ErrorMessage         *string    `json:"errorMessage,omitempty"`
	ErrorCode            *string    `json:"errorCode,omitempty"`
	InternalNotes        *string    `json:"internalNotes,omitempty"`
	CreatedAt            time.Time  `json:"createdAt"`
	UpdatedAt            *time.Time `json:"updatedAt,omitempty"`
}

// PaymentListResponse represents a paginated list of payments
type PaymentListResponse struct {
	Data       []PaymentResponse `json:"data"`
	Total      int64             `json:"total"`
	Page       int               `json:"page"`
	PageSize   int               `json:"pageSize"`
	TotalPages int               `json:"totalPages"`
}

// PaymentStatsResponse represents payment statistics
type PaymentStatsResponse struct {
	TotalCount           int64  `json:"totalCount"`
	CompletedCount       int64  `json:"completedCount"`
	FailedCount          int64  `json:"failedCount"`
	PendingCount         int64  `json:"pendingCount"`
	AuthorizedCount      int64  `json:"authorizedCount"`
	TotalCompletedAmount string `json:"totalCompletedAmount"`
	StartDate            string `json:"startDate,omitempty"`
	EndDate              string `json:"endDate,omitempty"`
}

// PaymentLinkResponse represents the response for a generated payment link
type PaymentLinkResponse struct {
	Token     string `json:"token"`
	URL       string `json:"url"`
	ExpiresAt string `json:"expiresAt"`
}

func NewPaymentLinkResponse(frontendBaseURL, token string, expiresAt time.Time) PaymentLinkResponse {
	return PaymentLinkResponse{
		Token:     token,
		URL:       fmt.Sprintf("%s/payment/%s", frontendBaseURL, token),
		ExpiresAt: expiresAt.Format(time.RFC3339),
	}
}

// PaymentTokenValidationResponse represents the response for a valid token check
type PaymentTokenValidationResponse struct {
	CustomerID string `json:"customerId"`
	MarinaID   string `json:"marinaId"`
}

// ConvertPaymentToResponse converts a db.Payment to PaymentResponse
func ConvertPaymentToResponse(payment db.Payment) PaymentResponse {
	// Convert CreatedAt timestamp
	var createdAt time.Time
	if payment.CreatedAt.Valid {
		createdAt = payment.CreatedAt.Time
	}

	response := PaymentResponse{
		ID:              payment.ID,
		MarinaID:        payment.MarinaID,
		OrganizationID:  payment.OrganizationID,
		EntityType:      payment.EntityType,
		EntityID:        payment.EntityID,
		Amount:          utils.NumericToString(payment.Amount),
		Currency:        payment.Currency,
		PaymentMethod:   payment.PaymentMethod,
		ReferenceNumber: payment.ReferenceNumber,
		Status:          payment.Status,
		CustomerID:      payment.CustomerID,
		LocationCode:    payment.LocationCode,
		InternalNotes:   payment.InternalNotes,
		CreatedAt:       createdAt,
	}

	// Convert optional fields
	if payment.AuthorizationStatus != nil {
		response.AuthorizationStatus = payment.AuthorizationStatus
	}
	if payment.BatchStatus != nil {
		response.BatchStatus = payment.BatchStatus
	}
	if payment.AdyenPspReference != nil {
		response.AdyenPSPReference = payment.AdyenPspReference
	}
	if payment.AdyenSessionID != nil {
		response.AdyenSessionID = payment.AdyenSessionID
	}
	if payment.AdyenPaymentPayload != nil {
		response.AdyenPaymentPayload = payment.AdyenPaymentPayload
	}
	if payment.AdyenPaymentResponse != nil {
		response.AdyenPaymentResponse = payment.AdyenPaymentResponse
	}
	if payment.DmeBatchRequest != nil {
		response.DmeBatchRequest = payment.DmeBatchRequest
	}
	if payment.DmeBatchResponse != nil {
		response.DmeBatchResponse = payment.DmeBatchResponse
	}
	if payment.BatchID != nil {
		response.BatchID = payment.BatchID
	}
	// BatchPaymentID is uuid.UUID (zero value is empty UUID, not null)
	if payment.BatchPaymentID != uuid.Nil {
		response.BatchPaymentID = &payment.BatchPaymentID
	}
	if payment.TransactionID != nil {
		response.TransactionID = payment.TransactionID
	}
	if payment.AuthCode != nil {
		response.AuthCode = payment.AuthCode
	}
	if payment.ErrorMessage != nil {
		response.ErrorMessage = payment.ErrorMessage
	}
	if payment.ErrorCode != nil {
		response.ErrorCode = payment.ErrorCode
	}

	// Convert timestamp fields
	if payment.PaymentDate.Valid {
		t := payment.PaymentDate.Time
		response.PaymentDate = &t
	}
	if payment.AuthorizedAt.Valid {
		t := payment.AuthorizedAt.Time
		response.AuthorizedAt = &t
	}
	if payment.CompletedAt.Valid {
		t := payment.CompletedAt.Time
		response.CompletedAt = &t
	}
	if payment.FailedAt.Valid {
		t := payment.FailedAt.Time
		response.FailedAt = &t
	}
	if payment.UpdatedAt.Valid {
		t := payment.UpdatedAt.Time
		response.UpdatedAt = &t
	}

	return response
}

// swag:response PaymentResponse
type SwagPaymentResponse = PaymentResponse

// swag:response PaymentListResponse
type SwagPaymentListResponse = PaymentListResponse

// swag:response PaymentStatsResponse
type SwagPaymentStatsResponse = PaymentStatsResponse
