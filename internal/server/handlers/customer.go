package handlers

import (
	"errors"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/dockworks/dm-web-backend/pkg/utils"
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
// @Description Retrieves a paginated list of customers (minimal fields)
// @Tags Customers
// @Accept json
// @Produce json
// @Param page query int true "Page number" minimum(1)
// @Param pageSize query int true "Page size" minimum(1) maximum(100)
// @Success 200 {object} responses.CustomerListMinimalResponse
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

	dmeResponse, err := h.server.DME.CustomersList(ctx, req.Page, req.PageSize, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list customers",
			zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("pageSize", req.PageSize))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertCustomerListMinimal(dmeResponse)
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
	var req requests.CustomerUpdateRequest
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

	// Get existing customer first
	existingCustomer, err := h.server.DME.CustomerRetrieve(ctx, req.ID, orgID, *systemID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve existing customer: "+err.Error()).JSON(c)
	}

	// Only update fields that are present in the request
	if req.Name != "" {
		existingCustomer.Name = req.Name
	}
	if req.FirstName != "" {
		existingCustomer.FirstName = req.FirstName
	}
	if req.LastName != "" {
		existingCustomer.LastName = req.LastName
	}
	if req.Email != "" {
		existingCustomer.Email = req.Email
	}
	if req.Address1 != "" {
		existingCustomer.Address1 = req.Address1
	}
	if req.Address2 != "" {
		existingCustomer.Address2 = req.Address2
	}
	if req.Address3 != "" {
		existingCustomer.Address3 = req.Address3
	}
	if req.City != "" {
		existingCustomer.City = req.City
	}
	if req.State != "" {
		existingCustomer.State = req.State
	}
	if req.Zip != "" {
		existingCustomer.Zip = req.Zip
	}
	if req.Country != "" {
		existingCustomer.Country = req.Country
	}
	if req.Phone != "" {
		existingCustomer.Phone = req.Phone
	}
	if req.AltFirstName != "" {
		existingCustomer.AltFirstName = req.AltFirstName
	}
	if req.AltLastName != "" {
		existingCustomer.AltLastName = req.AltLastName
	}
	if req.AltAddress1 != "" {
		existingCustomer.AltAddress1 = req.AltAddress1
	}
	if req.AltAddress2 != "" {
		existingCustomer.AltAddress2 = req.AltAddress2
	}
	if req.AltAddress3 != "" {
		existingCustomer.AltAddress3 = req.AltAddress3
	}
	if req.AltCity != "" {
		existingCustomer.AltCity = req.AltCity
	}
	if req.AltState != "" {
		existingCustomer.AltState = req.AltState
	}
	if req.AltZip != "" {
		existingCustomer.AltZip = req.AltZip
	}
	if req.AltCountry != "" {
		existingCustomer.AltCountry = req.AltCountry
	}
	if req.AltPhone != "" {
		existingCustomer.AltPhone = req.AltPhone
	}
	existingCustomer.UseAltAddress = req.UseAltAddress
	if req.WorkPhone != "" {
		existingCustomer.WorkPhone = req.WorkPhone
	}
	if req.CellPhone != "" {
		existingCustomer.CellPhone = req.CellPhone
	}
	if req.EmergencyContact != "" {
		existingCustomer.EmergencyContact = req.EmergencyContact
	}
	if req.EmergencyPhone != "" {
		existingCustomer.EmergencyPhone = req.EmergencyPhone
	}
	if req.CompanyName != "" {
		existingCustomer.CompanyName = req.CompanyName
	}
	if req.ShipmentMethod != "" {
		existingCustomer.ShipmentMethod = req.ShipmentMethod
	}
	if req.ShipmentMethodDescription != "" {
		existingCustomer.ShipmentMethodDescription = req.ShipmentMethodDescription
	}
	if req.CustomInformation != nil {
		existingCustomer.CustomInformation = req.CustomInformation
	}
	if req.Attachments != nil {
		existingCustomer.Attachments = req.Attachments
	}

	// Convert to CustomerUpdate
	customer := dme.CustomerUpdate{
		ID:                        existingCustomer.ID,
		Name:                      existingCustomer.Name,
		FirstName:                 existingCustomer.FirstName,
		LastName:                  existingCustomer.LastName,
		Email:                     existingCustomer.Email,
		Address1:                  existingCustomer.Address1,
		Address2:                  existingCustomer.Address2,
		Address3:                  existingCustomer.Address3,
		City:                      existingCustomer.City,
		State:                     existingCustomer.State,
		Zip:                       existingCustomer.Zip,
		Country:                   existingCustomer.Country,
		Phone:                     existingCustomer.Phone,
		AltFirstName:              existingCustomer.AltFirstName,
		AltLastName:               existingCustomer.AltLastName,
		AltAddress1:               existingCustomer.AltAddress1,
		AltAddress2:               existingCustomer.AltAddress2,
		AltAddress3:               existingCustomer.AltAddress3,
		AltCity:                   existingCustomer.AltCity,
		AltState:                  existingCustomer.AltState,
		AltZip:                    existingCustomer.AltZip,
		AltCountry:                existingCustomer.AltCountry,
		AltPhone:                  existingCustomer.AltPhone,
		UseAltAddress:             existingCustomer.UseAltAddress,
		WorkPhone:                 existingCustomer.WorkPhone,
		CellPhone:                 existingCustomer.CellPhone,
		EmergencyContact:          existingCustomer.EmergencyContact,
		EmergencyPhone:            existingCustomer.EmergencyPhone,
		CompanyName:               existingCustomer.CompanyName,
		ShipmentMethod:            existingCustomer.ShipmentMethod,
		ShipmentMethodDescription: existingCustomer.ShipmentMethodDescription,
		CustomInformation:         existingCustomer.CustomInformation,
		Attachments:               existingCustomer.Attachments,
	}

	dmeResponse, err := h.server.DME.CustomerUpdate(ctx, &customer, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to update customer",
			zap.Error(err),
			zap.String("customerId", req.ID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

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

	// Convert request to dme.CustomerCreate
	customer := &dme.CustomerCreate{
		Name:                      req.Name,
		FirstName:                 req.FirstName,
		LastName:                  req.LastName,
		Email:                     req.Email,
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
		CustomInformation:         req.CustomInformation,
		Attachments:               req.Attachments,
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

	queries := h.server.DB.Queries()
	customerSettings, err := queries.GetCustomerSettings(ctx, db.GetCustomerSettingsParams{
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
	})
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to get customer settings",
			zap.Error(err),
			zap.String("customerId", req.CustomerID),
			zap.String("marinaId", req.MarinaID.String()))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get customer settings").JSON(c)
	}
	enablePortal := customerSettings.EnablePortal
	if req.EnablePortal != nil {
		enablePortal = req.EnablePortal
	}

	// Use UpsertCustomerSettings to create or update
	settings, err := queries.UpsertCustomerSettings(ctx, db.UpsertCustomerSettingsParams{
		MarinaID:     customerSettings.MarinaID,
		CustomerID:   customerSettings.CustomerID,
		EnablePortal: enablePortal,
	})
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to upsert customer settings",
			zap.Error(err),
			zap.String("customerId", req.CustomerID),
			zap.String("marinaId", req.MarinaID.String()))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to upsert customer settings").JSON(c)
	}
	customerUsers, err := queries.GetMarinaCustomerUsersByCustomerID(ctx, db.GetMarinaCustomerUsersByCustomerIDParams{
		MarinaID:   settings.MarinaID,
		CustomerID: &settings.CustomerID,
	})
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to get customer users",
			zap.Error(err),
			zap.String("customerId", req.CustomerID),
			zap.String("marinaId", req.MarinaID.String()))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get customer users").JSON(c)
	}
	if len(customerUsers) > 0 {
		activateUser := settings.EnablePortal
		for _, customerUser := range customerUsers {
			if *activateUser {
				_, err := h.server.DB.Queries().ActivateUser(ctx, customerUser.ID)
				if err != nil {
					h.server.Logger.DesugarZap.Error("Failed to activate customer user",
						zap.Error(err),
						zap.String("customerId", req.CustomerID),
						zap.String("marinaId", req.MarinaID.String()))
					return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to activate customer user").JSON(c)
				}
			} else {
				_, err := h.server.DB.Queries().DeactivateUser(ctx, customerUser.ID)
				if err != nil {
					h.server.Logger.DesugarZap.Error("Failed to deactivate customer user",
						zap.Error(err),
						zap.String("customerId", req.CustomerID),
						zap.String("marinaId", req.MarinaID.String()))
					return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to deactivate customer user").JSON(c)
				}
			}
		}
	}

	// Convert settings to response format
	response := responses.ConvertCustomerSettings(settings)
	return c.JSON(http.StatusOK, response)
}

// @Summary Customer intake (public endpoint)
// @Description Creates a customer in DME and then creates a user for that customer - public endpoint
// @Tags Customers
// @Accept json
// @Produce json
// @Param Organization-ID header string true "Organization ID"
// @Param Marina-ID header string true "Marina ID"
// @Param customer body requests.CustomerIntakeRequest true "Customer and user information"
// @Success 201 {object} responses.UserResponseWrapper "Created user with customer"
// @Failure 400 {object} responses.Error "Bad request"
// @Failure 500 {object} responses.Error "Server error"
// @Router /customer-intake [post]
func (h *CustomerHandler) CustomerIntake(c echo.Context) error {
	ctx := c.Request().Context()

	// Parse and validate request
	var req requests.CustomerIntakeRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Parse UUIDs
	orgID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid Organization-ID format").JSON(c)
	}

	marinaID, err := uuid.Parse(req.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid Marina-ID format").JSON(c)
	}

	queries := h.server.DB.Queries()

	// Validate that organization and marina exist and are linked
	marina, err := queries.GetMarinaByID(ctx, marinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Marina not found").JSON(c)
	}

	organization, err := queries.GetOrganizationByID(ctx, orgID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Organization not found").JSON(c)
	}

	if marina.OrganizationID != organization.ID {
		return responses.NewErrorResponse(http.StatusBadRequest, "Marina and organization are not linked").JSON(c)
	}

	// Get system ID for DME operations
	if marina.SystemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Marina does not have a system ID configured").JSON(c)
	}
	systemID := *marina.SystemID

	name := req.FirstName + " " + req.LastName
	email := utils.LowerCase(req.Email)
	// Step 1: Create customer in DME
	customer := &dme.CustomerCreate{
		Name:             name,
		FirstName:        req.FirstName,
		LastName:         req.LastName,
		Email:            email,
		Address1:         req.Address1,
		Address2:         req.Address2,
		Address3:         req.Address3,
		City:             req.City,
		State:            req.State,
		Zip:              req.Zip,
		Country:          req.Country,
		Phone:            req.Phone,
		WorkPhone:        req.WorkPhone,
		CellPhone:        req.CellPhone,
		EmergencyContact: req.EmergencyContact,
		EmergencyPhone:   req.EmergencyPhone,
		CompanyName:      req.CompanyName,
	}

	dmeResponse, err := h.server.DME.CreateCustomer(ctx, customer, orgID, systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to create customer in DME",
			zap.Error(err),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to create customer: "+err.Error()).JSON(c)
	}

	// Step 2: Get customer_user role
	customerRole, err := queries.GetRoleByName(ctx, "customer_user")
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to get customer_user role",
			zap.Error(err),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Customer role not found").JSON(c)
	}

	// Step 3: Create user for the customer
	// Check if the email is already taken
	_, err = queries.GetUserByEmail(ctx, email)
	if err == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Email already taken").JSON(c)
	}

	// Generate username if not provided
	username := utils.GenerateUsername(req.FirstName)

	// Check if the username is already taken
	_, err = queries.GetUserByUsername(ctx, username)
	if err == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Username already taken").JSON(c)
	}

	// Set defaults for user creation
	isActive := true
	isSuperuser := false

	// Create customer user
	params := db.CreateCustomerUserParams{
		Username:            username,
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		Email:               email,
		Phone:               &req.Phone,
		OrganizationID:      orgID,
		MarinaID:            marinaID,
		RoleID:              customerRole.ID,
		CustomerID:          &dmeResponse.ID,
		IsCustomer:          utils.Pointer(true),
		IsSuperuser:         &isSuperuser,
		IsActive:            &isActive,
		EmailVerified:       utils.PgTimeNow(),
		LastLogin:           utils.PgTimeNow(),
		FailedLoginAttempts: utils.Pointer(int32(0)),
		LockedUntil:         utils.PgTimeNow(),
		LastPasswordReset:   utils.PgTimeNow(),
		UserAnalytics:       utils.Pointer(true),
	}

	user, err := queries.CreateCustomerUser(ctx, params)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to create customer user",
			zap.Error(err),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to create user: "+err.Error()).JSON(c)
	}

	// Assign user to marina
	assignUserToMarina := db.AssignUserToMarinaParams{
		UserID:     user.ID,
		MarinaID:   marinaID,
		CustomerID: &dmeResponse.ID,
	}

	err = queries.AssignUserToMarina(ctx, assignUserToMarina)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to assign user to marina",
			zap.Error(err),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to assign user to marina: "+err.Error()).JSON(c)
	}

	// Update customer settings to enable portal
	_, err = queries.UpsertCustomerSettings(ctx, db.UpsertCustomerSettingsParams{
		MarinaID:     marinaID,
		CustomerID:   dmeResponse.ID,
		EnablePortal: utils.Pointer(true),
	})
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to create customer settings",
			zap.Error(err),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to create customer settings: "+err.Error()).JSON(c)
	}

	// Hash the password using the utility function from utils
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to hash password",
			zap.Error(err),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to process password").JSON(c)
	}

	// Update the user with the new password
	userPasswordParams := db.UpdateUserInviteParams{
		ID:           user.ID,
		PasswordHash: utils.Pointer(passwordHash),
	}

	_, err = queries.UpdateUserInvite(ctx, userPasswordParams)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to update user password",
			zap.Error(err),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to update user password").JSON(c)
	}

	response := responses.NewUserResponseSuccess(user)
	return c.JSON(http.StatusCreated, response)
}
