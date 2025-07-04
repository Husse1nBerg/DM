package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// CriteriaHandler handles criteria-related operations
type CriteriaHandler struct {
	server *s.Server
}

// NewCriteriaHandler creates a new criteria handler
func NewCriteriaHandler(server *s.Server) *CriteriaHandler {
	return &CriteriaHandler{server: server}
}

// extractMarinaIDFromToken extracts marina ID from JWT token claims
func (h *CriteriaHandler) extractMarinaIDFromToken(c echo.Context) (uuid.UUID, error) {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return uuid.UUID{}, err
	}
	return user.MarinaID, nil
}

// ListCriteria lists all criteria for the authenticated user's marina
//
//	@Summary		List criteria
//	@Description	Lists all criteria for the authenticated user's marina
//	@Tags			Criteria
//	@Accept			json
//	@Produce		json
//	@Success		200			{object}	responses.CriteriaListResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/criteria [get]
func (h *CriteriaHandler) ListCriteria(c echo.Context) error {
	marinaID, err := h.extractMarinaIDFromToken(c)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token").JSON(c)
	}

	criteria, err := h.server.DB.Queries().ListCriteria(c.Request().Context(), marinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewCriteriaListResponse(criteria).JSON(c)
}

// ListCriteriaPaginated lists criteria for the authenticated user's marina with pagination
//
//	@Summary		List criteria (paginated)
//	@Description	Lists criteria for the authenticated user's marina with pagination
//	@Tags			Criteria
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number (default: 1)"
//	@Param			pageSize	query		int		false	"Page size (default: 10)"
//	@Success		200			{object}	responses.CriteriaPaginatedResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/criteria/paginated [get]
func (h *CriteriaHandler) ListCriteriaPaginated(c echo.Context) error {
	marinaID, err := h.extractMarinaIDFromToken(c)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token").JSON(c)
	}

	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	// Get paginated criteria
	criteria, err := h.server.DB.Queries().ListCriteriaPaginated(c.Request().Context(), db.ListCriteriaPaginatedParams{
		MarinaID: marinaID,
		Limit:    pagination.PageSize,
		Offset:   (pagination.Page - 1) * pagination.PageSize,
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Get total count for pagination
	total, err := h.server.DB.Queries().GetCriteriaCount(c.Request().Context(), marinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewCriteriaPaginatedResponse(criteria, total, pagination.PageSize, pagination.Page).JSON(c)
}

// GetCriteria retrieves a single criteria by ID
//
//	@Summary		Get criteria
//	@Description	Retrieves a single criteria by ID
//	@Tags			Criteria
//	@Accept			json
//	@Produce		json
//	@Param			criteriaId	path		string	true	"Criteria ID"	Format(uuid)
//	@Success		200			{object}	responses.CriteriaResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		404			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/criteria/{criteriaId} [get]
func (h *CriteriaHandler) GetCriteria(c echo.Context) error {
	criteriaIDStr := c.Param("criteriaId")
	criteriaID, err := uuid.Parse(criteriaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid criteria ID").JSON(c)
	}

	criteria, err := h.server.DB.Queries().GetCriteriaByID(c.Request().Context(), criteriaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Criteria not found").JSON(c)
	}

	// Verify that the criteria belongs to the user's marina
	marinaID, err := h.extractMarinaIDFromToken(c)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token").JSON(c)
	}

	if criteria.MarinaID != marinaID {
		return responses.NewErrorResponse(http.StatusNotFound, "Criteria not found").JSON(c)
	}

	return responses.NewCriteriaResponse(criteria).JSON(c)
}

// CreateCriteria creates a new criteria for the authenticated user's marina
//
//	@Summary		Create criteria
//	@Description	Creates a new criteria for the authenticated user's marina
//	@Tags			Criteria
//	@Accept			json
//	@Produce		json
//	@Param			criteria	body		requests.CreateCriteriaRequest	true	"Criteria details"
//	@Success		201			{object}	responses.CriteriaResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/criteria [post]
func (h *CriteriaHandler) CreateCriteria(c echo.Context) error {
	marinaID, err := h.extractMarinaIDFromToken(c)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token").JSON(c)
	}

	var req requests.CreateCriteriaRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := req.Validate(); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Convert fields to DB format
	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	// Convert interface{} to JSON bytes
	criteriaBytes, err := json.Marshal(req.Criteria)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid criteria format").JSON(c)
	}

	params := db.CreateCriteriaParams{
		MarinaID:    marinaID,
		Name:        req.Name,
		Description: description,
		Criteria:    criteriaBytes,
	}

	criteria, err := h.server.DB.Queries().CreateCriteria(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return c.JSON(http.StatusCreated, responses.NewCriteriaResponse(criteria))
}

// UpdateCriteria updates an existing criteria
//
//	@Summary		Update criteria
//	@Description	Updates an existing criteria
//	@Tags			Criteria
//	@Accept			json
//	@Produce		json
//	@Param			criteriaId	path		string							true	"Criteria ID"	Format(uuid)
//	@Param			criteria	body		requests.UpdateCriteriaRequest	true	"Criteria details"
//	@Success		200			{object}	responses.CriteriaResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		404			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/criteria/{criteriaId} [put]
func (h *CriteriaHandler) UpdateCriteria(c echo.Context) error {
	criteriaIDStr := c.Param("criteriaId")
	criteriaID, err := uuid.Parse(criteriaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid criteria ID").JSON(c)
	}

	marinaID, err := h.extractMarinaIDFromToken(c)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token").JSON(c)
	}

	var req requests.UpdateCriteriaRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := req.Validate(); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Check if criteria exists and belongs to user's marina
	existingCriteria, err := h.server.DB.Queries().GetCriteriaByID(c.Request().Context(), criteriaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Criteria not found").JSON(c)
	}

	if existingCriteria.MarinaID != marinaID {
		return responses.NewErrorResponse(http.StatusNotFound, "Criteria not found").JSON(c)
	}

	// Convert fields to DB format
	var description *string
	if req.Description != "" {
		description = &req.Description
	}

	// Convert interface{} to JSON bytes
	criteriaBytes, err := json.Marshal(req.Criteria)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid criteria format").JSON(c)
	}

	params := db.UpdateCriteriaParams{
		ID:          criteriaID,
		Name:        req.Name,
		Description: description,
		Criteria:    criteriaBytes,
	}

	criteria, err := h.server.DB.Queries().UpdateCriteria(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewCriteriaResponse(criteria).JSON(c)
}

// DeleteCriteria deletes a criteria (soft delete)
//
//	@Summary		Delete criteria
//	@Description	Deletes a criteria (soft delete)
//	@Tags			Criteria
//	@Accept			json
//	@Produce		json
//	@Param			criteriaId	path		string	true	"Criteria ID"	Format(uuid)
//	@Success		204			{object}	responses.BaseResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		404			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/criteria/{criteriaId} [delete]
func (h *CriteriaHandler) DeleteCriteria(c echo.Context) error {
	criteriaIDStr := c.Param("criteriaId")
	criteriaID, err := uuid.Parse(criteriaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid criteria ID").JSON(c)
	}

	marinaID, err := h.extractMarinaIDFromToken(c)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token").JSON(c)
	}

	// Check if criteria exists and belongs to user's marina
	existingCriteria, err := h.server.DB.Queries().GetCriteriaByID(c.Request().Context(), criteriaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Criteria not found").JSON(c)
	}

	if existingCriteria.MarinaID != marinaID {
		return responses.NewErrorResponse(http.StatusNotFound, "Criteria not found").JSON(c)
	}

	err = h.server.DB.Queries().DeleteCriteria(c.Request().Context(), criteriaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return c.NoContent(http.StatusNoContent)
}

// SearchCriteria searches criteria by name for the authenticated user's marina
//
//	@Summary		Search criteria
//	@Description	Searches criteria by name for the authenticated user's marina
//	@Tags			Criteria
//	@Accept			json
//	@Produce		json
//	@Param			query		query		string	true	"Search query"
//	@Success		200			{object}	responses.CriteriaListResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/criteria/search [get]
func (h *CriteriaHandler) SearchCriteria(c echo.Context) error {
	marinaID, err := h.extractMarinaIDFromToken(c)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token").JSON(c)
	}

	query := c.QueryParam("query")
	if query == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Query parameter is required").JSON(c)
	}

	criteria, err := h.server.DB.Queries().SearchCriteriaByName(c.Request().Context(), db.SearchCriteriaByNameParams{
		MarinaID: marinaID,
		Column2:  &query,
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewCriteriaListResponse(criteria).JSON(c)
}

// SearchCriteriaPaginated searches criteria by name with pagination for the authenticated user's marina
//
//	@Summary		Search criteria (paginated)
//	@Description	Searches criteria by name with pagination for the authenticated user's marina
//	@Tags			Criteria
//	@Accept			json
//	@Produce		json
//	@Param			query		query		string	true	"Search query"
//	@Param			page		query		int		false	"Page number (default: 1)"
//	@Param			pageSize	query		int		false	"Page size (default: 10)"
//	@Success		200			{object}	responses.CriteriaPaginatedResponse
//	@Failure		400			{object}	responses.Error
//	@Failure		500			{object}	responses.Error
//	@Security		ApiKeyAuth
//	@Router			/criteria/search/paginated [get]
func (h *CriteriaHandler) SearchCriteriaPaginated(c echo.Context) error {
	marinaID, err := h.extractMarinaIDFromToken(c)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token").JSON(c)
	}

	query := c.QueryParam("query")
	if query == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Query parameter is required").JSON(c)
	}

	// Parse pagination params
	page := 1
	pageSize := 10
	if pageStr := c.QueryParam("page"); pageStr != "" {
		if p, err := strconv.Atoi(pageStr); err == nil && p > 0 {
			page = p
		}
	}
	if pageSizeStr := c.QueryParam("pageSize"); pageSizeStr != "" {
		if ps, err := strconv.Atoi(pageSizeStr); err == nil && ps > 0 {
			pageSize = ps
		}
	}

	criteria, err := h.server.DB.Queries().SearchCriteriaByNamePaginated(c.Request().Context(), db.SearchCriteriaByNamePaginatedParams{
		MarinaID: marinaID,
		Column2:  &query,
		Limit:    int32(pageSize),
		Offset:   int32((page - 1) * pageSize),
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// For search results, we'll return the count of results rather than total count
	// as getting exact count for search can be expensive
	total := int64(len(criteria))

	return responses.NewCriteriaPaginatedResponse(criteria, total, int32(pageSize), int32(page)).JSON(c)
}
