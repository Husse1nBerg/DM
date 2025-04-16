package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DMEHandler handles DME API operations
type DMEHandler struct {
	server *s.Server
}

// NewDMEHandler creates a new DME handler
func NewDMEHandler(server *s.Server) *DMEHandler {
	return &DMEHandler{server: server}
}

// CallDMEAPI godoc
// @Summary Call DME API
// @Description Makes a request to the DME API using stored organization credentials
// @Tags DME API
// @Accept json
// @Produce json
// @Param request body requests.DMEAPIRequest true "DME API request details"
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /dme/api [post]
func (h *DMEHandler) CallDMEAPI(c echo.Context) error {
	req := new(requests.DMEAPIRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Get organization ID
	organizationID, err := uuid.Parse(req.OrganizationID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID").JSON(c)
	}

	// Check if credentials exist
	_, err = h.server.DB.Queries().GetDMECredentialsByOrgID(c.Request().Context(), organizationID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "DME credentials not found for this organization").JSON(c)
	}

	// Authenticate if necessary (this will happen internally in the client)
	if err := h.server.DME.Authenticate(c.Request().Context(), organizationID); err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Failed to authenticate with DME API: "+err.Error()).JSON(c)
	}

	// Make the API request
	var result map[string]interface{}
	err = h.server.DME.DoJSONRequest(
		c.Request().Context(),
		req.Method,
		req.Endpoint,
		req.Body,
		&result,
		organizationID,
		req.SystemID,
	)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "DME API request failed: "+err.Error()).JSON(c)
	}

	return c.JSON(http.StatusOK, result)
}

// GetVessels godoc
// @Summary Get vessels
// @Description Get vessels from DME API
// @Tags DME API
// @Accept json
// @Produce json
// @Param organizationId query string true "Organization ID"
// @Param systemId query string true "System ID"
// @Success 200 {array} dme.VesselDetails
// @Failure 400 {object} responses.ErrorResponse
// @Failure 401 {object} responses.ErrorResponse
// @Failure 500 {object} responses.ErrorResponse
// @Router /dme/vessels [get]
func (h *DMEHandler) GetVessels(c echo.Context) error {
	orgIDStr := c.QueryParam("organizationId")
	if orgIDStr == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Organization ID is required").JSON(c)
	}

	systemID := c.QueryParam("systemId")
	if systemID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required").JSON(c)
	}

	// Parse organization ID
	organizationID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID").JSON(c)
	}

	// Check if credentials exist
	_, err = h.server.DB.Queries().GetDMECredentialsByOrgID(c.Request().Context(), organizationID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "DME credentials not found for this organization").JSON(c)
	}

	// Authenticate if necessary
	if err := h.server.DME.Authenticate(c.Request().Context(), organizationID); err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Failed to authenticate with DME API: "+err.Error()).JSON(c)
	}

	// Get vessels using the high-level method
	vessels, err := h.server.DME.GetVessels(c.Request().Context(), organizationID, systemID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get vessels: "+err.Error()).JSON(c)
	}

	return c.JSON(http.StatusOK, vessels)
}
