package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// MarinaUsageHistoryHandler handles marina usage history-related operations
type MarinaUsageHistoryHandler struct {
	server *s.Server
}

// NewMarinaUsageHistoryHandler creates a new marina usage history handler
func NewMarinaUsageHistoryHandler(server *s.Server) *MarinaUsageHistoryHandler {
	return &MarinaUsageHistoryHandler{server: server}
}

// GetMarinaUsageHistoryByMarinaID handles retrieving marina usage history records for a specific marina
// @Summary Get marina usage history by marina ID
// @Description Retrieves all marina usage history records for a specific marina
// @Tags marina-usage-history
// @Produce json
// @Param marinaId path string true "Marina ID"
// @Success 200 {object} responses.MarinaUsageHistoryListResponse
// @Failure 400 {object} responses.BaseResponse
// @Failure 500 {object} responses.BaseResponse
// @Router /marina-usage-history/marina/{marinaId} [get]
func (h *MarinaUsageHistoryHandler) GetMarinaUsageHistoryByMarinaID(c echo.Context) error {
	marinaIDStr := c.Param("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	histories, err := h.server.DB.Queries().GetMarinaUsageHistoryByMarinaID(c.Request().Context(), marinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMarinaUsageHistoryListResponse(histories).JSON(c)
}

// GetMarinaUsageHistoryByDateRange handles retrieving marina usage history records within a date range
// @Summary Get marina usage history by date range
// @Description Retrieves marina usage history records within a specified date range for a specific marina
// @Tags marina-usage-history
// @Produce json
// @Param marinaId query string true "Marina ID"
// @Param startDate query string true "Start Date (YYYY-MM-DD)"
// @Param endDate query string true "End Date (YYYY-MM-DD)"
// @Success 200 {object} responses.MarinaUsageHistoryListResponse
// @Failure 400 {object} responses.BaseResponse
// @Failure 500 {object} responses.BaseResponse
// @Router /marina-usage-history/{marinaId}/date-range [get]
func (h *MarinaUsageHistoryHandler) GetMarinaUsageHistoryByDateRange(c echo.Context) error {
	marinaIDStr := c.Param("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	startDate := c.QueryParam("startDate")
	endDate := c.QueryParam("endDate")

	// Parse start date
	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Parse end date
	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Set end time to end of day
	endTime = endTime.Add(24*time.Hour - time.Second)

	// Convert to pgtype.Timestamptz
	startTimestamp := pgtype.Timestamptz{
		Time:  startTime,
		Valid: true,
	}
	endTimestamp := pgtype.Timestamptz{
		Time:  endTime,
		Valid: true,
	}

	histories, err := h.server.DB.Queries().GetMarinaUsageHistoryByDateRange(c.Request().Context(), db.GetMarinaUsageHistoryByDateRangeParams{
		MarinaID:    marinaID,
		CreatedAt:   startTimestamp,
		CreatedAt_2: endTimestamp,
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMarinaUsageHistoryListResponse(histories).JSON(c)
}

// GetLatestMarinaUsageHistory handles retrieving the latest marina usage history record for a specific marina
// @Summary Get latest marina usage history
// @Description Retrieves the most recent marina usage history record for a specific marina
// @Tags marina-usage-history
// @Produce json
// @Param marinaId path string true "Marina ID"
// @Success 200 {object} responses.MarinaUsageHistoryResponseWrapper
// @Failure 400 {object} responses.BaseResponse
// @Failure 404 {object} responses.BaseResponse
// @Failure 500 {object} responses.BaseResponse
// @Router /marina-usage-history/{marinaId}/latest [get]
func (h *MarinaUsageHistoryHandler) GetLatestMarinaUsageHistory(c echo.Context) error {
	marinaIDStr := c.Param("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	history, err := h.server.DB.Queries().GetLatestMarinaUsageHistory(c.Request().Context(), marinaID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.NewErrorResponse(http.StatusNotFound, err).JSON(c)
		}
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMarinaUsageHistoryResponseSuccess(history).JSON(c)
}

// GetMarinaUsageHistoryByMonth handles retrieving marina usage history records for a specific month
// @Summary Get marina usage history by month
// @Description Retrieves marina usage history records for a specific month for a specific marina
// @Tags marina-usage-history
// @Produce json
// @Param marinaId path string true "Marina ID"
// @Param month query string true "Month (YYYY-MM)"
// @Success 200 {object} responses.MarinaUsageHistoryListResponse
// @Failure 400 {object} responses.BaseResponse
// @Failure 500 {object} responses.BaseResponse
// @Router /marina-usage-history/{marinaId}/month [get]
func (h *MarinaUsageHistoryHandler) GetMarinaUsageHistoryByMonth(c echo.Context) error {
	marinaIDStr := c.Param("marinaId")
	month := c.QueryParam("month")

	if marinaIDStr == "" || month == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "marinaId and month are required").JSON(c)
	}

	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "invalid marinaId format").JSON(c)
	}

	monthTime, err := time.Parse("2006-01", month)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	monthTimestamp := pgtype.Timestamp{
		Time:  monthTime,
		Valid: true,
	}

	params := db.GetMarinaUsageHistoryByMonthParams{
		MarinaID: marinaID,
		Column2:  monthTimestamp,
	}

	history, err := h.server.DB.Queries().GetMarinaUsageHistoryByMonth(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// The query returns a single record, so wrap it in a slice for the response
	return responses.NewMarinaUsageHistoryListResponse([]db.MarinaUsageHistory{history}).JSON(c)
}

// GetMarinaUsageHistoryByID handles retrieving a marina usage history record by its ID
// @Summary Get marina usage history by ID
// @Description Retrieves a marina usage history record by its ID
// @Tags marina-usage-history
// @Produce json
// @Param id path int true "Marina Usage History ID"
// @Success 200 {object} responses.MarinaUsageHistoryResponseWrapper
// @Failure 400 {object} responses.BaseResponse
// @Failure 404 {object} responses.BaseResponse
// @Failure 500 {object} responses.BaseResponse
// @Router /marina-usage-history/usage/{id} [get]
func (h *MarinaUsageHistoryHandler) GetMarinaUsageHistoryByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid ID format").JSON(c)
	}

	history, err := h.server.DB.Queries().GetMarinaUsageHistoryByID(c.Request().Context(), id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return responses.NewErrorResponse(http.StatusNotFound, "Marina usage history not found").JSON(c)
		}
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMarinaUsageHistoryResponseSuccess(history).JSON(c)
}

// GetAllMarinaUsageHistory handles retrieving all marina usage history records with optional filters
// @Summary Get all marina usage history
// @Description Retrieves all marina usage history records, optionally filtered by marinaId and date range
// @Tags marina-usage-history
// @Produce json
// @Param marinaId query string false "Marina ID"
// @Param startDate query string false "Start Date (YYYY-MM-DD)"
// @Param endDate query string false "End Date (YYYY-MM-DD)"
// @Success 200 {object} responses.MarinaUsageHistoryListResponse
// @Failure 400 {object} responses.BaseResponse
// @Failure 500 {object} responses.BaseResponse
// @Router /marina-usage-history/all [get]
func (h *MarinaUsageHistoryHandler) GetAllMarinaUsageHistory(c echo.Context) error {
	marinaIDStr := c.QueryParam("marinaId")
	startDate := c.QueryParam("startDate")
	endDate := c.QueryParam("endDate")

	var (
		marinaID           uuid.UUID
		hasMarinaID        bool
		startTime, endTime time.Time
		hasStart, hasEnd   bool
	)

	if marinaIDStr != "" {
		var err error
		marinaID, err = uuid.Parse(marinaIDStr)
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marinaId format").JSON(c)
		}
		hasMarinaID = true
	}
	if startDate != "" {
		var err error
		startTime, err = time.Parse("2006-01-02", startDate)
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid startDate format").JSON(c)
		}
		hasStart = true
	}
	if endDate != "" {
		var err error
		endTime, err = time.Parse("2006-01-02", endDate)
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid endDate format").JSON(c)
		}
		// Set end time to end of day
		endTime = endTime.Add(24*time.Hour - time.Second)
		hasEnd = true
	}

	var histories []db.MarinaUsageHistory
	var err error

	switch {
	case hasMarinaID && hasStart && hasEnd:
		histories, err = h.server.DB.Queries().GetMarinaUsageHistoryByDateRange(c.Request().Context(), db.GetMarinaUsageHistoryByDateRangeParams{
			MarinaID:    marinaID,
			CreatedAt:   pgtype.Timestamptz{Time: startTime, Valid: true},
			CreatedAt_2: pgtype.Timestamptz{Time: endTime, Valid: true},
		})
	case hasMarinaID:
		histories, err = h.server.DB.Queries().GetMarinaUsageHistoryByMarinaID(c.Request().Context(), marinaID)
	case hasStart && hasEnd:
		histories, err = h.server.DB.Queries().GetAllMarinaUsageHistoryByDateRange(c.Request().Context(), db.GetAllMarinaUsageHistoryByDateRangeParams{
			CreatedAt:   pgtype.Timestamptz{Time: startTime, Valid: true},
			CreatedAt_2: pgtype.Timestamptz{Time: endTime, Valid: true},
		})
	default:
		histories, err = h.server.DB.Queries().GetAllMarinaUsageHistory(c.Request().Context())
	}

	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMarinaUsageHistoryListResponse(histories).JSON(c)
}
