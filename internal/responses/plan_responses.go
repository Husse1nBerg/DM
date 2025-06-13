package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// NotesMessagesPlanResponse represents a notes and messages plan in the system
// @Description Notes and messages plan data including limits and pricing
type NotesMessagesPlanResponse struct {
	ID            uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name          string     `json:"name" example:"Basic Plan"`
	MonthlyPrice  float64    `json:"monthlyPrice" example:"29.99"`
	TextLimit     *string    `json:"textLimit,omitempty" example:"1000"`
	EmailLimit    *string    `json:"emailLimit,omitempty" example:"Unlimited"`
	UserLimit     *string    `json:"userLimit,omitempty" example:"Unlimited"`
	IsMostPopular *bool      `json:"isMostPopular" example:"false"`
	CreatedAt     time.Time  `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt     *time.Time `json:"updatedAt,omitempty" example:"2024-01-02T00:00:00Z"`
}

// StoragePlanResponse represents a storage plan in the system
// @Description Storage plan data including limits and pricing
type StoragePlanResponse struct {
	ID             uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name           string     `json:"name" example:"Basic Storage"`
	MonthlyPrice   float64    `json:"monthlyPrice" example:"19.99"`
	StorageLimitGB *string    `json:"storageLimitGB,omitempty" example:"10"`
	UserLimit      *string    `json:"userLimit,omitempty" example:"Unlimited"`
	IsMostPopular  *bool      `json:"isMostPopular" example:"false"`
	CreatedAt      time.Time  `json:"createdAt" example:"2024-01-01T00:00:00Z"`
	UpdatedAt      *time.Time `json:"updatedAt,omitempty" example:"2024-01-02T00:00:00Z"`
}

// ConvertNotesMessagesPlanToResponse converts a database notes and messages plan to a response model
func ConvertNotesMessagesPlanToResponse(plan db.NotesMessagesPlan) NotesMessagesPlanResponse {
	return NotesMessagesPlanResponse{
		ID:            plan.ID,
		Name:          plan.Name,
		MonthlyPrice:  plan.MonthlyPrice,
		TextLimit:     plan.TextLimit,
		EmailLimit:    plan.EmailLimit,
		UserLimit:     plan.UserLimit,
		IsMostPopular: plan.IsMostPopular,
		CreatedAt:     plan.CreatedAt.Time,
		UpdatedAt:     utils.PgTimeToTimePtr(plan.UpdatedAt),
	}
}

// ConvertStoragePlanToResponse converts a database storage plan to a response model
func ConvertStoragePlanToResponse(plan db.StoragePlan) StoragePlanResponse {
	return StoragePlanResponse{
		ID:             plan.ID,
		Name:           plan.Name,
		MonthlyPrice:   plan.MonthlyPrice,
		StorageLimitGB: plan.StorageLimitGb,
		UserLimit:      plan.UserLimit,
		IsMostPopular:  plan.IsMostPopular,
		CreatedAt:      plan.CreatedAt.Time,
		UpdatedAt:      utils.PgTimeToTimePtr(plan.UpdatedAt),
	}
}

// NewNotesMessagesPlanResponseSuccess creates a new successful notes and messages plan response
func NewNotesMessagesPlanResponseSuccess(plan db.NotesMessagesPlan) BaseResponse {
	return NewSuccessResponse(ConvertNotesMessagesPlanToResponse(plan))
}

// NewStoragePlanResponseSuccess creates a new successful storage plan response
func NewStoragePlanResponseSuccess(plan db.StoragePlan) BaseResponse {
	return NewSuccessResponse(ConvertStoragePlanToResponse(plan))
}

// NewNotesMessagesPlansPaginatedResponse creates a paginated response for notes and messages plans
func NewNotesMessagesPlansPaginatedResponse(plans []db.NotesMessagesPlan, total int64, perPage, currentPage int32) BaseResponse {
	planResponses := make([]NotesMessagesPlanResponse, len(plans))
	for i, plan := range plans {
		planResponses[i] = ConvertNotesMessagesPlanToResponse(plan)
	}
	return NewPaginatedResponse(planResponses, total, perPage, currentPage)
}

// NewStoragePlansPaginatedResponse creates a paginated response for storage plans
func NewStoragePlansPaginatedResponse(plans []db.StoragePlan, total int64, perPage, currentPage int32) BaseResponse {
	planResponses := make([]StoragePlanResponse, len(plans))
	for i, plan := range plans {
		planResponses[i] = ConvertStoragePlanToResponse(plan)
	}
	return NewPaginatedResponse(planResponses, total, perPage, currentPage)
}
