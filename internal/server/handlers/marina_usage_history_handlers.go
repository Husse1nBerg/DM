package handlers

import (
	"database/sql"
	"errors"
	"net/http"
	"strconv"
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

// Helper to parse pagination params
func parsePaginationParams(c echo.Context) (page, pageSize int32) {
	page = 1
	pageSize = 10
	if p := c.QueryParam("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = int32(v)
		}
	}
	if pp := c.QueryParam("pageSize"); pp != "" {
		if v, err := strconv.Atoi(pp); err == nil && v > 0 {
			pageSize = int32(v)
		}
	}
	return
}

// GetMarinaUsageHistoryByMarinaID handles retrieving marina usage history records for a specific marina
// @Summary Get marina usage history by marina ID
// @Description Retrieves all marina usage history records for a specific marina
// @Tags marina-usage-history
// @Produce json
// @Param marinaId path string true "Marina ID"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
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

	page, pageSize := parsePaginationParams(c)
	offset := (page - 1) * pageSize
	limit := pageSize

	histories, err := h.server.DB.Queries().GetMarinaUsageHistoryByMarinaIDPaginated(c.Request().Context(), db.GetMarinaUsageHistoryByMarinaIDPaginatedParams{
		MarinaID: marinaID,
		Limit:    limit,
		Offset:   offset,
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total, _ := h.server.DB.Queries().GetMarinaUsageHistoryByMarinaIDTotal(c.Request().Context(), marinaID)
	return responses.NewMarinaUsageHistoryPaginatedResponse(histories, total, pageSize, page).JSON(c)
}

// GetMarinaUsageHistoryByDateRange handles retrieving marina usage history records within a date range
// @Summary Get marina usage history by date range
// @Description Retrieves marina usage history records within a specified date range for a specific marina
// @Tags marina-usage-history
// @Produce json
// @Param marinaId query string true "Marina ID"
// @Param startDate query string true "Start Date (YYYY-MM-DD)"
// @Param endDate query string true "End Date (YYYY-MM-DD)"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
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
	page, pageSize := parsePaginationParams(c)
	offset := (page - 1) * pageSize
	limit := pageSize

	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	endTime = endTime.Add(24*time.Hour - time.Second)

	histories, err := h.server.DB.Queries().GetMarinaUsageHistoryByDateRangePaginated(c.Request().Context(), db.GetMarinaUsageHistoryByDateRangePaginatedParams{
		MarinaID:    marinaID,
		CreatedAt:   pgtype.Timestamptz{Time: startTime, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: endTime, Valid: true},
		Limit:       limit,
		Offset:      offset,
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total, _ := h.server.DB.Queries().GetMarinaUsageHistoryByDateRangeTotal(c.Request().Context(), db.GetMarinaUsageHistoryByDateRangeTotalParams{
		MarinaID:    marinaID,
		CreatedAt:   pgtype.Timestamptz{Time: startTime, Valid: true},
		CreatedAt_2: pgtype.Timestamptz{Time: endTime, Valid: true},
	})
	return responses.NewMarinaUsageHistoryPaginatedResponse(histories, total, pageSize, page).JSON(c)
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
// @Description Retrieves all marina usage history records, optionally filtered by one or more marinaIds and date range
// @Tags marina-usage-history
// @Produce json
// @Param marinaId query []string false "Marina ID(s) (repeat for multiple)" collectionFormat(multi)
// @Param startDate query string false "Start Date (YYYY-MM-DD)"
// @Param endDate query string false "End Date (YYYY-MM-DD)"
// @Param page query int false "Page number"
// @Param pageSize query int false "Page size"
// @Success 200 {object} responses.MarinaUsageHistoryListResponse
// @Failure 400 {object} responses.BaseResponse
// @Failure 500 {object} responses.BaseResponse
// @Router /marina-usage-history/all [get]
func (h *MarinaUsageHistoryHandler) GetAllMarinaUsageHistory(c echo.Context) error {
	marinaIDStrs := c.QueryParams()["marinaId"]
	startDate := c.QueryParam("startDate")
	endDate := c.QueryParam("endDate")
	page, pageSize := parsePaginationParams(c)

	var (
		marinaIDs          []uuid.UUID
		hasMarinaIDs       bool
		startTime, endTime time.Time
		hasStart, hasEnd   bool
	)

	if len(marinaIDStrs) > 0 {
		for _, idStr := range marinaIDStrs {
			if idStr == "" {
				continue
			}
			id, err := uuid.Parse(idStr)
			if err != nil {
				return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marinaId format").JSON(c)
			}
			marinaIDs = append(marinaIDs, id)
		}
		hasMarinaIDs = len(marinaIDs) > 0
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
		endTime = endTime.Add(24*time.Hour - time.Second)
		hasEnd = true
	}

	var histories []db.MarinaUsageHistory
	var total int64
	var err error

	offset := (page - 1) * pageSize
	limit := pageSize

	switch {
	case hasMarinaIDs && hasStart && hasEnd:
		histories, err = h.server.DB.Queries().GetMarinaUsageHistoryByMarinaIDsAndDateRangePaginated(c.Request().Context(), db.GetMarinaUsageHistoryByMarinaIDsAndDateRangePaginatedParams{
			Column1:     marinaIDs,
			CreatedAt:   pgtype.Timestamptz{Time: startTime, Valid: true},
			CreatedAt_2: pgtype.Timestamptz{Time: endTime, Valid: true},
			Limit:       limit,
			Offset:      offset,
		})
		total, _ = h.server.DB.Queries().GetMarinaUsageHistoryByMarinaIDsAndDateRangeTotal(c.Request().Context(), db.GetMarinaUsageHistoryByMarinaIDsAndDateRangeTotalParams{
			Column1:     marinaIDs,
			CreatedAt:   pgtype.Timestamptz{Time: startTime, Valid: true},
			CreatedAt_2: pgtype.Timestamptz{Time: endTime, Valid: true},
		})
	case hasMarinaIDs:
		histories, err = h.server.DB.Queries().GetMarinaUsageHistoryByMarinaIDsPaginated(c.Request().Context(), db.GetMarinaUsageHistoryByMarinaIDsPaginatedParams{
			Column1: marinaIDs,
			Limit:   limit,
			Offset:  offset,
		})
		total, _ = h.server.DB.Queries().GetMarinaUsageHistoryByMarinaIDsTotal(c.Request().Context(), marinaIDs)
	case hasStart && hasEnd:
		histories, err = h.server.DB.Queries().GetAllMarinaUsageHistoryByDateRangePaginated(c.Request().Context(), db.GetAllMarinaUsageHistoryByDateRangePaginatedParams{
			CreatedAt:   pgtype.Timestamptz{Time: startTime, Valid: true},
			CreatedAt_2: pgtype.Timestamptz{Time: endTime, Valid: true},
			Limit:       limit,
			Offset:      offset,
		})
		total, _ = h.server.DB.Queries().GetAllMarinaUsageHistoryByDateRangeTotal(c.Request().Context(), db.GetAllMarinaUsageHistoryByDateRangeTotalParams{
			CreatedAt:   pgtype.Timestamptz{Time: startTime, Valid: true},
			CreatedAt_2: pgtype.Timestamptz{Time: endTime, Valid: true},
		})
	default:
		histories, err = h.server.DB.Queries().GetAllMarinaUsageHistoryPaginated(c.Request().Context(), db.GetAllMarinaUsageHistoryPaginatedParams{
			Limit:  limit,
			Offset: offset,
		})
		total, _ = h.server.DB.Queries().GetAllMarinaUsageHistoryTotal(c.Request().Context())
	}

	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMarinaUsageHistoryPaginatedResponse(histories, total, pageSize, page).JSON(c)
}
