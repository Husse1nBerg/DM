package handlers

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/token"
)

// CustomerHandler handles customer-related requests
type CustomerHandler struct {
	server *s.Server
}

// NewCustomerHandler creates a new customer handler
func NewCustomerHandler(server *s.Server) *CustomerHandler {
	return &CustomerHandler{
		server: server,
	}
}

// @Summary List customers by page
// @Description Retrieves a paginated list of customers
// @Tags Customers
// @Accept json
// @Produce json
// @Param page query int true "Page number" minimum(1)
// @Param pageSize query int true "Page size" minimum(1) maximum(100)
// @Success 200 {object} responses.CustomerListResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /customers/list [get]
func (h *CustomerHandler) ListCustomersByPage(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.CustomerListRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request parameters: "+err.Error())
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	marinaIDStr := claims.MarinaId
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	dmeResponse, err := h.server.DME.CustomersList(ctx, req.Page, req.PageSize, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list customers",
			zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("pageSize", req.PageSize))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list customers: "+err.Error())
	}

	// Convert DME response to API response
	response := responses.ConvertCustomerList(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve customer by ID
// @Description Retrieves a customer by their ID
// @Tags Customers
// @Accept json
// @Produce json
// @Param CustomerId query string true "Customer ID"
// @Success 200 {object} responses.CustomerResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /customers/retrieve [get]
func (h *CustomerHandler) RetrieveCustomer(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.CustomerRetrieveRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request parameters: "+err.Error())
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	marinaIDStr := claims.MarinaId
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	dmeResponse, err := h.server.DME.CustomerRetrieve(ctx, req.CustomerID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve customer",
			zap.Error(err),
			zap.String("customerId", req.CustomerID))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to retrieve customer: "+err.Error())
	}

	// Convert DME response to API response
	response := responses.ConvertCustomer(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Search customers
// @Description Searches for customers based on search string
// @Tags Customers
// @Accept json
// @Produce json
// @Param SearchString query string true "Search string"
// @Param DirectHit query bool false "Direct hit search" default(false)
// @Success 200 {object} responses.CustomerSearchResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /customers/search [get]
func (h *CustomerHandler) SearchCustomers(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.CustomerSearchRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request parameters: "+err.Error())
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	marinaIDStr := claims.MarinaId
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	dmeResponse, err := h.server.DME.CustomerSearch(ctx, req.SearchString, req.DirectHit, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to search customers",
			zap.Error(err),
			zap.String("searchString", req.SearchString),
			zap.Bool("directHit", req.DirectHit))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to search customers: "+err.Error())
	}

	// Convert DME response to API response
	response := responses.ConvertCustomerSearch(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Update customer
// @Description Updates a customer's information
// @Tags Customers
// @Accept json
// @Produce json
// @Param customer body dme.CustomerUpdate true "Customer information"
// @Success 200 {object} responses.CustomerResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /customers/update [post]
func (h *CustomerHandler) UpdateCustomer(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(dme.CustomerUpdate)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body: "+err.Error())
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	marinaIDStr := claims.MarinaId
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Convert to DME request
	dmeReq := &dme.CustomerUpdate{
		ID:                        req.ID,
		Name:                      req.Name,
		FirstName:                 req.FirstName,
		LastName:                  req.LastName,
		Address1:                  req.Address1,
		Address2:                  req.Address2,
		Address3:                  req.Address3,
		City:                      req.City,
		State:                     req.State,
		Zip:                       req.Zip,
		Country:                   req.Country,
		Phone:                     req.Phone,
		CompanyName:               req.CompanyName,
		ShipmentMethod:            req.ShipmentMethod,
		ShipmentMethodDescription: req.ShipmentMethodDescription,
	}

	dmeResponse, err := h.server.DME.CustomerUpdate(ctx, dmeReq, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to update customer",
			zap.Error(err),
			zap.String("customerId", req.ID))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to update customer: "+err.Error())
	}

	// Convert DME response to API response
	response := responses.ConvertCustomer(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary List customers short
// @Description Retrieves a paginated short list of customers
// @Tags Customers
// @Accept json
// @Produce json
// @Param page query int true "Page number" minimum(1)
// @Param pageSize query int true "Page size" minimum(1) maximum(100)
// @Success 200 {object} responses.CustomerListShortResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /customers/list-short [get]
func (h *CustomerHandler) ListCustomersShortByPage(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.CustomerListRequest)
	if err := c.Bind(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request parameters: "+err.Error())
	}

	if err := c.Validate(req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Validation failed: "+err.Error())
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	marinaIDStr := claims.MarinaId
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	dmeResponse, err := h.server.DME.CustomersListShort(ctx, req.Page, req.PageSize, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list customers short",
			zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("pageSize", req.PageSize))
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to list customers short: "+err.Error())
	}

	// Convert DME response to API response
	response := responses.ConvertCustomerListShort(dmeResponse)
	return c.JSON(http.StatusOK, response)
}
