package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DMECredentialHandler handles operations related to DME credentials
type DMECredentialHandler struct {
	server *s.Server
}

// NewDMECredentialHandler creates a new DME credential handler instance
func NewDMECredentialHandler(server *s.Server) *DMECredentialHandler {
	return &DMECredentialHandler{server: server}
}

// CreateDMECredential godoc
// @Summary Create DME credentials
// @Description Create new DME API credentials for an organization
// @Tags DME Credentials
// @Accept json
// @Produce json
// @Param request body requests.CreateDMECredentialRequest true "Create DME credential request"
// @Success 200 {object} responses.DMECredentialResponseWrapper
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/credentials [post]
func (h *DMECredentialHandler) CreateDMECredential(c echo.Context) error {
	req := new(requests.CreateDMECredentialRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	// Setup parameters
	params := db.CreateDMECredentialsParams{
		OrganizationID: req.OrganizationID,
		Username:       req.Username,
		Password:       req.Password,
		IsOldApi:       req.IsOldAPI,
	}

	credential, err := queries.CreateDMECredentials(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	err = h.server.DME.InitializeCredentials(c.Request().Context(), req.OrganizationID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewDMECredentialResponseSuccess(credential).JSON(c)
}

// GetDMECredentialByOrgID godoc
// @Summary Get DME credentials by organization ID
// @Description Get DME API credentials for an organization
// @Tags DME Credentials
// @Accept json
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Success 200 {object} responses.DMECredentialResponseWrapper
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/credentials/organization/{organizationId} [get]
func (h *DMECredentialHandler) GetDMECredentialByOrgID(c echo.Context) error {
	orgIDStr := c.Param("organizationId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID").JSON(c)
	}

	queries := h.server.DB.Queries()
	credential, err := queries.GetDMECredentialsByOrgID(c.Request().Context(), orgID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, err).JSON(c)
	}

	return responses.NewDMECredentialResponseSuccess(credential).JSON(c)
}

// UpdateDMECredential godoc
// @Summary Update DME credentials
// @Description Update DME API credentials for an organization
// @Tags DME Credentials
// @Accept json
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Param request body requests.UpdateDMECredentialRequest true "Update DME credential request"
// @Success 200 {object} responses.DMECredentialResponseWrapper
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/credentials/organization/{organizationId} [put]
func (h *DMECredentialHandler) UpdateDMECredential(c echo.Context) error {
	orgIDStr := c.Param("organizationId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID").JSON(c)
	}

	req := new(requests.UpdateDMECredentialRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	// First, get the current credentials
	existingCredential, err := queries.GetDMECredentialsByOrgID(c.Request().Context(), orgID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "DME credentials not found").JSON(c)
	}

	// Setup parameters
	params := db.UpdateDMECredentialsParams{
		OrganizationID: orgID,
		Username:       existingCredential.Username,
		Password:       existingCredential.Password,
		IsOldApi:       existingCredential.IsOldApi,
	}

	if req.Username != nil {
		params.Username = *req.Username
	}
	if req.Password != nil {
		params.Password = req.Password
	}
	if req.IsOldAPI != nil {
		params.IsOldApi = req.IsOldAPI
	}

	credential, err := queries.UpdateDMECredentials(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewDMECredentialResponseSuccess(credential).JSON(c)
}

// DeleteDMECredential godoc
// @Summary Delete DME credentials
// @Description Soft delete DME API credentials for an organization
// @Tags DME Credentials
// @Accept json
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Success 204 "No Content"
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/credentials/organization/{organizationId} [delete]
func (h *DMECredentialHandler) DeleteDMECredential(c echo.Context) error {
	orgIDStr := c.Param("organizationId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID").JSON(c)
	}

	queries := h.server.DB.Queries()

	// First, check if the credentials exist
	_, err = queries.GetDMECredentialsByOrgIDWithDeleted(c.Request().Context(), orgID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "DME credentials not found").JSON(c)
	}

	err = queries.HardDeleteDMECredentials(c.Request().Context(), orgID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return c.NoContent(http.StatusNoContent)
}
