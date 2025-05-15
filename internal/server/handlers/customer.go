package handlers

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/dockworks/dm-web-backend/internal/db"
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
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
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
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
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
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
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
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
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
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
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
			zap.String("directHit", req.DirectHit))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
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
	var reqMap map[string]interface{}
	if err := c.Bind(&reqMap); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Validate required field
	id, ok := reqMap["id"].(string)
	if !ok || id == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Missing or invalid 'id' field").JSON(c)
	}

	// Type/value validation: marshal to JSON, unmarshal into struct, validate
	jsonBytes, err := json.Marshal(reqMap)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Failed to marshal request").JSON(c)
	}
	var reqStruct dme.CustomerUpdate
	if err := json.Unmarshal(jsonBytes, &reqStruct); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Failed to parse request: "+err.Error()).JSON(c)
	}
	if err := c.Validate(&reqStruct); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
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

	dmeResponse, err := h.server.DME.CustomerUpdate(ctx, reqMap, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to update customer",
			zap.Error(err),
			zap.String("customerId", id))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
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
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
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
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertCustomerListShort(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Create customer
// @Description Creates a new customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param customer body requests.CustomerCreateRequest true "Customer information"
// @Success 200 {object} responses.CustomerResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /customers/create [post]
func (h *CustomerHandler) CreateCustomer(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.CustomerCreateRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Ensure no id is provided in the payload (defensive, in case client sends extra fields)
	var raw map[string]interface{}
	if err := c.Bind(&raw); err == nil {
		if _, hasID := raw["id"]; hasID {
			return responses.NewErrorResponse(http.StatusBadRequest, "'id' field must not be provided when creating a customer").JSON(c)
		}
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

	// Convert request to dme.CustomerCreate
	customer := &dme.CustomerCreate{
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
		AltFirstName:              req.AltFirstName,
		AltLastName:               req.AltLastName,
		AltAddress1:               req.AltAddress1,
		AltAddress2:               req.AltAddress2,
		AltAddress3:               req.AltAddress3,
		AltCity:                   req.AltCity,
		AltState:                  req.AltState,
		AltZip:                    req.AltZip,
		AltCountry:                req.AltCountry,
		AltPhone:                  req.AltPhone,
		UseAltAddress:             req.UseAltAddress,
		WorkPhone:                 req.WorkPhone,
		CellPhone:                 req.CellPhone,
		EmergencyContact:          req.EmergencyContact,
		EmergencyPhone:            req.EmergencyPhone,
		CompanyName:               req.CompanyName,
		ShipmentMethod:            req.ShipmentMethod,
		ShipmentMethodDescription: req.ShipmentMethodDescription,
	}

	dmeResponse, err := h.server.DME.CreateCustomer(ctx, customer, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to create customer",
			zap.Error(err),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertCustomer(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Get customer settings
// @Description Retrieves settings for a customer, creates with defaults if not found
// @Tags Customers
// @Accept json
// @Produce json
// @Param CustomerId query string true "Customer ID"
// @Success 200 {object} responses.CustomerSettingsResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /customers/settings [get]
func (h *CustomerHandler) GetCustomerSettings(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.CustomerSettingsRetrieveRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Try to get existing settings
	settings, err := h.server.DB.Queries().GetCustomerSettings(ctx, db.GetCustomerSettingsParams{
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
	})

	// If not found, create with default values
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			// Default value for enable_portal is false
			defaultEnablePortal := false
			settings, err = h.server.DB.Queries().CreateCustomerSettings(ctx, db.CreateCustomerSettingsParams{
				MarinaID:     req.MarinaID,
				CustomerID:   req.CustomerID,
				EnablePortal: &defaultEnablePortal,
			})
			if err != nil {
				h.server.Logger.DesugarZap.Error("Failed to create customer settings",
					zap.Error(err),
					zap.String("customerId", req.CustomerID),
					zap.String("marinaId", req.MarinaID.String()))
				return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to create customer settings").JSON(c)
			}
		} else {
			h.server.Logger.DesugarZap.Error("Failed to get customer settings",
				zap.Error(err),
				zap.String("customerId", req.CustomerID),
				zap.String("marinaId", req.MarinaID.String()))
			return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get customer settings").JSON(c)
		}
	}

	// Convert settings to response format
	response := responses.ConvertCustomerSettings(settings)
	return c.JSON(http.StatusOK, response)
}

// @Summary Update customer settings
// @Description Updates settings for a customer
// @Tags Customers
// @Accept json
// @Produce json
// @Param settings body requests.CustomerSettingsUpdateRequest true "Customer settings"
// @Success 200 {object} responses.CustomerSettingsResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /customers/settings [post]
func (h *CustomerHandler) UpdateCustomerSettings(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.CustomerSettingsUpdateRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Use UpsertCustomerSettings to create or update
	settings, err := h.server.DB.Queries().UpsertCustomerSettings(ctx, db.UpsertCustomerSettingsParams{
		MarinaID:     req.MarinaID,
		CustomerID:   req.CustomerID,
		EnablePortal: req.EnablePortal,
	})

	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to update customer settings",
			zap.Error(err),
			zap.String("customerId", req.CustomerID),
			zap.String("marinaId", req.MarinaID.String()))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to update customer settings").JSON(c)
	}

	// Convert settings to response format
	response := responses.ConvertCustomerSettings(settings)
	return c.JSON(http.StatusOK, response)
}
