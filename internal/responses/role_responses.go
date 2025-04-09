package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/models"
	"github.com/google/uuid"
)

// RoleResponse represents the role data returned in API responses
// @Description Role representation for API responses
type RoleResponse struct {
	ID          uuid.UUID           `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Name        string              `json:"name" example:"Admin"`
	Description *string             `json:"description,omitempty" example:"Administrator role with full access"`
	Permissions *models.Permissions `json:"permissions"`
	IsActive    bool                `json:"isActive" example:"true"`
	CreatedAt   time.Time           `json:"createdAt" example:"2023-01-01T00:00:00Z"`
	UpdatedAt   *time.Time          `json:"updatedAt,omitempty" example:"2023-01-02T00:00:00Z"`
}

// NewRolesPaginatedResponse creates a paginated response for roles
func NewRolesPaginatedResponse(roles []db.Role, total int64, perPage, currentPage int32) BaseResponse {
	roleResponses := make([]RoleResponse, 0, len(roles))
	for _, role := range roles {
		roleResponses = append(roleResponses, convertDBRoleToResponse(role))
	}

	// Return roles directly as the data array instead of nesting them under a 'roles' key
	return NewPaginatedResponse(roleResponses, total, perPage, currentPage)
}

// NewRoleResponseSuccess creates a success response with role data
func NewRoleResponseSuccess(role db.Role) BaseResponse {
	roleResponse := convertDBRoleToResponse(role)
	// Return the role directly instead of wrapping it in a 'role' object
	return NewSuccessResponse(roleResponse)
}

// convertDBRoleToResponse converts a DB role to a response object
func convertDBRoleToResponse(role db.Role) RoleResponse {
	// Parse JSON permissions from the DB
	permissions := &models.Permissions{}
	if role.Permissions != nil {
		if err := permissions.FromBytes(role.Permissions); err != nil {
			// Default empty permissions if can't parse
			permissions = &models.Permissions{}
		}
	}

	response := RoleResponse{
		ID:          role.ID,
		Name:        role.Name,
		Description: role.Description,
		Permissions: permissions,
		IsActive:    true, // Default to true if nil
		CreatedAt:   role.CreatedAt.Time,
	}

	if role.IsActive != nil {
		response.IsActive = *role.IsActive
	}

	if role.UpdatedAt.Valid {
		updatedAt := role.UpdatedAt.Time
		response.UpdatedAt = &updatedAt
	}

	return response
}
