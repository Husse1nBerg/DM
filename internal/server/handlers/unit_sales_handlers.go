package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// UnitSalesHandler handles operations related to Unit Sales
type UnitSalesHandler struct {
	server *s.Server
}

// NewUnitSalesHandler creates a new UnitSales handler instance
func NewUnitSalesHandler(server *s.Server) *UnitSalesHandler {
	return &UnitSalesHandler{server: server}
}

// RetrieveCustomerContractsHandler godoc
// @Summary Retrieve customer contracts
// @Description Retrieves contracts for a specific customer from DME UnitSales API
// @Tags UnitSales
// @Accept json
// @Produce json
// @Param CustomerId query string true "Customer ID"
// @Success 200 {object} responses.CustomerContractsResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /api/v1/unit-sales/customer-contracts [get]
func (h *UnitSalesHandler) RetrieveCustomerContractsHandler(c echo.Context) error {
	ctx := c.Request().Context()

	// Bind and validate request
	req := new(requests.CustomerContractsRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Get user and resolve org and system IDs
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID

	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	// Call DME API
	contracts, err := h.server.DME.RetrieveCustomerContracts(ctx, req.CustomerID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve customer contracts",
			zap.Error(err),
			zap.String("customerId", req.CustomerID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewSuccessResponse(responses.NewCustomerContractsResponse(contracts)).JSON(c)
}

// RetrieveCustomerQuotesHandler godoc
// @Summary Retrieve customer quotes
// @Description Retrieves quotes for a specific customer from DME UnitSales API
// @Tags UnitSales
// @Accept json
// @Produce json
// @Param CustomerId query string true "Customer ID"
// @Success 200 {object} responses.CustomerQuotesResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /api/v1/unit-sales/customer-quotes [get]
func (h *UnitSalesHandler) RetrieveCustomerQuotesHandler(c echo.Context) error {
	ctx := c.Request().Context()

	// Bind and validate request
	req := new(requests.CustomerContractsRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Get user and resolve org and system IDs
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID

	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	// Call DME API
	quotes, err := h.server.DME.RetrieveCustomerQuotes(ctx, req.CustomerID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve customer quotes",
			zap.Error(err),
			zap.String("customerId", req.CustomerID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewSuccessResponse(responses.NewCustomerQuotesResponse(quotes)).JSON(c)
}
