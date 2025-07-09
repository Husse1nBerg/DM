package responses

import (
	"encoding/json"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// CriteriaResponse represents a single criteria response
// @Description Criteria response model
// @Schema responses.CriteriaResponse
type CriteriaResponse struct {
	ID          uuid.UUID   `json:"id" example:"123e4567-e89b-12d3-a456-426614174000"`
	MarinaID    uuid.UUID   `json:"marinaId" example:"123e4567-e89b-12d3-a456-426614174000"`
	Name        string      `json:"name" example:"Customer Filter"`
	Description *string     `json:"description,omitempty" example:"Filter for active customers"`
	Criteria    interface{} `json:"criteria"`
	CreatedAt   *time.Time  `json:"createdAt,omitempty" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   *time.Time  `json:"updatedAt,omitempty" example:"2023-01-01T00:00:00Z"`
}

// CriteriaListResponse represents a list of criteria response
// @Description Criteria list response model
// @Schema responses.CriteriaListResponse
type CriteriaListResponse struct {
	BaseResponse
	Data []CriteriaResponse `json:"data"`
}

// CriteriaPaginatedResponse represents a paginated list of criteria response
// @Description Criteria paginated response model
// @Schema responses.CriteriaPaginatedResponse
type CriteriaPaginatedResponse struct {
	BaseResponse
	Data []CriteriaResponse `json:"data"`
}

// ConvertCriteriaToResponse converts a database criteria to a response model
func ConvertCriteriaToResponse(criteria db.Criterium) CriteriaResponse {
	var criteriaObj interface{}
	if len(criteria.Criteria) > 0 {
		if err := json.Unmarshal(criteria.Criteria, &criteriaObj); err != nil {
			// If unmarshal fails, return the raw bytes as string
			criteriaObj = string(criteria.Criteria)
		}
	}

	return CriteriaResponse{
		ID:          criteria.ID,
		MarinaID:    criteria.MarinaID,
		Name:        criteria.Name,
		Description: criteria.Description,
		Criteria:    criteriaObj,
		CreatedAt:   utils.PgTimeToTimePtr(criteria.CreatedAt),
		UpdatedAt:   utils.PgTimeToTimePtr(criteria.UpdatedAt),
	}
}

// NewCriteriaResponse creates a new criteria response
func NewCriteriaResponse(criteria db.Criterium) BaseResponse {
	return NewSuccessResponse(ConvertCriteriaToResponse(criteria))
}

// NewCriteriaListResponse creates a new criteria list response
func NewCriteriaListResponse(criteria []db.Criterium) BaseResponse {
	criteriaResponses := make([]CriteriaResponse, len(criteria))
	for i, c := range criteria {
		criteriaResponses[i] = ConvertCriteriaToResponse(c)
	}

	return NewSuccessResponse(criteriaResponses)
}

// NewCriteriaPaginatedResponse creates a new criteria paginated response
func NewCriteriaPaginatedResponse(criteria []db.Criterium, total int64, perPage, currentPage int32) BaseResponse {
	criteriaResponses := make([]CriteriaResponse, len(criteria))
	for i, c := range criteria {
		criteriaResponses[i] = ConvertCriteriaToResponse(c)
	}

	return NewPaginatedResponse(criteriaResponses, total, perPage, currentPage)
}
