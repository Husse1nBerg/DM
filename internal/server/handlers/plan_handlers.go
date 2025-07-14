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

// PlanHandler handles operations related to plans
type PlanHandler struct {
	server *s.Server
}

// NewPlanHandler creates a new plan handler
func NewPlanHandler(server *s.Server) *PlanHandler {
	return &PlanHandler{server: server}
}

// ListNotesMessagesPlans lists all notes and messages plans
//
//	@Summary		List notes and messages plans
//	@Description	Get all notes and messages plans with pagination
//	@Tags			Plans
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Success		200	{object}	responses.BaseResponse	"Paginated list of plans"
//	@Failure		500	{object}	responses.Error	"Server error"
//	@Router			/plans/notes-messages [get]
func (h *PlanHandler) ListNotesMessagesPlans(c echo.Context) error {
	// Parse pagination parameters
	var req requests.PaginationQuery
	if err := c.Bind(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	// Set defaults if not provided
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	queries := h.server.DB.Queries()

	// Get total count
	allPlans, err := queries.ListNotesMessagesPlans(c.Request().Context())
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allPlans))

	// Get paginated data
	params := db.ListNotesMessagesPlansPaginatedParams{
		Limit:  req.PageSize,
		Offset: (req.Page - 1) * req.PageSize,
	}
	plans, err := queries.ListNotesMessagesPlansPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewNotesMessagesPlansPaginatedResponse(plans, total, req.PageSize, req.Page).JSON(c)
}

// GetNotesMessagesPlan gets a specific notes and messages plan by ID
//
//	@Summary		Get notes and messages plan
//	@Description	Get a specific notes and messages plan by ID
//	@Tags			Plans
//	@Accept			json
//	@Produce		json
//	@Param			planId	path		string	true	"Plan ID"
//	@Success		200	{object}	responses.BaseResponse	"Plan details"
//	@Failure		400	{object}	responses.Error	"Invalid plan ID"
//	@Failure		404	{object}	responses.Error	"Plan not found"
//	@Failure		500	{object}	responses.Error	"Server error"
//	@Router			/plans/notes-messages/{planId} [get]
func (h *PlanHandler) GetNotesMessagesPlan(c echo.Context) error {
	planIDStr := c.Param("planId")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid plan ID").JSON(c)
	}

	queries := h.server.DB.Queries()
	plan, err := queries.GetNotesMessagesPlanByID(c.Request().Context(), planID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Plan not found").JSON(c)
	}

	return responses.NewNotesMessagesPlanResponseSuccess(plan).JSON(c)
}

// ListStoragePlans lists all storage plans
//
//	@Summary		List storage plans
//	@Description	Get all storage plans with pagination
//	@Tags			Plans
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Success		200	{object}	responses.BaseResponse	"Paginated list of plans"
//	@Failure		500	{object}	responses.Error	"Server error"
//	@Router			/plans/storage [get]
func (h *PlanHandler) ListStoragePlans(c echo.Context) error {
	// Parse pagination parameters
	var req requests.PaginationQuery
	if err := c.Bind(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	// Set defaults if not provided
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	queries := h.server.DB.Queries()

	// Get total count
	allPlans, err := queries.ListStoragePlans(c.Request().Context())
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allPlans))

	// Get paginated data
	params := db.ListStoragePlansPaginatedParams{
		Limit:  req.PageSize,
		Offset: (req.Page - 1) * req.PageSize,
	}
	plans, err := queries.ListStoragePlansPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewStoragePlansPaginatedResponse(plans, total, req.PageSize, req.Page).JSON(c)
}

// GetStoragePlan gets a specific storage plan by ID
//
//	@Summary		Get storage plan
//	@Description	Get a specific storage plan by ID
//	@Tags			Plans
//	@Accept			json
//	@Produce		json
//	@Param			planId	path		string	true	"Plan ID"
//	@Success		200	{object}	responses.BaseResponse	"Plan details"
//	@Failure		400	{object}	responses.Error	"Invalid plan ID"
//	@Failure		404	{object}	responses.Error	"Plan not found"
//	@Failure		500	{object}	responses.Error	"Server error"
//	@Router			/plans/storage/{planId} [get]
func (h *PlanHandler) GetStoragePlan(c echo.Context) error {
	planIDStr := c.Param("planId")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid plan ID").JSON(c)
	}

	queries := h.server.DB.Queries()
	plan, err := queries.GetStoragePlanByID(c.Request().Context(), planID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Plan not found").JSON(c)
	}

	return responses.NewStoragePlanResponseSuccess(plan).JSON(c)
}

// ListDocumentPlans lists all document plans
//
//	@Summary		List document plans
//	@Description	Get all document plans with pagination
//	@Tags		Plans
//	@Accept		json
//	@Produce		json
//	@Param		page		query		int	false	"Page number"	default(1)
//	@Param		pageSize	query		int	false	"Page size"	default(10)
//	@Success	200	{object} responses.BaseResponse	"Paginated list of plans"
//	@Failure	500	{object} responses.Error	"Server error"
//	@Router		/plans/document [get]
func (h *PlanHandler) ListDocumentPlans(c echo.Context) error {
	var req requests.PaginationQuery
	if err := c.Bind(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	queries := h.server.DB.Queries()
	allPlans, err := queries.GetAllDocumentPlans(c.Request().Context())
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allPlans))
	start := (req.Page - 1) * req.PageSize
	end := start + req.PageSize
	if int64(start) > total {
		start = int32(total)
	}
	if int64(end) > total {
		end = int32(total)
	}
	plans := allPlans[start:end]
	return responses.NewDocumentPlansPaginatedResponse(plans, total, req.PageSize, req.Page).JSON(c)
}

// GetDocumentPlan gets a specific document plan by ID
//
//	@Summary		Get document plan
//	@Description	Get a specific document plan by ID
//	@Tags		Plans
//	@Accept		json
//	@Produce		json
//	@Param		planId	path	string	true	"Plan ID"
//	@Success	200	{object} responses.BaseResponse	"Plan details"
//	@Failure	400	{object} responses.Error	"Invalid plan ID"
//	@Failure	404	{object} responses.Error	"Plan not found"
//	@Failure	500	{object} responses.Error	"Server error"
//	@Router		/plans/document/{planId} [get]
func (h *PlanHandler) GetDocumentPlan(c echo.Context) error {
	planIDStr := c.Param("planId")
	planID, err := uuid.Parse(planIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid plan ID").JSON(c)
	}
	queries := h.server.DB.Queries()
	plan, err := queries.GetDocumentPlanByID(c.Request().Context(), planID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Plan not found").JSON(c)
	}
	return responses.NewDocumentPlanResponseSuccess(plan).JSON(c)
}
