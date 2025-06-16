package requests

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/go-playground/validator/v10"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/labstack/echo/v4"
)

// CreateMarinaUsageHistoryRequest represents the required parameters to create a new marina usage history record
type CreateMarinaUsageHistoryRequest struct {
	MarinaID     uuid.UUID `json:"marinaId" validate:"required" example:"550e8400-e29b-41d4-a716-446655440001"`
	StorageUsage float64   `json:"storageUsage" validate:"required" example:"1.5"`
	EmailUsage   int16     `json:"emailUsage" validate:"required" example:"100"`
	TextUsage    int16     `json:"textUsage" validate:"required" example:"50"`
}

// UpdateMarinaUsageHistoryRequest represents the parameters that can be updated for a marina usage history record
type UpdateMarinaUsageHistoryRequest struct {
	StorageUsage *float64 `json:"storageUsage,omitempty" example:"1.5"`
	EmailUsage   *int16   `json:"emailUsage,omitempty" example:"100"`
	TextUsage    *int16   `json:"textUsage,omitempty" example:"50"`
}

// Validate performs custom validation on the create marina usage history request
func (r *CreateMarinaUsageHistoryRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// Validate performs custom validation on the update marina usage history request
func (r *UpdateMarinaUsageHistoryRequest) Validate() error {
	validate := validator.New()
	return validate.Struct(r)
}

// CreateMarinaUsageHistory handles the creation of a new marina usage history record
func CreateMarinaUsageHistory(c echo.Context, queries *db.Queries) error {
	var req CreateMarinaUsageHistoryRequest
	if err := c.Bind(&req); err != nil {
		return err
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	// Convert storage usage from GB to bytes
	const bytesInGB = 1024 * 1024 * 1024 // 1 GB in bytes
	storageUsageBytes := int64(req.StorageUsage * float64(bytesInGB))

	// Create marina usage history record
	history, err := queries.CreateMarinaUsageHistory(c.Request().Context(), db.CreateMarinaUsageHistoryParams{
		MarinaID:     req.MarinaID,
		StorageUsage: storageUsageBytes,
		EmailUsage:   req.EmailUsage,
		TextUsage:    req.TextUsage,
	})
	if err != nil {
		return err
	}

	return c.JSON(200, history)
}

// GetMarinaUsageHistoryByID handles retrieving a marina usage history record by its ID
func GetMarinaUsageHistoryByID(c echo.Context, queries *db.Queries) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return err
	}

	history, err := queries.GetMarinaUsageHistoryByID(c.Request().Context(), id)
	if err != nil {
		return err
	}

	return c.JSON(200, history)
}

// GetMarinaUsageHistoryByMarinaID handles retrieving marina usage history records for a specific marina
func GetMarinaUsageHistoryByMarinaID(c echo.Context, queries *db.Queries) error {
	marinaIDStr := c.Param("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return err
	}

	histories, err := queries.GetMarinaUsageHistoryByMarinaID(c.Request().Context(), marinaID)
	if err != nil {
		return err
	}

	return c.JSON(200, histories)
}

// GetMarinaUsageHistoryByDateRange handles retrieving marina usage history records within a date range
func GetMarinaUsageHistoryByDateRange(c echo.Context, queries *db.Queries) error {
	marinaIDStr := c.Param("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return err
	}

	startDate := c.QueryParam("startDate")
	endDate := c.QueryParam("endDate")

	// Parse start date
	startTime, err := time.Parse("2006-01-02", startDate)
	if err != nil {
		return err
	}

	// Parse end date
	endTime, err := time.Parse("2006-01-02", endDate)
	if err != nil {
		return err
	}

	// Set end time to end of day
	endTime = endTime.Add(24*time.Hour - time.Second)

	// Convert to pgtype.Timestamptz
	startTimestamp := pgtype.Timestamptz{Time: startTime, Valid: true}
	endTimestamp := pgtype.Timestamptz{Time: endTime, Valid: true}

	histories, err := queries.GetMarinaUsageHistoryByDateRange(c.Request().Context(), db.GetMarinaUsageHistoryByDateRangeParams{
		MarinaID:    marinaID,
		CreatedAt:   startTimestamp,
		CreatedAt_2: endTimestamp,
	})
	if err != nil {
		return err
	}

	return c.JSON(200, histories)
}

// GetLatestMarinaUsageHistory handles retrieving the latest marina usage history record for a specific marina
func GetLatestMarinaUsageHistory(c echo.Context, queries *db.Queries) error {
	marinaIDStr := c.Param("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return err
	}

	history, err := queries.GetLatestMarinaUsageHistory(c.Request().Context(), marinaID)
	if err != nil {
		return err
	}

	return c.JSON(200, history)
}

// GetMarinaUsageHistoryByMonth handles retrieving marina usage history records for a specific month
func GetMarinaUsageHistoryByMonth(c echo.Context, queries *db.Queries) error {
	marinaIDStr := c.Param("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return err
	}

	month := c.QueryParam("month")
	monthTime, err := time.Parse("2006-01", month)
	if err != nil {
		return err
	}

	monthTimestamp := pgtype.Timestamp{Time: monthTime, Valid: true}

	history, err := queries.GetMarinaUsageHistoryByMonth(c.Request().Context(), db.GetMarinaUsageHistoryByMonthParams{
		MarinaID: marinaID,
		Column2:  monthTimestamp,
	})
	if err != nil {
		return err
	}

	return c.JSON(200, history)
}
