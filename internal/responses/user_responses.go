package responses

import (
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/models"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// UserResponse represents a user profile in the system
// @Description User profile data including personal information and system roles
type UserResponse struct {
	ID                  uuid.UUID           `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username            string              `json:"username" example:"johndoe"`
	FirstName           string              `json:"firstName" example:"John"`
	LastName            string              `json:"lastName" example:"Doe"`
	Email               string              `json:"email" example:"john.doe@example.com"`
	EmailVerified       *time.Time          `json:"emailVerified,omitempty"`
	Phone               *string             `json:"phone,omitempty" example:"+15551234567"`
	Title               *string             `json:"title,omitempty" example:"Manager"`
	Image               *string             `json:"image,omitempty" example:"/images/profiles/johndoe.jpg"`
	LastLogin           *time.Time          `json:"lastLogin,omitempty"`
	FailedLoginAttempts *int32              `json:"failedLoginAttempts,omitempty" example:"0"`
	LockedUntil         *time.Time          `json:"lockedUntil,omitempty"`
	LastPasswordReset   *time.Time          `json:"lastPasswordReset,omitempty"`
	OrganizationID      uuid.UUID           `json:"organizationId" example:"550e8400-e29b-41d4-a716-446655440001"`
	MarinaID            uuid.UUID           `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440002"`
	RoleID              uuid.UUID           `json:"roleId" example:"550e8400-e29b-41d4-a716-446655440003"`
	RoleName            *string             `json:"roleName,omitempty" example:"Admin"`
	CustomerID          *string             `json:"customerId,omitempty" example:"1234567890"`
	IsCustomer          *bool               `json:"isCustomer,omitempty" example:"false"`
	IsSuperuser         *bool               `json:"isSuperuser,omitempty" example:"false"`
	IsActive            *bool               `json:"isActive,omitempty" example:"true"`
	CreatedAt           *time.Time          `json:"createdAt,omitempty"`
	UpdatedAt           *time.Time          `json:"updatedAt,omitempty"`
	Permissions         *models.Permissions `json:"permissions,omitempty"`
	Modules             *models.Modules     `json:"modules,omitempty"`
}

func NewUserResponse(user db.User) UserResponse {
	// Create instances to fill from DB byte arrays
	var permissions models.Permissions
	var modules models.Modules

	// Convert byte arrays to structs
	if user.Permissions != nil {
		if err := permissions.FromBytes(user.Permissions); err != nil {
			// Handle error or set to nil (using default zero values is fine)
		}
	}

	if user.Modules != nil {
		if err := modules.FromBytes(user.Modules); err != nil {
			// Handle error or set to nil (using default zero values is fine)
		}
	}

	return UserResponse{
		ID:                  user.ID,
		Username:            user.Username,
		FirstName:           user.FirstName,
		LastName:            user.LastName,
		Email:               user.Email,
		EmailVerified:       utils.PgTimeToTimePtr(user.EmailVerified),
		Phone:               user.Phone,
		Title:               user.Title,
		Image:               utils.GetFullImageURL(user.Image),
		LastLogin:           utils.PgTimeToTimePtr(user.LastLogin),
		FailedLoginAttempts: user.FailedLoginAttempts,
		LockedUntil:         utils.PgTimeToTimePtr(user.LockedUntil),
		LastPasswordReset:   utils.PgTimeToTimePtr(user.LastPasswordReset),
		OrganizationID:      user.OrganizationID,
		MarinaID:            user.MarinaID,
		RoleID:              user.RoleID,
		RoleName:            user.RoleName,
		CustomerID:          user.CustomerID,
		IsCustomer:          user.IsCustomer,
		IsSuperuser:         user.IsSuperuser,
		IsActive:            user.IsActive,
		CreatedAt:           utils.PgTimeToTimePtr(user.CreatedAt),
		UpdatedAt:           utils.PgTimeToTimePtr(user.UpdatedAt),
		Permissions:         &permissions,
		Modules:             &modules,
	}
}

func NewUserResponseSuccess(user db.User) BaseResponse {
	return NewSuccessResponse(NewUserResponse(user))
}

func NewUsersPaginatedResponse(users []db.User, total int64, perPage, page int32) BaseResponse {
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		userResponses[i] = NewUserResponse(user)
	}

	return NewPaginatedResponse(userResponses, total, perPage, page)
}

// UserListResponse is purely for Swagger documentation
type UserListResponse struct {
	Data        []UserResponse `json:"data"`
	Total       int64          `json:"total" example:"42"`
	PerPage     int32          `json:"perPage" example:"10"`
	CurrentPage int32          `json:"currentPage" example:"1"`
	LastPage    int32          `json:"lastPage" example:"5"`
}

// UserResponseWrapper is purely for Swagger documentation
type UserResponseWrapper struct {
	Data    UserResponse `json:"data"`
	Message string       `json:"message,omitempty" example:"User profile retrieved successfully"`
}
