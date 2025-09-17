package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	"github.com/dockworks/dm-web-backend/pkg/adyen"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type PaymentHandler struct {
	config       *config.Config
	queries      *db.Queries
	adyenService *adyen.AdyenService
	logger       *zap.Logger
}

func NewPaymentHandler(cfg *config.Config, queries *db.Queries, logger *zap.Logger) (*PaymentHandler, error) {
	adyenService, err := adyen.NewAdyenService(&cfg.Adyen, logger)
	if err != nil {
		return nil, err
	}

	return &PaymentHandler{
		config:       cfg,
		queries:      queries,
		adyenService: adyenService,
		logger:       logger,
	}, nil
}

// CreatePaymentSession godoc
// @Summary Create a new payment session
// @Description Creates a new payment session with Adyen
// @Tags payments
// @Accept json
// @Produce json
// @Param request body requests.CreatePaymentSessionRequest true "Payment session request"
// @Success 200 {object} responses.PaymentSessionResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /payments/session [post]
func (h *PaymentHandler) CreatePaymentSession(c echo.Context) error {
	var req requests.CreatePaymentSessionRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Get payment credentials for the marina
	creds, err := h.queries.GetPaymentCredentials(c.Request().Context(), db.GetPaymentCredentialsParams{
		OrganizationID: req.OrganizationID,
		MarinaID:       req.MarinaID,
	})
	if err != nil {
		h.logger.Error("Failed to get payment credentials",
			zap.Error(err),
			zap.String("marina_id", req.MarinaID.String()),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get payment credentials")
	}

	// Generate a unique reference number
	reference := uuid.New().String()

	// Convert amount to Numeric
	var amount pgtype.Numeric
	err = amount.Scan(req.Amount)
	if err != nil {
		h.logger.Error("Failed to convert amount to Numeric",
			zap.Error(err),
			zap.Float64("amount", req.Amount),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process payment amount")
	}

	// Create payment record
	payment, err := h.queries.CreatePayment(c.Request().Context(), db.CreatePaymentParams{
		OrganizationID: req.OrganizationID,
		MarinaID:       req.MarinaID,
		CustomerID:     req.CustomerID,
		Amount:         amount,
		Currency:       req.Currency,
		Status:         string(adyen.PaymentStatusPending),
		PaymentType:    req.PaymentType,
		Reference:      reference,
		Description:    &req.Description,
	})
	if err != nil {
		h.logger.Error("Failed to create payment record",
			zap.Error(err),
			zap.Any("request", req),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create payment record")
	}

	// Create payment session
	session, err := h.adyenService.CreateCheckoutSession(
		c.Request().Context(),
		req.Amount,
		req.Currency,
		req.ReturnURL,
		reference,
		req.CountryCode,
		req.ShopperLocale,
		req.CustomerID.String(),
		creds.StoreID,
	)
	if err != nil {
		h.logger.Error("Failed to create Adyen payment session",
			zap.Error(err),
			zap.Any("request", req),
		)

		errMsg := err.Error()
		// Update payment status to failed
		_, err = h.queries.UpdatePaymentStatus(c.Request().Context(), db.UpdatePaymentStatusParams{
			ID:           payment.ID,
			Status:       string(adyen.PaymentStatusFailed),
			ErrorMessage: &errMsg,
		})
		if err != nil {
			h.logger.Error("Failed to update payment status",
				zap.Error(err),
				zap.String("payment_id", payment.ID.String()),
			)
		}

		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create payment session")
	}

	// Update payment with Adyen session details
	_, err = h.queries.UpdatePaymentStatus(c.Request().Context(), db.UpdatePaymentStatusParams{
		ID:     payment.ID,
		Status: string(adyen.PaymentStatusPending),
	})
	if err != nil {
		h.logger.Error("Failed to update payment with session details",
			zap.Error(err),
			zap.String("payment_id", payment.ID.String()),
		)
	}

	return c.JSON(http.StatusOK, &responses.PaymentSessionResponse{
		ID:          session.Id,
		SessionData: *session.SessionData,
		ExpiresAt:   session.ExpiresAt,
		Reference:   reference,
	})
}

// GetPaymentStatus godoc
// @Summary Get payment status
// @Description Gets the current status of a payment
// @Tags payments
// @Accept json
// @Produce json
// @Param id path string true "Payment ID"
// @Success 200 {object} responses.PaymentResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 404 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /payments/{id} [get]
func (h *PaymentHandler) GetPaymentStatus(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid payment ID")
	}

	payment, err := h.queries.GetPayment(c.Request().Context(), id)
	if err != nil {
		if err == pgx.ErrNoRows {
			return echo.NewHTTPError(http.StatusNotFound, "Payment not found")
		}
		h.logger.Error("Failed to get payment",
			zap.Error(err),
			zap.String("payment_id", id.String()),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to get payment")
	}

	// If payment is still pending and we have an Adyen PSP reference, check the status
	if payment.Status == string(adyen.PaymentStatusPending) && payment.AdyenPspReference != nil {
		status, err := h.adyenService.GetPaymentStatus(c.Request().Context(), *payment.AdyenPspReference)
		if err != nil {
			h.logger.Error("Failed to get payment status from Adyen",
				zap.Error(err),
				zap.String("payment_id", id.String()),
			)
		} else {
			// Update payment status based on Adyen response
			newStatus := string(adyen.PaymentStatusPending)
			var pspRef *string
			if status.ResultCode != nil {
				switch *status.ResultCode {
				case adyen.PaymentResultAuthorised:
					newStatus = string(adyen.PaymentStatusAuthorized)
					pspRef = status.PspReference
				case adyen.PaymentResultCaptured:
					newStatus = string(adyen.PaymentStatusCaptured)
					pspRef = status.PspReference
				case adyen.PaymentResultRefused, adyen.PaymentResultError:
					newStatus = string(adyen.PaymentStatusFailed)
					pspRef = status.PspReference
				case adyen.PaymentResultCancelled:
					newStatus = string(adyen.PaymentStatusCancelled)
					pspRef = status.PspReference
				}
			}

			if newStatus != payment.Status {
				payment, err = h.queries.UpdatePaymentStatus(c.Request().Context(), db.UpdatePaymentStatusParams{
					ID:                id,
					Status:            newStatus,
					AdyenPspReference: pspRef,
				})
				if err != nil {
					h.logger.Error("Failed to update payment status",
						zap.Error(err),
						zap.String("payment_id", id.String()),
					)
				}
			}
		}
	}

	// Convert amount to float64
	var amount float64
	err = payment.Amount.Scan(&amount)
	if err != nil {
		h.logger.Error("Failed to convert amount to float64",
			zap.Error(err),
			zap.String("payment_id", id.String()),
		)
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to process payment amount")
	}

	resp := responses.PaymentResponse{
		ID:             payment.ID,
		OrganizationID: payment.OrganizationID,
		MarinaID:       payment.MarinaID,
		CustomerID:     payment.CustomerID,
		Amount:         amount,
		Currency:       payment.Currency,
		Status:         payment.Status,
		PaymentType:    payment.PaymentType,
		Reference:      payment.Reference,
		CreatedAt:      payment.CreatedAt.Time,
		UpdatedAt:      payment.UpdatedAt.Time,
	}

	if payment.PaymentMethod != nil {
		resp.PaymentMethod = *payment.PaymentMethod
	}
	if payment.Description != nil {
		resp.Description = *payment.Description
	}
	if payment.AdyenPaymentID != nil {
		resp.AdyenPaymentID = *payment.AdyenPaymentID
	}
	if payment.AdyenMerchantReference != nil {
		resp.AdyenMerchantReference = *payment.AdyenMerchantReference
	}
	if payment.AdyenPspReference != nil {
		resp.AdyenPspReference = *payment.AdyenPspReference
	}

	return c.JSON(http.StatusOK, resp)
}

// HandleWebhook godoc
// @Summary Handle Adyen webhook
// @Description Handles payment status updates from Adyen
// @Tags payments
// @Accept json
// @Produce json
// @Param request body requests.PaymentWebhookRequest true "Webhook notification"
// @Success 200 {object} responses.WebhookResponse
// @Failure 400 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /payments/webhook [post]
func (h *PaymentHandler) HandleWebhook(c echo.Context) error {
	var req requests.PaymentWebhookRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	if err := c.Validate(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}

	// Process each notification item
	for _, item := range req.NotificationItems {
		notification := item.NotificationRequestItem

		// Create payment event
		eventData, err := json.Marshal(notification)
		if err != nil {
			h.logger.Error("Failed to marshal notification data",
				zap.Error(err),
				zap.Any("notification", notification),
			)
			continue
		}

		// Find payment by merchant reference
		payment, err := h.queries.GetPaymentByReference(c.Request().Context(), notification.MerchantReference)
		if err != nil {
			h.logger.Error("Failed to find payment",
				zap.Error(err),
				zap.String("reference", notification.MerchantReference),
			)
			continue
		}

		// Create payment event
		_, err = h.queries.CreatePaymentEvent(c.Request().Context(), db.CreatePaymentEventParams{
			PaymentID: payment.ID,
			EventType: notification.EventCode,
			EventData: eventData,
		})
		if err != nil {
			h.logger.Error("Failed to create payment event",
				zap.Error(err),
				zap.String("payment_id", payment.ID.String()),
			)
			continue
		}

		// Update payment status based on event code
		newStatus := payment.Status
		switch notification.EventCode {
		case adyen.EventAuthorisation:
			if notification.Success {
				newStatus = string(adyen.PaymentStatusAuthorized)
			} else {
				newStatus = string(adyen.PaymentStatusFailed)
			}
		case adyen.EventCapture:
			if notification.Success {
				newStatus = string(adyen.PaymentStatusCaptured)
			}
		case adyen.EventCancellation:
			if notification.Success {
				newStatus = string(adyen.PaymentStatusCancelled)
			}
		case adyen.EventRefund:
			if notification.Success {
				newStatus = string(adyen.PaymentStatusRefunded)
			}
		}

		if newStatus != payment.Status {
			pspRef := notification.PSPReference
			_, err = h.queries.UpdatePaymentStatus(c.Request().Context(), db.UpdatePaymentStatusParams{
				ID:                payment.ID,
				Status:            newStatus,
				AdyenPspReference: &pspRef,
			})
			if err != nil {
				h.logger.Error("Failed to update payment status",
					zap.Error(err),
					zap.String("payment_id", payment.ID.String()),
				)
			}
		}
	}

	// Return [accepted] to acknowledge receipt
	return c.JSON(http.StatusOK, responses.WebhookResponse{
		NotificationResponse: "[accepted]",
	})
}
