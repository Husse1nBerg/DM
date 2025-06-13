package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// MarinaUsageHistoryResponse represents a marina usage history record
// @Description Marina usage history data including storage, email, and text usage
type MarinaUsageHistoryResponse struct {
	ID           uuid.UUID            `json:"id" example:"550e8400-e29b-41d4-a716-446655440001"`
	MarinaID     uuid.UUID            `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440001"`
	StorageUsage utils.StorageUsageGB `json:"storageUsage" example:"0.00"`
	EmailUsage   int16                `json:"emailUsage" example:"0"`
	TextUsage    int16                `json:"textUsage" example:"0"`
	CreatedAt    time.Time            `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt    time.Time            `json:"updatedAt" example:"2024-01-01T00:00:00Z"`
}

// ConvertMarinaUsageHistoryToResponse converts a database marina usage history to a response model
func ConvertMarinaUsageHistoryToResponse(history db.MarinaUsageHistory) MarinaUsageHistoryResponse {
	// Convert storage usage from bytes to GB
	const bytesInGB = 1024 * 1024 * 1024 // 1 GB in bytes
	storageUsageGB := utils.StorageUsageGB(float64(history.StorageUsage) / float64(bytesInGB))

	return MarinaUsageHistoryResponse{
		ID:           history.ID,
		MarinaID:     history.MarinaID,
		StorageUsage: storageUsageGB,
		EmailUsage:   history.EmailUsage,
		TextUsage:    history.TextUsage,
		CreatedAt:    history.CreatedAt.Time,
		UpdatedAt:    history.UpdatedAt.Time,
	}
}

// NewMarinaUsageHistoryResponseSuccess creates a new successful marina usage history response
func NewMarinaUsageHistoryResponseSuccess(history db.MarinaUsageHistory) BaseResponse {
	return NewSuccessResponse(ConvertMarinaUsageHistoryToResponse(history))
}

// NewMarinaUsageHistoryListResponse creates a response for a list of marina usage history records
func NewMarinaUsageHistoryListResponse(histories []db.MarinaUsageHistory) BaseResponse {
	historyResponses := make([]MarinaUsageHistoryResponse, len(histories))
	for i, history := range histories {
		historyResponses[i] = ConvertMarinaUsageHistoryToResponse(history)
	}

	return NewSuccessResponse(historyResponses)
}

// NewMarinaUsageHistoryPaginatedResponse creates a paginated response of marina usage history records
func NewMarinaUsageHistoryPaginatedResponse(histories []db.MarinaUsageHistory, total int64, perPage, page int32) BaseResponse {
	historyResponses := make([]MarinaUsageHistoryResponse, len(histories))
	for i, history := range histories {
		historyResponses[i] = ConvertMarinaUsageHistoryToResponse(history)
	}

	return NewPaginatedResponse(historyResponses, total, perPage, page)
}

// MarinaUsageHistoryListResponse is purely for Swagger documentation
type MarinaUsageHistoryListResponse struct {
	Data        []MarinaUsageHistoryResponse `json:"data"`
	Total       int64                        `json:"total" example:"42"`
	PerPage     int32                        `json:"perPage" example:"10"`
	CurrentPage int32                        `json:"currentPage" example:"1"`
	LastPage    int32                        `json:"lastPage" example:"5"`
}

// MarinaUsageHistoryResponseWrapper is purely for Swagger documentation
type MarinaUsageHistoryResponseWrapper struct {
	Data MarinaUsageHistoryResponse `json:"data"`
}
