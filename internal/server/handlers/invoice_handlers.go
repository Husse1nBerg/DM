package handlers

import (
	"fmt"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"

	// "github.com/dockworks/dm-web-backend/pkg/adyen"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/google/uuid"
)

// InvoiceHandler handles invoice-related requests
type InvoiceHandler struct {
	server *s.Server
}

// NewInvoiceHandler creates a new invoice handler
func NewInvoiceHandler(server *s.Server) *InvoiceHandler {
	return &InvoiceHandler{
		server: server,
	}
}

// @Summary Get customer invoices
// @Description Retrieves invoices for a specific customer. If `token` is provided, validates the short-lived payment token and uses its customer/marina context. Otherwise, expects authenticated user context.
// @Tags Invoices
// @Accept json
// @Produce json
// @Param customerId query string true "Customer ID"
// @Param invoiceDate query string false "Invoice date (format: YYYY-MM-DD)"
// @Param marinaId query string true "Marina ID" Format(uuid)
// @Param token query string false "Short-lived payment token"
// @Success 200 {object} responses.InvoiceListResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Security BearerAuth
// @Router /invoices/customer [get]
func (h *InvoiceHandler) GetCustomerInvoices(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.GetCustomerInvoicesRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Require either a valid JWT (set by middleware) or a short-lived payment token
	if c.Get("user") == nil && req.Token == "" {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Authorization required: Bearer token or payment token").JSON(c)
	}

	// If token is provided, validate and override customer/marina
	if req.Token != "" {
		link, err := h.server.DB.Queries().GetValidPaymentLinkByToken(ctx, req.Token)
		if err != nil {
			return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid or expired token").JSON(c)
		}
		if link.ExpiresAt.Time.Before(time.Now()) || link.Revoked {
			return responses.NewErrorResponse(http.StatusUnauthorized, "Token expired or revoked").JSON(c)
		}
		// Override request values to enforce token scope
		req.CustomerID = link.CustomerID
		req.MarinaID = link.MarinaID
	} else {
		// No token: ensure required query params are present
		if req.CustomerID == "" || req.MarinaID == uuid.Nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "customerId and marinaId are required when no token is provided").JSON(c)
		}
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), req.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	dmeResponse, err := h.server.DME.RetrieveCustomerInvoices(ctx, req.CustomerID, req.InvoiceDate, orgID, *systemID)

	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve customer invoices",
			zap.Error(err),
			zap.String("customerId", req.CustomerID),
			zap.String("invoiceDate", req.InvoiceDate))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertCustomerInvoiceInquiryList(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// GetCustomerInvoice godoc
//
//	@Summary		Get specific customer invoice
//	@Description	Retrieves a specific invoice for a customer. If `token` is provided, validates the short-lived payment token and uses its customer/marina context; otherwise expects authenticated user with `customerId` and `marinaId`.
//	@Tags			Invoices
//	@Accept			json
//	@Produce		json
//	@Param			invoiceId	query	string	true	"Invoice ID"
//	@Param			customerId	query	string	false	"Customer ID (required without token)"
//	@Param			marinaId	query	string	false	"Marina ID (UUID, required without token)" Format(uuid)
//	@Param			token		query	string	false	"Short-lived payment token"
//	@Success		200		{object}	responses.InvoiceResponse
//	@Failure		400		{object}	responses.Error
//	@Failure		401		{object}	responses.Error
//	@Failure		404		{object}	responses.Error
//	@Failure		500		{object}	responses.Error
//	@Security		BearerAuth
//	@Router			/invoices/customer/invoice [get]
func (h *InvoiceHandler) GetCustomerInvoice(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.GetCustomerInvoiceRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Require either a valid JWT (set by middleware) or a short-lived payment token
	if c.Get("user") == nil && req.Token == "" {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Authorization required: Bearer token or payment token").JSON(c)
	}

	// Resolve context
	if req.Token != "" {
		link, err := h.server.DB.Queries().GetValidPaymentLinkByToken(ctx, req.Token)
		if err != nil || link.ExpiresAt.Time.Before(time.Now()) || link.Revoked {
			return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid or expired token").JSON(c)
		}
		req.CustomerID = link.CustomerID
		req.MarinaID = link.MarinaID
	} else {
		if req.CustomerID == "" {
			return responses.NewErrorResponse(http.StatusBadRequest, "customerId is required when no token is provided").JSON(c)
		}
		// If marinaId is not provided, derive it from the authenticated user context
		if req.MarinaID == uuid.Nil {
			userToken := c.Get("user").(*jwt.Token)
			claims := userToken.Claims.(*token.JwtCustomClaims)
			user, err := h.server.DB.Queries().GetUserByID(ctx, claims.ID)
			if err != nil {
				return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user").JSON(c)
			}
			req.MarinaID = user.MarinaID
		}
	}

	// Load marina/org/system
	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, req.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}
	if marina.SystemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	// Retrieve invoice by ID
	invoices, err := h.server.DME.RetrieveInvoices(ctx, []string{req.InvoiceID}, marina.OrganizationID, *marina.SystemID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to retrieve invoice", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve invoice").JSON(c)
	}
	if len(invoices) == 0 {
		return responses.NewErrorResponse(http.StatusNotFound, "Invoice not found").JSON(c)
	}

	inv := invoices[0]
	// Ensure the invoice belongs to the resolved customer
	if inv.CustomerID != req.CustomerID {
		return responses.NewErrorResponse(http.StatusForbidden, "Access denied").JSON(c)
	}

	return c.JSON(http.StatusOK, responses.InvoiceResponse{Data: inv})
}

// GetNextReference godoc
// @Summary Get next AR reference number
// @Description Retrieves the next available AR reference number for a given customer from DME
// @Tags Invoices
// @Accept json
// @Produce json
// @Param customerId query string true "Customer ID"
// @Success 200 {object} responses.NextReferenceResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Security BearerAuth
// @Router /invoices/next-reference [get]
func (h *InvoiceHandler) GetNextReference(c echo.Context) error {
	ctx := c.Request().Context()

	var req requests.GetNextReferenceRequest
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

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina").JSON(c)
	}

	if marina.SystemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	ref, err := h.server.DME.NextReferenceNumber(ctx, req.CustomerID, marina.OrganizationID, *marina.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve next reference number",
			zap.Error(err),
			zap.String("customer_id", req.CustomerID),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve next reference number").JSON(c)
	}

	return c.JSON(http.StatusOK, responses.NextReferenceResponse{ReferenceNumber: ref})
}

// SubmitBatch godoc
// @Summary Submit a batch of payments
// @Description Submits a batch of cash receipts to DME for processing
// @Tags Invoices
// @Accept json
// @Produce json
// @Param request body requests.SubmitBatchRequest true "Batch submission request"
// @Success 200 {object} responses.BatchSubmissionResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Security BearerAuth
// @Router /invoices/batch/submit [post]
func (h *InvoiceHandler) SubmitBatch(c echo.Context) error {

	var req requests.SubmitBatchRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get user context (assuming this is called from a protected route)
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID

	// Get user and marina information
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to get user",
			zap.Error(err),
			zap.String("user_id", userID.String()),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get user information")
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to get marina",
			zap.Error(err),
			zap.String("marina_id", user.MarinaID.String()),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get marina information")
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Marina system ID is not configured")
	}

	// Convert request cash receipts to DME format
	var dmeCashReceipts []dme.CashReceipt
	totalAmount := 0.0
	for _, receipt := range req.CashReceipts {
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
		c.Request().Context(),
		req.LocationCode,
		dmeCashReceipts,
		req.PostBatch,
		orgID,
		*systemID,
	)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to submit batch to DME",
			zap.Error(err),
			zap.String("location_code", req.LocationCode),
			zap.Int("receipt_count", len(req.CashReceipts)),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to submit batch: "+err.Error())
	}

	// Convert totalAmount to Numeric
	var totalAmountNumeric pgtype.Numeric
	err = totalAmountNumeric.Scan(fmt.Sprintf("%.2f", totalAmount))
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to convert total amount to Numeric",
			zap.Error(err),
			zap.Float64("total_amount", totalAmount),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process total amount")
	}

	// Convert submittedAt to Timestamptz
	var submittedAtTimestamptz pgtype.Timestamptz
	err = submittedAtTimestamptz.Scan(dmeResponse.SubmittedAt)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to convert submitted at to Timestamptz",
			zap.Error(err),
			zap.Time("submitted_at", dmeResponse.SubmittedAt),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process submission time")
	}

	// Create batch payment record in database
	batchPayment, err := h.server.DB.Queries().CreateBatchPayment(c.Request().Context(), db.CreateBatchPaymentParams{
		OrganizationID: orgID,
		MarinaID:       user.MarinaID,
		LocationCode:   req.LocationCode,
		BatchID:        dmeResponse.BatchID,
		PostBatch:      req.PostBatch,
		TotalAmount:    totalAmountNumeric,
		ReceiptCount:   int32(len(req.CashReceipts)),
		Status:         "submitted",
		SubmittedBy:    userID.String(),
		SubmittedAt:    submittedAtTimestamptz,
	})
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to create batch payment record",
			zap.Error(err),
			zap.String("batch_id", dmeResponse.BatchID),
		)
		// Don't fail the request, just log the error
	}

	// Create individual receipt records
	for _, receipt := range req.CashReceipts {
		// Convert total payment amount to Numeric
		var amountNumeric pgtype.Numeric
		err = amountNumeric.Scan(fmt.Sprintf("%.2f", receipt.TotalPayment))
		if err != nil {
			h.server.Logger.DesugarZap.Error("Failed to convert receipt total payment to Numeric",
				zap.Error(err),
				zap.Float64("total_payment", receipt.TotalPayment),
				zap.String("customer_id", receipt.CustomerID),
			)
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
					h.server.Logger.DesugarZap.Error("Failed to parse payment timestamp",
						zap.Error(err),
						zap.String("cc_transaction_timestamp", receipt.CCTransactionTimeStamp),
						zap.String("customer_id", receipt.CustomerID),
					)
					// Use current time as fallback
					paymentDate = time.Now()
				}
			}
			err = paymentDateTimestamptz.Scan(paymentDate)
			if err != nil {
				h.server.Logger.DesugarZap.Error("Failed to convert payment timestamp to Timestamptz",
					zap.Error(err),
					zap.Time("payment_date", paymentDate),
					zap.String("customer_id", receipt.CustomerID),
				)
				// Use current time as fallback
				paymentDateTimestamptz.Scan(time.Now())
			}
		} else {
			// Use current time if no timestamp provided
			paymentDateTimestamptz.Scan(time.Now())
		}

		// Get invoice ID from first inv payment or use empty string if none
		invoiceID := ""
		if len(receipt.InvPayments) > 0 {
			invoiceID = receipt.InvPayments[0].InvoiceID
		}

		_, err = h.server.DB.Queries().CreateBatchPaymentReceipt(c.Request().Context(), db.CreateBatchPaymentReceiptParams{
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
			h.server.Logger.DesugarZap.Error("Failed to create batch payment receipt record",
				zap.Error(err),
				zap.String("batch_payment_id", batchPayment.ID.String()),
				zap.String("customer_id", receipt.CustomerID),
			)
			// Continue processing other receipts
		}
	}

	// Update batch payment with reference IDs if available
	if len(dmeResponse.ReferenceIDs) > 0 {
		_, err = h.server.DB.Queries().UpdateBatchPaymentStatus(c.Request().Context(), db.UpdateBatchPaymentStatusParams{
			ID:           batchPayment.ID,
			Status:       "completed",
			PostResult:   &dmeResponse.PostResult,
			ReferenceIds: dmeResponse.ReferenceIDs,
		})
		if err != nil {
			h.server.Logger.DesugarZap.Error("Failed to update batch payment status",
				zap.Error(err),
				zap.String("batch_payment_id", batchPayment.ID.String()),
			)
		}
	}

	// Convert to response format
	response := responses.BatchSubmissionResponse{
		BatchID:      dmeResponse.BatchID,
		LocationCode: dmeResponse.LocationCode,
		PostBatch:    dmeResponse.PostBatch,
		ReferenceIDs: dmeResponse.ReferenceIDs,
		PostResult:   dmeResponse.PostResult,
		SubmittedAt:  dmeResponse.SubmittedAt,
		TotalAmount:  dmeResponse.TotalAmount,
		ReceiptCount: dmeResponse.ReceiptCount,
	}

	return c.JSON(http.StatusOK, response)
}
