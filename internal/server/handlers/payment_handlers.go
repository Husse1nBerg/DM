package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/adyen"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	server *s.Server
}

func NewPaymentHandler(server *s.Server) *PaymentHandler {
	return &PaymentHandler{server: server}
}

// CreatePaymentSession creates a new payment session
//
//	@Summary		Create payment session
//	@Description	Creates a new Adyen payment session for processing payments
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.CreatePaymentSessionRequest	true	"Payment session request"
//	@Success		200		{object}	responses.PaymentSessionResponse		"Payment session created successfully"
//	@Failure		400		{object}	responses.Error				"Invalid request"
//	@Failure		500		{object}	responses.Error				"Internal server error"
//	@Router			/payments/sessions [post]
func (h *PaymentHandler) CreatePaymentSession(c echo.Context) error {
	var req requests.CreatePaymentSessionRequest
	if err := c.Bind(&req); err != nil {
		h.server.Logger.Zap.Error("Failed to bind payment session request", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		h.server.Logger.Zap.Error("Payment session request validation failed", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Validation failed").JSON(c)
	}

	// Convert request to Adyen service request
	adyenReq := adyen.CreateCheckoutSessionRequest{
		Amount:      req.Amount,
		Currency:    req.Currency,
		CountryCode: req.CountryCode,
		ReturnURL:   req.ReturnURL,
		ShopperIP:   req.ShopperIP,
		LineItems:   req.LineItems,
		Metadata:    req.Metadata,
	}

	// Create checkout session
	session, err := h.server.PaymentService.CreateCheckoutSession(context.Background(), adyenReq)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to create payment session", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to create payment session").JSON(c)
	}

	// Return the complete Adyen session response directly like the Go example
	// Don't add clientKey to the session response - pass it separately in frontend
	return c.JSON(http.StatusOK, session)
}

// HandlePaymentRedirect handles payment completion redirects
//
//	@Summary		Handle payment redirect
//	@Description	Handles payment completion redirects from Adyen
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Param			request	body		requests.PaymentDetailsRequest	true	"Payment details request"
//	@Success		200		{object}	responses.PaymentResultResponse	"Payment result"
//	@Failure		400		{object}	responses.Error			"Invalid request"
//	@Failure		500		{object}	responses.Error			"Internal server error"
//	@Router			/payments/details [post]
func (h *PaymentHandler) HandlePaymentRedirect(c echo.Context) error {
	var req requests.PaymentDetailsRequest
	if err := c.Bind(&req); err != nil {
		h.server.Logger.Zap.Error("Failed to bind payment details request", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		h.server.Logger.Zap.Error("Payment details request validation failed", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Validation failed").JSON(c)
	}

	// Convert request to Adyen service request
	adyenReq := adyen.PaymentDetailsRequest{
		PaymentData:    req.PaymentData,
		RedirectResult: req.RedirectResult,
		Payload:        req.Payload,
	}

	// Handle payment details
	result, err := h.server.PaymentService.HandlePaymentDetails(context.Background(), adyenReq)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to handle payment details", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to process payment").JSON(c)
	}

	// Create response
	response := responses.NewPaymentResultResponse(result)
	return responses.NewSuccessResponse(response).JSON(c)
}

// ProcessWebhook processes incoming Adyen webhooks
//
//	@Summary		Process webhook
//	@Description	Processes incoming webhook notifications from Adyen
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.WebhookResponse	"Webhook processed successfully"
//	@Failure		401	{object}	responses.Error		"Invalid HMAC signature"
//	@Failure		500	{object}	responses.Error		"Internal server error"
//	@Router			/payments/webhooks [post]
func (h *PaymentHandler) ProcessWebhook(c echo.Context) error {
	// Read the request body
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to read webhook body", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Failed to read request body").JSON(c)
	}

	// Process webhook
	notificationRequest, err := h.server.PaymentService.ProcessWebhook(string(bodyBytes))
	if err != nil {
		h.server.Logger.Zap.Error("Failed to process webhook", zap.Error(err))
		return responses.NewErrorResponse(http.StatusBadRequest, "Failed to process webhook").JSON(c)
	}

	// Validate HMAC signature for each notification
	notifications := notificationRequest.GetNotificationItems()
	if len(notifications) == 0 {
		h.server.Logger.Zap.Warn("No notification items found in webhook")
		return responses.NewErrorResponse(http.StatusBadRequest, "No notification items found").JSON(c)
	}

	for i, notification := range notifications {
		if notification == nil {
			h.server.Logger.Zap.Error("Nil notification item found", zap.Int("index", i))
			continue
		}

		if !h.server.PaymentService.ValidateWebhook(*notification) {
			pspRef := ""
			eventCode := ""
			if notification.PspReference != "" {
				pspRef = notification.PspReference
			}
			if notification.EventCode != "" {
				eventCode = notification.EventCode
			}
			h.server.Logger.Zap.Error("Invalid HMAC signature for webhook",
				zap.String("psp_reference", pspRef),
				zap.String("event_code", eventCode))
			return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid HMAC signature").JSON(c)
		}

		// Process the notification asynchronously
		go h.processNotification(*notification)
	}

	response := &responses.WebhookResponse{
		Message: "Webhook processed successfully",
		Status:  "accepted",
	}

	return responses.NewSuccessResponse(response).JSON(c)
}

// processNotification processes a webhook notification asynchronously
func (h *PaymentHandler) processNotification(notification interface{}) {
	ctx := context.Background()

	// Extract and parse metadata from additionalData
	// Adyen sends metadata in additionalData with "metadata." prefix
	var notifMap map[string]interface{}
	if notifBytes, err := json.Marshal(notification); err == nil {
		if err := json.Unmarshal(notifBytes, &notifMap); err == nil {
			// Check for additionalData field
			if additionalData, ok := notifMap["additionalData"].(map[string]interface{}); ok && additionalData != nil {
				// Extract metadata fields from additionalData (they have "metadata." prefix)
				batchDataStr, hasBatchData := additionalData["metadata.BatchData"].(string)
				marinaIDStr, _ := additionalData["metadata.MarinaID"].(string)
				entityType, _ := additionalData["metadata.EntityType"].(string)
				entityID, _ := additionalData["metadata.EntityID"].(string)
				sessionID, _ := additionalData["checkoutSessionId"].(string)

				if hasBatchData && batchDataStr != "" {
					// Parse BatchData JSON string
					var batchData requests.SubmitBatchRequest
					if err := json.Unmarshal([]byte(batchDataStr), &batchData); err != nil {
						h.server.Logger.Zap.Error("Failed to parse BatchData from webhook",
							zap.Error(err))
						return
					}

					// Parse marina ID
					marinaID, err := uuid.Parse(marinaIDStr)
					if err != nil {
						h.server.Logger.Zap.Error("Invalid marina ID",
							zap.Error(err))
						return
					}

					// Get marina to find organization ID
					marina, err := h.server.DB.Queries().GetMarinaByID(ctx, marinaID)
					if err != nil {
						h.server.Logger.Zap.Error("Failed to get marina",
							zap.Error(err))
						return
					}

					// Check if payment was successful
					if eventCode, ok := notifMap["eventCode"].(string); ok && eventCode == "AUTHORISATION" {
						if success, ok := notifMap["success"].(string); ok && success == "true" {
							// Step 1: Create payment record with "authorized" status
							paymentRecord, err := h.createPaymentRecord(ctx, notifMap, additionalData, batchData, marinaID, marina.OrganizationID, sessionID, entityType, entityID)
							if err != nil {
								h.server.Logger.Zap.Error("Failed to create payment record",
									zap.Error(err))
							}

							// Create metadata map for submitBatchToDME
							metadataForDME := map[string]interface{}{
								"MarinaID": marinaIDStr,
							}

							// Step 2: Submit to DME and update payment to "completed"
							if err := h.submitBatchToDME(batchData, metadataForDME, paymentRecord); err != nil {
								h.server.Logger.Zap.Error("Failed to submit batch to DME",
									zap.Error(err))

								// Update payment to failed if we have a payment record
								if paymentRecord != nil {
									h.updatePaymentFailed(ctx, paymentRecord.ID, err.Error())
								}
							} else {
								h.server.Logger.Zap.Info("Successfully submitted batch to DME",
									zap.String("reference", batchData.CashReceipts[0].ReferenceNum),
									zap.Int("receipt_count", len(batchData.CashReceipts)))
							}
						} else {
							h.server.Logger.Zap.Warn("Payment not successful, skipping DME submission",
								zap.String("success", success))
						}
					}
				}
			}
		}
	}
}

// submitBatchToDME submits batch data to DME when payment is successful
func (h *PaymentHandler) submitBatchToDME(batchData requests.SubmitBatchRequest, metadataMap map[string]interface{}, paymentRecord *db.Payment) error {
	ctx := context.Background()

	// Get marina ID from metadata
	marinaIDStr, ok := metadataMap["MarinaID"].(string)
	if !ok || marinaIDStr == "" {
		return fmt.Errorf("MarinaID not found in metadata")
	}

	// Parse marina ID (convert string to uuid.UUID)
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return fmt.Errorf("invalid MarinaID in metadata: %w", err)
	}

	// Get marina information
	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, marinaID)
	if err != nil {
		return fmt.Errorf("failed to get marina: %w", err)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return fmt.Errorf("marina system ID is not configured")
	}

	// Serialize batch request for logging
	batchRequestJSON, _ := json.Marshal(batchData)

	// Convert request cash receipts to DME format
	var dmeCashReceipts []dme.CashReceipt
	totalAmount := 0.0
	for _, receipt := range batchData.CashReceipts {
		// Convert InvPayments to DME format
		var dmeInvPayments []dme.InvPayment
		for _, invPayment := range receipt.InvPayments {
			dmeInvPayments = append(dmeInvPayments, dme.InvPayment{
				InvoiceID:    invPayment.InvoiceID,
				LocationCode: invPayment.LocationCode,
				DepositType:  invPayment.DepositType,
				PaymentAmt:   invPayment.PaymentAmt,
				Description:  invPayment.Description,
				CustomerID:   invPayment.CustomerID,
			})
		}

		dmeCashReceipts = append(dmeCashReceipts, dme.CashReceipt{
			CustomerID:             receipt.CustomerID,
			ReferenceNum:           receipt.ReferenceNum,
			PayType:                receipt.PayType,
			TotalPayment:           receipt.TotalPayment,
			StatementDesc:          receipt.StatementDesc,
			CCAuthCode:             receipt.CCAuthCode,
			CCTransactionID:        receipt.CCTransactionID,
			CCTransactionTimeStamp: receipt.CCTransactionTimeStamp,
			CCSurcharge:            receipt.CCSurcharge,
			CCSurchargeTax:         receipt.CCSurchargeTax,
			CCSurchargeTaxSchema:   receipt.CCSurchargeTaxSchema,
			CCSurchargeTaxIds:      receipt.CCSurchargeTaxIds,
			InvPayments:            dmeInvPayments,
		})
		totalAmount += receipt.TotalPayment
	}

	// Submit batch to DME
	dmeResponse, err := h.server.DME.SubmitBatch(
		ctx,
		batchData.LocationCode,
		dmeCashReceipts,
		batchData.PostBatch,
		orgID,
		*systemID,
	)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to submit batch to DME",
			zap.Error(err),
			zap.String("location_code", batchData.LocationCode),
			zap.Int("receipt_count", len(batchData.CashReceipts)))
		return fmt.Errorf("failed to submit batch to DME: %w", err)
	}

	// Convert totalAmount to Numeric
	var totalAmountNumeric pgtype.Numeric
	if err := totalAmountNumeric.Scan(fmt.Sprintf("%.2f", totalAmount)); err != nil {
		h.server.Logger.Zap.Error("Failed to convert total amount to Numeric",
			zap.Error(err),
			zap.Float64("total_amount", totalAmount))
		return fmt.Errorf("failed to process total amount: %w", err)
	}

	// Convert submittedAt to Timestamptz
	var submittedAtTimestamptz pgtype.Timestamptz
	if err := submittedAtTimestamptz.Scan(dmeResponse.SubmittedAt); err != nil {
		h.server.Logger.Zap.Error("Failed to convert submitted at to Timestamptz",
			zap.Error(err),
			zap.Time("submitted_at", dmeResponse.SubmittedAt))
		return fmt.Errorf("failed to process submission time: %w", err)
	}

	// Create batch payment record in database
	batchPayment, err := h.server.DB.Queries().CreateBatchPayment(ctx, db.CreateBatchPaymentParams{
		OrganizationID: orgID,
		MarinaID:       marinaID,
		LocationCode:   batchData.LocationCode,
		BatchID:        dmeResponse.BatchID,
		PostBatch:      batchData.PostBatch,
		TotalAmount:    totalAmountNumeric,
		ReceiptCount:   int32(len(batchData.CashReceipts)),
		Status:         "submitted",
		SubmittedBy:    "webhook",
		SubmittedAt:    submittedAtTimestamptz,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Failed to create batch payment record",
			zap.Error(err),
			zap.String("batch_id", dmeResponse.BatchID))
		// Don't fail the request, just log the error
	} else {
		// Create individual receipt records
		for _, receipt := range batchData.CashReceipts {
			// Convert total payment amount to Numeric
			var amountNumeric pgtype.Numeric
			if err := amountNumeric.Scan(fmt.Sprintf("%.2f", receipt.TotalPayment)); err != nil {
				h.server.Logger.Zap.Error("Failed to convert receipt total payment to Numeric",
					zap.Error(err),
					zap.Float64("total_payment", receipt.TotalPayment),
					zap.String("customer_id", receipt.CustomerID))
				continue
			}

			// Parse payment timestamp if available
			var paymentDateTimestamptz pgtype.Timestamptz
			if receipt.CCTransactionTimeStamp != "" {
				paymentDate, err := time.Parse("2006-01-02T15:04:05.000Z", receipt.CCTransactionTimeStamp)
				if err != nil {
					// Try alternative format
					paymentDate, err = time.Parse("2006-01-02T15:04:05Z", receipt.CCTransactionTimeStamp)
					if err != nil {
						h.server.Logger.Zap.Error("Failed to parse payment timestamp",
							zap.Error(err),
							zap.String("cc_transaction_timestamp", receipt.CCTransactionTimeStamp))
						paymentDate = time.Now()
					}
				}
				paymentDateTimestamptz.Scan(paymentDate)
			} else {
				paymentDateTimestamptz.Scan(time.Now())
			}

			// Get invoice ID from first inv payment or use empty string if none
			invoiceID := ""
			if len(receipt.InvPayments) > 0 {
				invoiceID = receipt.InvPayments[0].InvoiceID
			}

			_, err = h.server.DB.Queries().CreateBatchPaymentReceipt(ctx, db.CreateBatchPaymentReceiptParams{
				BatchPaymentID: batchPayment.ID,
				CustomerID:     receipt.CustomerID,
				InvoiceID:      invoiceID,
				Amount:         amountNumeric,
				PaymentMethod:  receipt.PayType,
				Reference:      receipt.ReferenceNum,
				Description:    &receipt.StatementDesc,
				PaymentDate:    paymentDateTimestamptz,
			})
			if err != nil {
				h.server.Logger.Zap.Error("Failed to create batch payment receipt record",
					zap.Error(err),
					zap.String("batch_payment_id", batchPayment.ID.String()),
					zap.String("customer_id", receipt.CustomerID))
			}
		}

		// Update batch payment with reference IDs if available
		if len(dmeResponse.ReferenceIDs) > 0 {
			_, err = h.server.DB.Queries().UpdateBatchPaymentStatus(ctx, db.UpdateBatchPaymentStatusParams{
				ID:           batchPayment.ID,
				Status:       "completed",
				PostResult:   &dmeResponse.PostResult,
				ReferenceIds: dmeResponse.ReferenceIDs,
			})
			if err != nil {
				h.server.Logger.Zap.Error("Failed to update batch payment status",
					zap.Error(err),
					zap.String("batch_payment_id", batchPayment.ID.String()))
			}
		}

		// Update payment record to completed if we have one
		if paymentRecord != nil {
			batchResponseJSONBytes, _ := json.Marshal(dmeResponse)
			batchResponseJSON := string(batchResponseJSONBytes)
			batchRequestJSONStr := string(batchRequestJSON)
			completedAt := pgtype.Timestamptz{}
			completedAt.Scan(time.Now())

			_, err = h.server.DB.Queries().UpdatePaymentCompleted(ctx, db.UpdatePaymentCompletedParams{
				ID:               paymentRecord.ID,
				BatchID:          &dmeResponse.BatchID,
				BatchPaymentID:   batchPayment.ID,
				DmeBatchRequest:  &batchRequestJSONStr,
				DmeBatchResponse: &batchResponseJSON,
				CompletedAt:      completedAt,
			})
			if err != nil {
				h.server.Logger.Zap.Error("Failed to update payment to completed",
					zap.Error(err))
			}
		}
	}

	return nil
}

// createPaymentRecord creates a payment record with authorized status
func (h *PaymentHandler) createPaymentRecord(ctx context.Context, notifMap map[string]interface{}, additionalData map[string]interface{}, batchData requests.SubmitBatchRequest, marinaID uuid.UUID, orgID uuid.UUID, sessionID string, metadataEntityType string, metadataEntityID string) (*db.Payment, error) {
	// Extract payment details from webhook
	pspReference, _ := notifMap["pspReference"].(string)
	authCode, _ := additionalData["authCode"].(string)
	transactionID, _ := additionalData["networkTxReference"].(string)

	// Get the first cash receipt for payment details
	if len(batchData.CashReceipts) == 0 {
		return nil, fmt.Errorf("no cash receipts found in batch data")
	}
	receipt := batchData.CashReceipts[0]

	// Determine entity type and ID - use metadata if provided, otherwise fall back to invoice payments
	entityType := metadataEntityType
	entityID := metadataEntityID

	// Fallback to invoice logic if not provided in metadata
	if entityType == "" {
		entityType = "invoice"
	}
	if entityID == "" && len(receipt.InvPayments) > 0 {
		entityID = receipt.InvPayments[0].InvoiceID
	}

	// Get payment amount from notifMap
	var amountNumeric pgtype.Numeric
	if amountMap, ok := notifMap["amount"].(map[string]interface{}); ok {
		if value, ok := amountMap["value"].(float64); ok {
			amount := value / 100 // Convert from minor units
			amountNumeric.Scan(fmt.Sprintf("%.2f", amount))
		}
	}

	// Get currency
	currency := "USD"
	if amountMap, ok := notifMap["amount"].(map[string]interface{}); ok {
		if curr, ok := amountMap["currency"].(string); ok {
			currency = curr
		}
	}

	// Parse payment timestamp
	paymentDate := pgtype.Timestamptz{}
	paymentDate.Scan(time.Now())

	authorizedAt := pgtype.Timestamptz{}
	authorizedAt.Scan(time.Now())

	// Serialize webhook payload
	webhookJSONBytes, _ := json.Marshal(notifMap)
	webhookJSON := string(webhookJSONBytes)

	// Get payment method from additionalData
	paymentMethod, _ := additionalData["paymentMethod"].(string)

	// Create payment record
	payment, err := h.server.DB.Queries().CreatePayment(ctx, db.CreatePaymentParams{
		MarinaID:            marinaID,
		OrganizationID:      orgID,
		EntityType:          &entityType,
		EntityID:            &entityID,
		Amount:              amountNumeric,
		Currency:            currency,
		PaymentMethod:       &paymentMethod,
		ReferenceNumber:     receipt.ReferenceNum,
		Status:              "authorized",
		AuthorizationStatus: &authCode,
		AdyenSessionID:      &sessionID,
		CustomerID:          &receipt.CustomerID,
		LocationCode:        &batchData.LocationCode,
		PaymentDate:         paymentDate,
		InternalNotes:       nil,
	})

	if err != nil {
		return nil, fmt.Errorf("failed to create payment record: %w", err)
	}

	// Update with authorization details
	payment, err = h.server.DB.Queries().UpdatePaymentAuthorized(ctx, db.UpdatePaymentAuthorizedParams{
		ID:                  payment.ID,
		AuthorizationStatus: &authCode,
		AdyenPspReference:   &pspReference,
		AuthCode:            &authCode,
		TransactionID:       &transactionID,
		AuthorizedAt:        authorizedAt,
		AdyenWebhookPayload: &webhookJSON,
	})

	if err != nil {
		h.server.Logger.Zap.Error("Failed to update payment authorization details",
			zap.Error(err))
		return &payment, nil // Return the payment even if update fails
	}

	return &payment, nil
}

// updatePaymentFailed updates a payment record to failed status
func (h *PaymentHandler) updatePaymentFailed(ctx context.Context, paymentID uuid.UUID, errorMsg string) {
	failedAt := pgtype.Timestamptz{}
	failedAt.Scan(time.Now())

	_, err := h.server.DB.Queries().UpdatePaymentFailed(ctx, db.UpdatePaymentFailedParams{
		ID:           paymentID,
		ErrorMessage: &errorMsg,
		ErrorCode:    nil,
		FailedAt:     failedAt,
	})

	if err != nil {
		h.server.Logger.Zap.Error("Failed to update payment to failed status",
			zap.Error(err))
	}
}

// ListPayments godoc
//
//	@Summary		List payments
//	@Description	Retrieves a paginated list of payments with optional filtering
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		true	"Page number"	minimum(1)
//	@Param			pageSize	query		int		true	"Page size"		minimum(1)	maximum(100)
//	@Param			status		query		string	false	"Filter by status (pending, authorized, completed, failed)"
//	@Param			entityType	query		string	false	"Filter by entity type (invoice, boat, customer, etc.)"
//	@Param			entityId	query		string	false	"Filter by entity ID"
//	@Param			startDate	query		string	false	"Filter by start date (YYYY-MM-DD)"
//	@Param			endDate		query		string	false	"Filter by end date (YYYY-MM-DD)"
//	@Success		200			{object}	responses.PaymentListResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		BearerAuth
//	@Router			/payments [get]
func (h *PaymentHandler) ListPayments(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.ListPaymentsRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	}

	// Get user and marina info
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	user, err := h.server.DB.Queries().GetUserByID(ctx, claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user").JSON(c)
	}

	marinaID := user.MarinaID
	offset := int32((req.Page - 1) * req.PageSize)
	limit := int32(req.PageSize)

	var payments []db.Payment
	var total int64

	// Handle different filtering scenarios
	if req.EntityType != "" && req.EntityID != "" {
		// Filter by entity
		payments, err = h.server.DB.Queries().ListPaymentsByEntity(ctx, db.ListPaymentsByEntityParams{
			MarinaID:   marinaID,
			EntityType: &req.EntityType,
			EntityID:   &req.EntityID,
		})
		total = int64(len(payments))
	} else if req.Status != "" {
		// Filter by status
		payments, err = h.server.DB.Queries().ListPaymentsByStatus(ctx, db.ListPaymentsByStatusParams{
			MarinaID: marinaID,
			Status:   req.Status,
			Limit:    limit,
			Offset:   offset,
		})
		if err == nil {
			total, err = h.server.DB.Queries().CountPaymentsByStatus(ctx, db.CountPaymentsByStatusParams{
				MarinaID: marinaID,
				Status:   req.Status,
			})
		}
	} else if req.StartDate != "" && req.EndDate != "" {
		// Filter by date range
		startDate, err1 := time.Parse("2006-01-02", req.StartDate)
		endDate, err2 := time.Parse("2006-01-02", req.EndDate)
		if err1 != nil || err2 != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid date format. Use YYYY-MM-DD").JSON(c)
		}

		var startTimestamp, endTimestamp pgtype.Timestamptz
		startTimestamp.Scan(startDate)
		endTimestamp.Scan(endDate)

		payments, err = h.server.DB.Queries().ListPaymentsByDateRange(ctx, db.ListPaymentsByDateRangeParams{
			MarinaID:      marinaID,
			PaymentDate:   startTimestamp,
			PaymentDate_2: endTimestamp,
		})
		total = int64(len(payments))
	} else {
		// No specific filter, list all
		payments, err = h.server.DB.Queries().ListPaymentsByMarina(ctx, db.ListPaymentsByMarinaParams{
			MarinaID: marinaID,
			Limit:    limit,
			Offset:   offset,
		})
		if err == nil {
			total, err = h.server.DB.Queries().CountPaymentsByMarina(ctx, marinaID)
		}
	}

	if err != nil {
		h.server.Logger.Zap.Error("Failed to list payments", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to list payments").JSON(c)
	}

	// Convert to response format
	paymentResponses := make([]responses.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		paymentResponses = append(paymentResponses, responses.ConvertPaymentToResponse(payment))
	}

	totalPages := int(total) / req.PageSize
	if int(total)%req.PageSize > 0 {
		totalPages++
	}

	response := responses.PaymentListResponse{
		Data:       paymentResponses,
		Total:      total,
		Page:       req.Page,
		PageSize:   req.PageSize,
		TotalPages: totalPages,
	}

	return c.JSON(http.StatusOK, response)
}

// GetPaymentByID godoc
//
//	@Summary		Get payment by ID
//	@Description	Retrieves a single payment by its ID
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Payment ID (UUID)"
//	@Success		200	{object}	responses.PaymentResponse
//	@Failure		400	{object}	responses.Error
//	@Failure		404	{object}	responses.Error
//	@Failure		500	{object}	responses.Error
//	@Security		BearerAuth
//	@Router			/payments/{id} [get]
func (h *PaymentHandler) GetPaymentByID(c echo.Context) error {
	ctx := c.Request().Context()

	paymentIDStr := c.Param("id")
	paymentID, err := uuid.Parse(paymentIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid payment ID").JSON(c)
	}

	// Get user and marina info for authorization
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	user, err := h.server.DB.Queries().GetUserByID(ctx, claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user").JSON(c)
	}

	payment, err := h.server.DB.Queries().GetPaymentByID(ctx, paymentID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to get payment", zap.Error(err))
		return responses.NewErrorResponse(http.StatusNotFound, "Payment not found").JSON(c)
	}

	// Verify payment belongs to user's marina
	if payment.MarinaID != user.MarinaID {
		return responses.NewErrorResponse(http.StatusForbidden, "Access denied").JSON(c)
	}

	response := responses.ConvertPaymentToResponse(payment)
	return c.JSON(http.StatusOK, response)
}

// GetPaymentsByEntity godoc
//
//	@Summary		Get payments by entity
//	@Description	Retrieves all payments for a specific entity (e.g., invoice, boat)
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Param			entityType	query		string	true	"Entity type (invoice, boat, customer, etc.)"
//	@Param			entityId	query		string	true	"Entity ID"
//	@Success		200			{array}		responses.PaymentResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		BearerAuth
//	@Router			/payments/entity [get]
func (h *PaymentHandler) GetPaymentsByEntity(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.GetPaymentsByEntityRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request").JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	}

	// Get user and marina info
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	user, err := h.server.DB.Queries().GetUserByID(ctx, claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user").JSON(c)
	}

	payments, err := h.server.DB.Queries().ListPaymentsByEntity(ctx, db.ListPaymentsByEntityParams{
		MarinaID:   user.MarinaID,
		EntityType: &req.EntityType,
		EntityID:   &req.EntityID,
	})

	if err != nil {
		h.server.Logger.Zap.Error("Failed to get payments by entity", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get payments").JSON(c)
	}

	// Convert to response format
	paymentResponses := make([]responses.PaymentResponse, 0, len(payments))
	for _, payment := range payments {
		paymentResponses = append(paymentResponses, responses.ConvertPaymentToResponse(payment))
	}

	return c.JSON(http.StatusOK, paymentResponses)
}

// GetPaymentStats godoc
//
//	@Summary		Get payment statistics
//	@Description	Retrieves payment statistics for a date range
//	@Tags			Payments
//	@Accept			json
//	@Produce		json
//	@Param			startDate	query		string	false	"Start date (YYYY-MM-DD, defaults to 30 days ago)"
//	@Param			endDate		query		string	false	"End date (YYYY-MM-DD, defaults to today)"
//	@Success		200			{object}	responses.PaymentStatsResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		BearerAuth
//	@Router			/payments/stats [get]
func (h *PaymentHandler) GetPaymentStats(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.GetPaymentStatsRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request").JSON(c)
	}

	// Get user and marina info
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	user, err := h.server.DB.Queries().GetUserByID(ctx, claims.ID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user").JSON(c)
	}

	// Set default date range if not provided (last 30 days)
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -30)

	if req.StartDate != "" {
		parsedStart, err := time.Parse("2006-01-02", req.StartDate)
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid start date format. Use YYYY-MM-DD").JSON(c)
		}
		startDate = parsedStart
	}

	if req.EndDate != "" {
		parsedEnd, err := time.Parse("2006-01-02", req.EndDate)
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid end date format. Use YYYY-MM-DD").JSON(c)
		}
		endDate = parsedEnd
	}

	var startTimestamp, endTimestamp pgtype.Timestamptz
	startTimestamp.Scan(startDate)
	endTimestamp.Scan(endDate)

	stats, err := h.server.DB.Queries().GetPaymentStats(ctx, db.GetPaymentStatsParams{
		MarinaID:      user.MarinaID,
		PaymentDate:   startTimestamp,
		PaymentDate_2: endTimestamp,
	})

	if err != nil {
		h.server.Logger.Zap.Error("Failed to get payment stats", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get payment statistics").JSON(c)
	}

	// Convert TotalCompletedAmount from interface{} to string
	var totalAmountStr string
	if stats.TotalCompletedAmount != nil {
		if numericVal, ok := stats.TotalCompletedAmount.(pgtype.Numeric); ok {
			totalAmountStr = utils.NumericToString(numericVal)
		} else {
			totalAmountStr = "0.00"
		}
	} else {
		totalAmountStr = "0.00"
	}

	response := responses.PaymentStatsResponse{
		TotalCount:           stats.TotalCount,
		CompletedCount:       stats.CompletedCount,
		FailedCount:          stats.FailedCount,
		PendingCount:         stats.PendingCount,
		AuthorizedCount:      stats.AuthorizedCount,
		TotalCompletedAmount: totalAmountStr,
		StartDate:            startDate.Format("2006-01-02"),
		EndDate:              endDate.Format("2006-01-02"),
	}

	return c.JSON(http.StatusOK, response)
}
