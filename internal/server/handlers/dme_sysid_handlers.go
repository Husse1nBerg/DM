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

// DMESysIDHandler handles operations related to DME System IDs
type DMESysIDHandler struct {
	server *s.Server
}

// NewDMESysIDHandler creates a new DME System ID handler instance
func NewDMESysIDHandler(server *s.Server) *DMESysIDHandler {
	return &DMESysIDHandler{server: server}
}

// CreateDMESysID godoc
// @Summary Create DME System ID
// @Description Create a new DME System ID for an organization
// @Tags DME System IDs
// @Accept json
// @Produce json
// @Param request body requests.CreateDMESysIDRequest true "Create DME System ID request"
// @Success 200 {object} responses.DMESysIDResponseWrapper
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/sysids [post]
func (h *DMESysIDHandler) CreateDMESysID(c echo.Context) error {
	req := new(requests.CreateDMESysIDRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	params := db.CreateDMESysIDParams{
		OrganizationID: req.OrganizationID,
		MarinaID:       uuid.Nil,
		Name:           req.Name,
		Description:    req.Description,
		SystemID:       req.SystemID,
		IsActive:       req.IsActive,
	}

	if req.MarinaID != nil {
		params.MarinaID = *req.MarinaID
	}

	sysid, err := queries.CreateDMESysID(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewDMESysIDResponseSuccess(sysid).JSON(c)
}

// GetDMESysIDByID godoc
// @Summary Get DME System ID by ID
// @Description Get DME System ID by its unique identifier
// @Tags DME System IDs
// @Accept json
// @Produce json
// @Param id path string true "DME System ID identifier"
// @Success 200 {object} responses.DMESysIDResponseWrapper
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/sysids/{id} [get]
func (h *DMESysIDHandler) GetDMESysIDByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid ID").JSON(c)
	}

	queries := h.server.DB.Queries()
	sysid, err := queries.GetDMESysIDByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, err).JSON(c)
	}

	return responses.NewDMESysIDResponseSuccess(sysid).JSON(c)
}

// GetDMESysIDBySystemID godoc
// @Summary Get DME System ID by system identifier
// @Description Get DME System ID by its external system identifier
// @Tags DME System IDs
// @Accept json
// @Produce json
// @Param systemId path string true "System identifier"
// @Success 200 {object} responses.DMESysIDResponseWrapper
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/sysids/system/{systemId} [get]
func (h *DMESysIDHandler) GetDMESysIDBySystemID(c echo.Context) error {
	systemID := c.Param("systemId")
	if systemID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required").JSON(c)
	}

	queries := h.server.DB.Queries()
	sysid, err := queries.GetDMESysIDBySystemID(c.Request().Context(), systemID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, err).JSON(c)
	}

	return responses.NewDMESysIDResponseSuccess(sysid).JSON(c)
}

// GetDMESysIDsByOrgID godoc
// @Summary Get DME System IDs by organization ID
// @Description Get all DME System IDs for an organization
// @Tags DME System IDs
// @Accept json
// @Produce json
// @Param organizationId path string true "Organization ID"
// @Success 200 {object} responses.DMESysIDListResponseWrapper
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/sysids/organization/{organizationId} [get]
func (h *DMESysIDHandler) GetDMESysIDsByOrgID(c echo.Context) error {
	orgIDStr := c.Param("organizationId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID").JSON(c)
	}

	queries := h.server.DB.Queries()
	sysids, err := queries.GetDMESysIDsByOrgID(c.Request().Context(), orgID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewDMESysIDListResponse(sysids).JSON(c)
}

// GetDMESysIDByMarinaID godoc
// @Summary Get DME System ID by marina ID
// @Description Get DME System ID linked to a specific marina
// @Tags DME System IDs
// @Accept json
// @Produce json
// @Param marinaId path string true "Marina ID"
// @Success 200 {object} responses.DMESysIDResponseWrapper
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/sysids/marina/{marinaId} [get]
func (h *DMESysIDHandler) GetDMESysIDByMarinaID(c echo.Context) error {
	marinaIDStr := c.Param("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID").JSON(c)
	}

	queries := h.server.DB.Queries()
	sysid, err := queries.GetDMESysIDByMarinaID(c.Request().Context(), marinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, err).JSON(c)
	}

	return responses.NewDMESysIDResponseSuccess(sysid).JSON(c)
}

// ListDMESysIDs godoc
// @Summary List all DME System IDs
// @Description Get a list of all DME System IDs
// @Tags DME System IDs
// @Accept json
// @Produce json
// @Success 200 {object} responses.DMESysIDListResponseWrapper
// @Failure 500 {object} responses.Error
// @Router /dme/sysids [get]
func (h *DMESysIDHandler) ListDMESysIDs(c echo.Context) error {
	queries := h.server.DB.Queries()
	sysids, err := queries.ListDMESysIDs(c.Request().Context())
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewDMESysIDListResponse(sysids).JSON(c)
}

// UpdateDMESysID godoc
// @Summary Update DME System ID
// @Description Update a DME System ID
// @Tags DME System IDs
// @Accept json
// @Produce json
// @Param id path string true "DME System ID identifier"
// @Param request body requests.UpdateDMESysIDRequest true "Update DME System ID request"
// @Success 200 {object} responses.DMESysIDResponseWrapper
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/sysids/{id} [put]
func (h *DMESysIDHandler) UpdateDMESysID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid ID").JSON(c)
	}

	req := new(requests.UpdateDMESysIDRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	// First, get the current system ID to get the original values
	originalSysID, err := queries.GetDMESysIDByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "DME System ID not found").JSON(c)
	}

	// Set up parameters for update
	params := db.UpdateDMESysIDParams{
		ID:             id,
		OrganizationID: originalSysID.OrganizationID,
		MarinaID:       originalSysID.MarinaID,
		Name:           originalSysID.Name,
		Description:    originalSysID.Description,
		SystemID:       originalSysID.SystemID,
		IsActive:       originalSysID.IsActive,
	}

	// Update with new values where provided
	if req.MarinaID != nil {
		params.MarinaID = *req.MarinaID
	}
	if req.Name != nil {
		params.Name = *req.Name
	}
	if req.Description != nil {
		params.Description = req.Description
	}
	if req.SystemID != nil {
		params.SystemID = *req.SystemID
	}
	if req.IsActive != nil {
		params.IsActive = req.IsActive
	}

	sysid, err := queries.UpdateDMESysID(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewDMESysIDResponseSuccess(sysid).JSON(c)
}

// LinkDMESysIDToMarina godoc
// @Summary Link DME System ID to marina
// @Description Link a DME System ID to a specific marina
// @Tags DME System IDs
// @Accept json
// @Produce json
// @Param id path string true "DME System ID identifier"
// @Param request body requests.LinkDMESysIDRequest true "Link DME System ID request"
// @Success 200 {object} responses.DMESysIDResponseWrapper
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/sysids/{id}/link [patch]
func (h *DMESysIDHandler) LinkDMESysIDToMarina(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid ID").JSON(c)
	}

	req := new(requests.LinkDMESysIDRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	// Check if the system ID exists
	_, err = queries.GetDMESysIDByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "DME System ID not found").JSON(c)
	}

	params := db.LinkDMESysIDToMarinaParams{
		ID:       id,
		MarinaID: req.MarinaID,
	}
	marina, err := queries.GetMarinaByID(c.Request().Context(), req.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(c)
	}

	sysid, err := queries.LinkDMESysIDToMarina(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	_, err = queries.UpdateMarinaSystemID(c.Request().Context(), db.UpdateMarinaSystemIDParams{
		ID:       marina.ID,
		SystemID: &sysid.SystemID,
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewDMESysIDResponseSuccess(sysid).JSON(c)
}

// UnlinkDMESysIDFromMarina godoc
// @Summary Unlink DME System ID from marina
// @Description Remove the link between a DME System ID and a marina
// @Tags DME System IDs
// @Accept json
// @Produce json
// @Param id path string true "DME System ID identifier"
// @Success 200 {object} responses.DMESysIDResponseWrapper
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/sysids/{id}/unlink [patch]
func (h *DMESysIDHandler) UnlinkDMESysIDFromMarina(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid ID").JSON(c)
	}

	queries := h.server.DB.Queries()

	// Check if the system ID exists
	sysid, err := queries.GetDMESysIDByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "DME System ID not found").JSON(c)
	}

	marina, err := queries.GetMarinaByID(c.Request().Context(), sysid.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(c)
	}

	updatedSysID, err := queries.UnlinkDMESysIDFromMarina(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	_, err = queries.UpdateMarinaSystemID(c.Request().Context(), db.UpdateMarinaSystemIDParams{
		ID:       marina.ID,
		SystemID: nil,
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewDMESysIDResponseSuccess(updatedSysID).JSON(c)
}

// DeleteDMESysID godoc
// @Summary Delete DME System ID
// @Description Soft delete a DME System ID
// @Tags DME System IDs
// @Accept json
// @Produce json
// @Param id path string true "DME System ID identifier"
// @Success 204 "No Content"
// @Failure 400 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /dme/sysids/{id} [delete]
func (h *DMESysIDHandler) DeleteDMESysID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid ID").JSON(c)
	}

	queries := h.server.DB.Queries()

	// Check if the system ID exists
	_, err = queries.GetDMESysIDByIDWithDeleted(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "DME System ID not found").JSON(c)
	}

	err = queries.HardDeleteDMESysID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return c.NoContent(http.StatusNoContent)
}
