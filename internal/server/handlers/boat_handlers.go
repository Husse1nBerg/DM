package handlers

import (
	"encoding/json"
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

// BoatHandler handles boat-related requests
type BoatHandler struct {
	server *s.Server
}

// NewBoatHandler creates a new boat handler
func NewBoatHandler(server *s.Server) *BoatHandler {
	return &BoatHandler{
		server: server,
	}
}

// @Summary List boats by page
// @Description Retrieves a paginated list of boats
// @Tags Boats
// @Accept json
// @Produce json
// @Param page query int true "Page number" minimum(1)
// @Param pageSize query int true "Page size" minimum(1) maximum(100)
// @Success 200 {object} responses.BoatListResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/list [get]
func (h *BoatHandler) ListBoatsByPage(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.BoatsListRequest)
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

	dmeResponse, err := h.server.DME.BoatsList(ctx, req.Page, req.PageSize, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list boats",
			zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("pageSize", req.PageSize))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Get total count estimate
	totalPages := 1
	if len(dmeResponse) == req.PageSize {
		// If we got a full page, assume there's at least one more page
		totalPages = req.Page + 1
	} else if len(dmeResponse) == 0 && req.Page > 1 {
		// If current page is empty but we're past page 1, use previous page as max
		totalPages = req.Page - 1
	}

	// Convert response to API response format
	response := responses.ConvertBoatList(dmeResponse, req.Page, req.PageSize, totalPages)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve boat by ID
// @Description Retrieves a boat by its ID
// @Tags Boats
// @Accept json
// @Produce json
// @Param BoatId query string true "Boat ID"
// @Success 200 {object} responses.BoatResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/retrieve [get]
func (h *BoatHandler) RetrieveBoat(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.BoatRetrieveRequest)
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

	dmeResponse, err := h.server.DME.RetrieveBoatByID(ctx, req.BoatID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve boat",
			zap.Error(err),
			zap.String("boatId", req.BoatID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertBoat(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve boats for customer
// @Description Retrieves boats associated with a customer
// @Tags Boats
// @Accept json
// @Produce json
// @Param CustomerId query string true "Customer ID"
// @Success 200 {object} responses.BoatSearchResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/customer [get]
func (h *BoatHandler) RetrieveBoatsForCustomer(c echo.Context) error {
	ctx := c.Request().Context()
	customerID := c.QueryParam("CustomerId")
	if customerID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Customer ID is required").JSON(c)
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

	// Get boats for the customer
	boats, err := h.server.DME.RetrieveBoatsForCustomer(ctx, customerID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve boats for customer",
			zap.Error(err),
			zap.String("customerId", customerID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Create a response with the list of boats
	response := responses.ConvertBoatList(boats, 1, 1, len(boats))
	return c.JSON(http.StatusOK, response)
}

// @Summary Search boats
// @Description Searches for boats based on search string
// @Tags Boats
// @Accept json
// @Produce json
// @Param SearchString query string true "Search string"
// @Param DirectHit query bool false "Direct hit search" default(false)
// @Success 200 {object} responses.BoatSearchResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/search [get]
func (h *BoatHandler) SearchBoats(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.BoatSearchRequest)
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

	dmeResponse, err := h.server.DME.SearchBoats(ctx, req.SearchString, req.DirectHit, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to search boats",
			zap.Error(err),
			zap.String("searchString", req.SearchString),
			zap.Bool("directHit", req.DirectHit))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertBoatSearch(&dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Update boat
// @Description Updates a boat's information
// @Tags Boats
// @Accept json
// @Produce json
// @Param boat body dme.BoatUpdate true "Boat information"
// @Success 200 {object} responses.BoatResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/update [post]
func (h *BoatHandler) UpdateBoat(c echo.Context) error {
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
	var reqStruct dme.BoatUpdate
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

	dmeResponse, err := h.server.DME.UpdateBoat(ctx, &reqStruct, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to update boat",
			zap.Error(err),
			zap.String("boatId", id))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertBoat(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Create boat
// @Description Creates a new boat
// @Tags Boats
// @Accept json
// @Produce json
// @Param boat body requests.BoatCreateRequest true "Boat information"
// @Success 200 {object} responses.BoatResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/create [post]
func (h *BoatHandler) CreateBoat(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.BoatCreateRequest
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
			return responses.NewErrorResponse(http.StatusBadRequest, "'id' field must not be provided when creating a boat").JSON(c)
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

	// Convert request to dme.BoatCreate with all fields properly mapped
	boat := &dme.BoatCreate{
		Name:                req.Name,
		OwnerID:             req.OwnerID,
		Registration:        req.Registration,
		Year:                req.Year,
		Make:                req.Make,
		Model:               req.Model,
		HIN:                 req.Hin,
		LOA:                 req.LOA,
		LWL:                 req.LWL,
		Draft:               req.Draft,
		Beam:                req.Beam,
		Height:              req.Height,
		Color:               req.Color,
		TrailerMake:         req.TrailerMake,
		TrailerModel:        req.TrailerModel,
		TrailerSerial:       req.TrailerSerial,
		TrailerRegistration: req.TrailerRegistration,
		TrailerLocation:     req.TrailerLocation,
		SummerSlip:          req.SummerSlip,
		WinterSlip:          req.WinterSlip,
		InsuranceCompany:    req.InsuranceCompany,
		InsuranceExpDate:    req.InsuranceExpDate,
		SlipID:              req.SlipID,
	}

	dmeResponse, err := h.server.DME.CreateBoat(ctx, boat, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to create boat",
			zap.Error(err),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertBoat(dmeResponse)
	return c.JSON(http.StatusOK, response)
}
