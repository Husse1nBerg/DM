package responses

import (
	"context"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
)

// UserResponse represents a user profile in the system
// @Description User profile data including personal information and system roles
type UserResponse struct {
	ID                  uuid.UUID  `json:"id" example:"550e8400-e29b-41d4-a716-446655440000"`
	Username            string     `json:"username" example:"johndoe"`
	FirstName           string     `json:"firstName" example:"John"`
	LastName            string     `json:"lastName" example:"Doe"`
	Email               string     `json:"email" example:"john.doe@example.com"`
	EmailVerified       *time.Time `json:"emailVerified,omitempty"`
	Phone               *string    `json:"phone,omitempty" example:"+15551234567"`
	Title               *string    `json:"title,omitempty" example:"Manager"`
	Image               *string    `json:"image,omitempty" example:"/images/profiles/johndoe.jpg"`
	LastLogin           *time.Time `json:"lastLogin,omitempty"`
	FailedLoginAttempts *int32     `json:"failedLoginAttempts,omitempty" example:"0"`
	LockedUntil         *time.Time `json:"lockedUntil,omitempty"`
	LastPasswordReset   *time.Time `json:"lastPasswordReset,omitempty"`
	OrganizationID      uuid.UUID  `json:"organizationId" example:"550e8400-e29b-41d4-a716-446655440001"`
	MarinaID            uuid.UUID  `json:"marinaId" example:"550e8400-e29b-41d4-a716-446655440002"`
	RoleID              uuid.UUID  `json:"roleId" example:"550e8400-e29b-41d4-a716-446655440003"`
	RoleName            *string    `json:"roleName,omitempty" example:"Admin"`
	CustomerID          *string    `json:"customerId,omitempty" example:"1234567890"`
	CustomerName        *string    `json:"customerName,omitempty" example:"John's Marina"`
	IsCustomer          *bool      `json:"isCustomer,omitempty" example:"false"`
	IsSuperuser         *bool      `json:"isSuperuser,omitempty" example:"false"`
	IsActive            *bool      `json:"isActive,omitempty" example:"true"`
	UserAnalytics       *bool      `json:"userAnalytics,omitempty" example:"true"`
	CreatedAt           *time.Time `json:"createdAt,omitempty"`
	UpdatedAt           *time.Time `json:"updatedAt,omitempty"`
}

// UserWithRole embeds User and adds RoleName field
type UserWithRole struct {
	db.User
	RoleName *string `json:"roleName,omitempty" example:"Admin"`
}

func NewUserResponse(user db.User) UserResponse {
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
		CustomerID:          user.CustomerID,
		IsCustomer:          user.IsCustomer,
		IsSuperuser:         user.IsSuperuser,
		IsActive:            user.IsActive,
		UserAnalytics:       user.UserAnalytics,
		CreatedAt:           utils.PgTimeToTimePtr(user.CreatedAt),
		UpdatedAt:           utils.PgTimeToTimePtr(user.UpdatedAt),
	}
}

func NewUserResponseFromRow(r db.GetUsersByMarinaPaginatedRow, server *server.Server) *UserResponse {

	response := &UserResponse{
		ID:                  r.ID,
		Username:            r.Username,
		FirstName:           r.FirstName,
		LastName:            r.LastName,
		Email:               r.Email,
		EmailVerified:       utils.PgTimeToTimePtr(r.EmailVerified),
		Phone:               r.Phone,
		Title:               r.Title,
		Image:               utils.GetFullImageURL(r.Image),
		LastLogin:           utils.PgTimeToTimePtr(r.LastLogin),
		FailedLoginAttempts: r.FailedLoginAttempts,
		LockedUntil:         utils.PgTimeToTimePtr(r.LockedUntil),
		LastPasswordReset:   utils.PgTimeToTimePtr(r.LastPasswordReset),
		OrganizationID:      r.OrganizationID,
		MarinaID:            r.MarinaID,
		RoleID:              r.RoleID,
		IsSuperuser:         r.IsSuperuser,
		IsActive:            r.IsActive,
		UserAnalytics:       r.UserAnalytics,
		CreatedAt:           utils.PgTimeToTimePtr(r.CreatedAt),
		UpdatedAt:           utils.PgTimeToTimePtr(r.UpdatedAt),
		CustomerID:          r.CustomerID,
		IsCustomer:          r.IsCustomer,
		RoleName:            r.RoleName,
	}

	// If CustomerID is present, get the customer name from DME
	if r.CustomerID != nil && server != nil {
		// Get the marina to get the system ID
		marina, err := server.DB.Queries().GetMarinaByID(context.Background(), r.MarinaID)
		if err == nil && marina.SystemID != nil {
			// Get customer name from DME
			customer, err := server.DME.CustomerRetrieve(context.Background(), *r.CustomerID, r.OrganizationID, *marina.SystemID)
			if err == nil {
				response.CustomerName = &customer.Name
			}
		}
	}

	return response
}

func NewUserResponseFromMarinaListRow(r db.GetMarinaUsersListPaginatedRow, server *server.Server) *UserResponse {

	response := &UserResponse{
		ID:                  r.ID,
		Username:            r.Username,
		FirstName:           r.FirstName,
		LastName:            r.LastName,
		Email:               r.Email,
		EmailVerified:       utils.PgTimeToTimePtr(r.EmailVerified),
		Phone:               r.Phone,
		Title:               r.Title,
		Image:               utils.GetFullImageURL(r.Image),
		LastLogin:           utils.PgTimeToTimePtr(r.LastLogin),
		FailedLoginAttempts: r.FailedLoginAttempts,
		LockedUntil:         utils.PgTimeToTimePtr(r.LockedUntil),
		LastPasswordReset:   utils.PgTimeToTimePtr(r.LastPasswordReset),
		OrganizationID:      r.OrganizationID,
		MarinaID:            r.MarinaID,
		RoleID:              r.RoleID,
		IsSuperuser:         r.IsSuperuser,
		IsActive:            r.IsActive,
		UserAnalytics:       r.UserAnalytics,
		CreatedAt:           utils.PgTimeToTimePtr(r.CreatedAt),
		UpdatedAt:           utils.PgTimeToTimePtr(r.UpdatedAt),
		CustomerID:          r.CustomerID,
		IsCustomer:          r.IsCustomer,
		RoleName:            r.RoleName,
	}

	// If CustomerID is present, get the customer name from DME
	if r.CustomerID != nil && server != nil {
		// Get the marina to get the system ID
		marina, err := server.DB.Queries().GetMarinaByID(context.Background(), r.MarinaID)
		if err == nil && marina.SystemID != nil {
			// Get customer name from DME
			customer, err := server.DME.CustomerRetrieve(context.Background(), *r.CustomerID, r.OrganizationID, *marina.SystemID)
			if err == nil {
				response.CustomerName = &customer.Name
			}
		}
	}

	return response
}

func NewUserResponseFromUserMarinasAssignmentRow(r db.ListUserMarinasAssignmentsPaginatedAscRow, server *server.Server) *UserResponse {
	response := &UserResponse{
		ID:                  r.ID,
		Username:            r.Username,
		FirstName:           r.FirstName,
		LastName:            r.LastName,
		Email:               r.Email,
		EmailVerified:       utils.PgTimeToTimePtr(r.EmailVerified),
		Phone:               r.Phone,
		Title:               r.Title,
		Image:               utils.GetFullImageURL(r.Image),
		LastLogin:           utils.PgTimeToTimePtr(r.LastLogin),
		FailedLoginAttempts: r.FailedLoginAttempts,
		LockedUntil:         utils.PgTimeToTimePtr(r.LockedUntil),
		LastPasswordReset:   utils.PgTimeToTimePtr(r.LastPasswordReset),
		OrganizationID:      r.OrganizationID,
		MarinaID:            r.MarinaID_2,
		RoleID:              r.RoleID,
		IsSuperuser:         r.IsSuperuser,
		IsActive:            r.IsActive,
		UserAnalytics:       r.UserAnalytics,
		CreatedAt:           utils.PgTimeToTimePtr(r.CreatedAt),
		UpdatedAt:           utils.PgTimeToTimePtr(r.UpdatedAt),
		CustomerID:          r.CustomerID_2,
		IsCustomer:          r.IsCustomer,
		RoleName:            r.RoleName,
	}

	// If CustomerID is present, get the customer name from DME
	if r.CustomerID_2 != nil && server != nil {
		marina, err := server.DB.Queries().GetMarinaByID(context.Background(), r.MarinaID_2)
		if err == nil && marina.SystemID != nil {
			customer, err := server.DME.CustomerRetrieve(context.Background(), *r.CustomerID_2, r.OrganizationID, *marina.SystemID)
			if err == nil {
				response.CustomerName = &customer.Name
			}
		}
	}

	return response
}
func NewUserResponseFromUserMarinasAssignmentRowAdmin(r db.ListUserMarinasAssignmentsPaginatedAdminRow, server *server.Server) *UserResponse {
	response := &UserResponse{
		ID:                  r.ID,
		Username:            r.Username,
		FirstName:           r.FirstName,
		LastName:            r.LastName,
		Email:               r.Email,
		EmailVerified:       utils.PgTimeToTimePtr(r.EmailVerified),
		Phone:               r.Phone,
		Title:               r.Title,
		Image:               utils.GetFullImageURL(r.Image),
		LastLogin:           utils.PgTimeToTimePtr(r.LastLogin),
		FailedLoginAttempts: r.FailedLoginAttempts,
		LockedUntil:         utils.PgTimeToTimePtr(r.LockedUntil),
		LastPasswordReset:   utils.PgTimeToTimePtr(r.LastPasswordReset),
		OrganizationID:      r.OrganizationID,
		MarinaID:            r.MarinaID_2,
		RoleID:              r.RoleID,
		IsSuperuser:         r.IsSuperuser,
		IsActive:            r.IsActive,
		UserAnalytics:       r.UserAnalytics,
		CreatedAt:           utils.PgTimeToTimePtr(r.CreatedAt),
		UpdatedAt:           utils.PgTimeToTimePtr(r.UpdatedAt),
		CustomerID:          r.CustomerID_2,
		IsCustomer:          r.IsCustomer,
		RoleName:            r.RoleName,
	}

	// If CustomerID is present, get the customer name from DME
	if r.CustomerID_2 != nil && server != nil {
		marina, err := server.DB.Queries().GetMarinaByID(context.Background(), r.MarinaID_2)
		if err == nil && marina.SystemID != nil {
			customer, err := server.DME.CustomerRetrieve(context.Background(), *r.CustomerID_2, r.OrganizationID, *marina.SystemID)
			if err == nil {
				response.CustomerName = &customer.Name
			}
		}
	}

	return response
}

func NewUserResponseFromUserMarinasAssignmentRowAdminOnly(r db.ListUserMarinasAssignmentsPaginatedAdminOnlyRow, server *server.Server) *UserResponse {
	response := &UserResponse{
		ID:                  r.ID,
		Username:            r.Username,
		FirstName:           r.FirstName,
		LastName:            r.LastName,
		Email:               r.Email,
		EmailVerified:       utils.PgTimeToTimePtr(r.EmailVerified),
		Phone:               r.Phone,
		Title:               r.Title,
		Image:               utils.GetFullImageURL(r.Image),
		LastLogin:           utils.PgTimeToTimePtr(r.LastLogin),
		FailedLoginAttempts: r.FailedLoginAttempts,
		LockedUntil:         utils.PgTimeToTimePtr(r.LockedUntil),
		LastPasswordReset:   utils.PgTimeToTimePtr(r.LastPasswordReset),
		OrganizationID:      r.OrganizationID,
		MarinaID:            r.MarinaID_2,
		RoleID:              r.RoleID,
		IsSuperuser:         r.IsSuperuser,
		IsActive:            r.IsActive,
		UserAnalytics:       r.UserAnalytics,
		CreatedAt:           utils.PgTimeToTimePtr(r.CreatedAt),
		UpdatedAt:           utils.PgTimeToTimePtr(r.UpdatedAt),
		CustomerID:          r.CustomerID_2,
		IsCustomer:          r.IsCustomer,
		RoleName:            r.RoleName,
	}

	return response
}

func NewUsersPaginatedResponseFromRows(users []db.GetUsersByMarinaPaginatedRow, total int64, perPage, page int32) BaseResponse {
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		response := NewUserResponseFromRow(user, nil)
		if response != nil {
			userResponses[i] = *response
		}
	}
	return NewPaginatedResponse(userResponses, total, perPage, page)
}

func NewUsersPaginatedResponseFromMarinaRows(users []db.GetMarinaUsersListPaginatedRow, total int64, perPage, page int32) BaseResponse {
	userResponses := make([]UserResponse, len(users))
	for i, user := range users {
		response := NewUserResponseFromMarinaListRow(user, nil)
		if response != nil {
			userResponses[i] = *response
		}
	}
	return NewPaginatedResponse(userResponses, total, perPage, page)
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
