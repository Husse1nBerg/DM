package handlers

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
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
// @Description Retrieves invoices for a specific customer
// @Tags Invoices
// @Accept json
// @Produce json
// @Param customerId query string true "Customer ID"
// @Param invoiceDate query string false "Invoice date (format: YYYY-MM-DD)"
// @Success 200 {object} responses.InvoiceListResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
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

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
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

// @Summary Get invoices by IDs
// @Description Retrieves invoices by their IDs
// @Tags Invoices
// @Accept json
// @Produce json
// @Param invoiceIds body requests.GetInvoicesByIDsRequest true "Invoice IDs"
// @Success 200 {object} responses.InvoiceListResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /invoices/retrieve [post]
// func (h *InvoiceHandler) GetInvoicesByIDs(c echo.Context) error {
// 	ctx := c.Request().Context()
// 	var req requests.GetInvoicesByIDsRequest
// 	if err := c.Bind(&req); err != nil {
// 		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
// 	}

// 	if err := c.Validate(&req); err != nil {
// 		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
// 	}

// 	userToken := c.Get("user").(*jwt.Token)
// 	claims := userToken.Claims.(*token.JwtCustomClaims)
// 	userID := claims.ID
// 	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
// 	if err != nil {
// 		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
// 	}

// 	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
// 	if err != nil {
// 		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
// 	}

// 	orgID := marina.OrganizationID
// 	systemID := marina.SystemID

// 	// Check if systemID is nil before dereferencing
// 	if systemID == nil {
// 		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
// 	}

// 	dmeResponse, err := h.server.DME.RetrieveInvoices(ctx, req.InvoiceIDs, orgID, *systemID)
// 	if err != nil {
// 		h.server.Logger.DesugarZap.Error("Failed to retrieve invoices",
// 			zap.Error(err),
// 			zap.Strings("invoiceIds", req.InvoiceIDs))
// 		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
// 	}

// 	// Convert DME response to API response
// 	response := responses.ConvertInvoiceList(dmeResponse)
// 	return c.JSON(http.StatusOK, response)
// }

// @Summary Initiate payment for invoice
// @Description Initiates a payment process for a specific invoice
// @Tags Invoices
// @Accept json
// @Produce json
// @Param payment body requests.InitiatePaymentRequest true "Payment information"
// @Success 200 {object} responses.PaymentInitiationResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /invoices/pay [post]
func (h *InvoiceHandler) InitiatePayment(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.InitiatePaymentRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	dmeResponse, err := h.server.DME.InitiatePayment(ctx, req.CustomerID, req.InvoiceID, req.Amount, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to initiate payment",
			zap.Error(err),
			zap.String("customerId", req.CustomerID),
			zap.String("invoiceId", req.InvoiceID),
			zap.Float64("amount", req.Amount))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertPaymentInitiation(dmeResponse)
	return c.JSON(http.StatusOK, response)
}
