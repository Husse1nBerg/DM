package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// GeneralHandler handles operations related to DME General API endpoints
type GeneralHandler struct {
	server *s.Server
}

// NewGeneralHandler creates a new general handler instance
func NewGeneralHandler(server *s.Server) *GeneralHandler {
	return &GeneralHandler{server: server}
}

// ListClerksHandler godoc
// @Summary List system clerks (users)
// @Description Retrieves a list of system clerks (users) from the DME API
// @Tags General
// @Accept json
// @Produce json
// @Param includeInactive query bool false "Include inactive clerks"
// @Param systemId query string true "System ID"
// @Success 200 {object} responses.ClerkListResponse
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /api/v1/general/clerks/list [get]
func (h *GeneralHandler) ListClerksHandler(c echo.Context) error {
	// Get user from token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	req := new(requests.ClerkListRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Use organization ID from claims if not provided in request
	if req.OrganizationID.String() == "00000000-0000-0000-0000-000000000000" {
		req.OrganizationID = claims.OrgId
	}

	// Default includeInactive to false if not provided
	includeInactive := false
	if req.IncludeInactive != nil {
		includeInactive = *req.IncludeInactive
	}

	// Call DME API to list clerks
	clerks, err := h.server.DME.ListClerks(c.Request().Context(), req.OrganizationID, req.SystemID, includeInactive)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to list clerks from DME API", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve clerks").JSON(c)
	}

	// Convert to response format
	response := responses.NewClerkListResponse(clerks)

	return responses.NewSuccessResponse(response).JSON(c)
}

// RetrieveClerkHandler godoc
// @Summary Retrieve a specific clerk (user) record
// @Description Retrieves a specific clerk (user) record from the DME API
// @Tags General
// @Accept json
// @Produce json
// @Param clerkId query string true "Clerk ID"
// @Param systemId query string true "System ID"
// @Success 200 {object} responses.ClerkResponse
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /api/v1/general/clerks [get]
func (h *GeneralHandler) RetrieveClerkHandler(c echo.Context) error {
	// Get user from token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	req := new(requests.ClerkRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Use organization ID from claims if not provided in request
	if req.OrganizationID.String() == "00000000-0000-0000-0000-000000000000" {
		req.OrganizationID = claims.OrgId
	}

	// Call DME API to retrieve clerk
	clerk, err := h.server.DME.RetrieveClerk(c.Request().Context(), req.ClerkID, req.OrganizationID, req.SystemID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to retrieve clerk from DME API", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve clerk").JSON(c)
	}

	// Convert to response format
	response := responses.NewClerkResponse(*clerk)

	return responses.NewSuccessResponse(response).JSON(c)
}

// ListLocationsHandler godoc
// @Summary List business locations
// @Description Retrieves a list of business locations from the DME API
// @Tags General
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Success 200 {object} responses.LocationListResponse
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /api/v1/general/locations [get]
func (h *GeneralHandler) ListLocationsHandler(c echo.Context) error {
	// Get user from token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	req := new(requests.LocationListRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Use organization ID from claims if not provided in request
	if req.OrganizationID.String() == "00000000-0000-0000-0000-000000000000" {
		req.OrganizationID = claims.OrgId
	}

	// Call DME API to list locations
	locations, err := h.server.DME.ListLocations(c.Request().Context(), req.OrganizationID, req.SystemID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to list locations from DME API", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve locations").JSON(c)
	}

	// Convert to response format
	response := responses.NewLocationListResponse(locations)

	return responses.NewSuccessResponse(response).JSON(c)
}
