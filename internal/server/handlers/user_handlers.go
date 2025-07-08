package handlers

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/notifications"
	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	server              *s.Server
	notificationService *notifications.NotificationService
}

func NewUserHandler(server *s.Server) *UserHandler {
	// Initialize notification service
	notificationService := notifications.NewNotificationService(
		server.DB.Queries(),
		server.Redis,
		server.Logger,
	)

	return &UserHandler{
		server:              server,
		notificationService: notificationService,
	}
}

// ListUsersHandler lists all existing users
//
//	@Summary		List users
//	@Description	get users
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Success		200	{object}	responses.UserListResponse "Paginated list of users"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/list [get]
func (g *UserHandler) ListUsersHandler(c echo.Context) error {
	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()

	// Get paginated users
	params := db.GetAllUsersPaginatedParams{
		Limit:  pagination.PageSize,
		Offset: (pagination.Page - 1) * pagination.PageSize,
	}
	users, err := queries.GetAllUsersPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Get total count for pagination
	allUsers, err := queries.GetAllUsers(c.Request().Context())
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allUsers))

	return responses.NewUsersPaginatedResponse(users, total, pagination.PageSize, pagination.Page).JSON(c)
}

// CreateUserHandler creates a new user
//
//	@Summary		Create user
//	@Description	Create a new user
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			user	body		requests.CreateUserRequest	true	"User information"
//	@Success		201		{object}	responses.UserResponseWrapper "Created user"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user [post]
func (g *UserHandler) CreateUserHandler(c echo.Context) error {
	// Parse and validate the request body
	req := new(requests.CreateUserRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Create the user
	queries := g.server.DB.Queries()

	// Check if the role exists
	_, err := queries.GetRoleByID(c.Request().Context(), req.RoleID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Role not found").JSON(c)
	}

	// Check if the marina and organization exist and linked
	marina, err := queries.GetMarinaByID(c.Request().Context(), req.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Marina not found").JSON(c)
	}

	organization, err := queries.GetOrganizationByID(c.Request().Context(), req.OrganizationID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Organization not found").JSON(c)
	}

	if marina.OrganizationID != organization.ID {
		return responses.NewErrorResponse(http.StatusBadRequest, "Marina and organization are not linked").JSON(c)
	}

	role, err := queries.GetRoleByID(c.Request().Context(), req.RoleID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Role not found").JSON(c)
	}

	email := utils.LowerCase(req.Email)
	// Check if the email is already taken
	userByEmail, err := queries.GetUserByEmail(c.Request().Context(), email)
	if err == nil {
		// Email exists, check if user is already assigned to this marina
		canAccess, err := queries.UserCanAccessMarina(c.Request().Context(), db.UserCanAccessMarinaParams{
			UserID:   userByEmail.ID,
			MarinaID: req.MarinaID,
		})
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		if canAccess {
			return responses.NewErrorResponse(http.StatusBadRequest, "Email already taken for this marina").JSON(c)
		}
		// Assign the existing user to the marina with CustomerID
		assignUserToMarina := db.AssignUserToMarinaParams{
			UserID:   userByEmail.ID,
			MarinaID: req.MarinaID,
		}
		err = queries.AssignUserToMarina(c.Request().Context(), assignUserToMarina)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		// Upsert customer settings as in the original logic
		_, err = queries.UpsertCustomerSettings(c.Request().Context(), db.UpsertCustomerSettingsParams{
			MarinaID:     req.MarinaID,
			EnablePortal: utils.Pointer(true),
		})
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		return responses.NewMessageResponse(http.StatusOK, "User assigned to marina").JSON(c)
	}
	username := req.Username
	if username == "" {
		username = utils.GenerateUsername(req.FirstName)
	}

	// Check if the username is already taken
	_, err = queries.GetUserByUsername(c.Request().Context(), username)
	if err == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Username already taken").JSON(c)
	}

	// Create the user
	// Hash the password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password").JSON(c)
	}

	// Set default values for nullable fields if not provided
	failedLoginAttempts := int32(0)
	isActive := true
	isSuperuser := false
	if role.Type == "internal" {
		isSuperuser = true
	}
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	params := db.CreateUserParams{
		Username:            username,
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		Email:               email,
		Phone:               req.Phone,
		Title:               req.Title,
		Image:               req.Image,
		PasswordHash:        utils.Pointer(passwordHash),
		FailedLoginAttempts: &failedLoginAttempts,
		LastPasswordReset:   utils.PgTimeNow(),
		OrganizationID:      req.OrganizationID,
		MarinaID:            req.MarinaID,
		RoleID:              req.RoleID,
		IsSuperuser:         &isSuperuser,
		IsActive:            &isActive,
		UserAnalytics:       utils.Pointer(true),
	}

	user, err := queries.CreateUser(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	assignUserToMarina := db.AssignUserToMarinaParams{
		UserID:   user.ID,
		MarinaID: req.MarinaID,
	}

	err = queries.AssignUserToMarina(c.Request().Context(), assignUserToMarina)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	response := responses.NewUserResponseSuccess(user)
	return c.JSON(http.StatusCreated, response)
}

// GetUserHandler gets a specific user by ID
//
//	@Summary		Get user
//	@Description	Get a specific user by ID
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			userId	path		string	true	"User ID"
//	@Success		200		{object}	responses.UserResponseWrapper "User details"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		404		{object}	responses.Error "User not found"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/{userId} [get]
func (g *UserHandler) GetUserHandler(c echo.Context) error {
	// Parse and validate user ID from the path
	param := new(requests.UserIDParam)
	if err := c.Bind(param); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(param); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := g.server.DB.Queries()
	user, err := queries.GetUserByID(c.Request().Context(), param.UserID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, err).JSON(c)
	}

	response := responses.NewUserResponseSuccess(user)
	response.Pretty = true
	return response.JSON(c)
}

// Get My User
//
//	@Summary		Get my user
//	@Description	get my user
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	responses.UserResponseWrapper "Current user's profile"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/profile [get]
func (g *UserHandler) GetMyUserHandler(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	id := claims.ID
	queries := g.server.DB.Queries()
	user, err := queries.GetUserByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.NewUserResponseSuccess(user)
	response.Pretty = true
	return response.JSON(c)
}

// DeleteUserHandler soft-deletes a user
//
//	@Summary		Delete user
//	@Description	Soft-delete a user (mark as deleted)
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			userId	path		string	true	"User ID"
//	@Success		204		{object}	nil "No content"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/{userId} [delete]
func (g *UserHandler) DeleteUserHandler(c echo.Context) error {
	// Parse and validate user ID from the path
	param := new(requests.UserIDParam)
	if err := c.Bind(param); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(param); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := g.server.DB.Queries()
	err := queries.SoftDeleteUser(c.Request().Context(), param.UserID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return c.NoContent(http.StatusNoContent)
}

// GetUsersByRoleHandler gets users by role ID
//
//	@Summary		Get users by role
//	@Description	Get all users with a specific role
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			roleId	path		string	true	"Role ID"
//	@Param			page	query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Success		200		{object}	responses.UserListResponse "List of users with the specified role"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/role/{roleId} [get]
func (g *UserHandler) GetUsersByRoleHandler(c echo.Context) error {
	roleIDStr := c.Param("roleId")
	roleID, err := uuid.Parse(roleIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()

	// Get paginated users by role
	params := db.GetUsersByRolePaginatedParams{
		RoleID: roleID,
		Limit:  pagination.PageSize,
		Offset: (pagination.Page - 1) * pagination.PageSize,
	}
	users, err := queries.GetUsersByRolePaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Get total count for pagination
	allUsers, err := queries.GetUsersByRole(c.Request().Context(), roleID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allUsers))

	return responses.NewUsersPaginatedResponse(users, total, pagination.PageSize, pagination.Page).JSON(c)
}

// GetUsersByOrganizationHandler gets users by organization ID
//
//	@Summary		Get users by organization
//	@Description	Get all users in a specific organization
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"
//	@Param			page				query		int		false	"Page number"	default(1)
//	@Param			pageSize			query		int		false	"Page size"		default(10)
//	@Success		200				{object}	responses.UserListResponse "Paginated list of users in the organization"
//	@Failure		400				{object}	responses.Error "Bad request"
//	@Failure		500				{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/organization/{organizationId} [get]
func (g *UserHandler) GetUsersByOrganizationHandler(c echo.Context) error {
	// Parse organization ID
	orgIDStr := c.Param("organizationId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()

	// Get paginated users by organization
	params := db.GetUsersByOrganizationPaginatedParams{
		OrganizationID: orgID,
		Limit:          pagination.PageSize,
		Offset:         (pagination.Page - 1) * pagination.PageSize,
	}
	users, err := queries.GetUsersByOrganizationPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Get total count for pagination
	// In a real implementation, you would have a separate query to get the total count
	// For simplicity, we'll use the same unpaginated query to get the total
	allUsers, err := queries.GetUsersByOrganization(c.Request().Context(), orgID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allUsers))

	return responses.NewUsersPaginatedResponse(users, total, pagination.PageSize, pagination.Page).JSON(c)
}

// GetUsersByMarinaHandler gets users by marina ID
//
//	@Summary		Get users by marina
//	@Description	Get all users in a specific marina
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	path		string	true	"Marina ID"
//	@Param			isCustomer	query		bool	false	"Filter by customer status. If not provided, returns all users"	default()
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Success		200			{object}	responses.UserListResponse "Paginated list of users in the marina"
//	@Failure		400			{object}	responses.Error "Bad request"
//	@Failure		500			{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/marina/{marinaId} [get]
func (g *UserHandler) GetUsersByMarinaHandler(c echo.Context) error {
	// Parse marina ID
	marinaIDStr := c.Param("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Parse isCustomer parameter
	isCustomerStr := c.QueryParam("isCustomer")
	var isCustomer *bool
	if isCustomerStr != "" {
		value := isCustomerStr == "true"
		isCustomer = &value
	}

	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()

	// Use the new ListUserMarinasAssignmentsPaginated query
	var isCustomerVal bool
	if isCustomer != nil {
		isCustomerVal = *isCustomer
	}
	params := db.ListUserMarinasAssignmentsPaginatedParams{
		MarinaID: marinaID,
		Column2:  isCustomerVal,
		Limit:    pagination.PageSize,
		Offset:   (pagination.Page - 1) * pagination.PageSize,
	}
	rows, err := queries.ListUserMarinasAssignmentsPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Map to response DTOs
	userResponses := make([]responses.UserResponse, len(rows))
	for i, row := range rows {
		response := responses.NewUserResponseFromUserMarinasAssignmentRow(row, g.server)
		if response != nil {
			userResponses[i] = *response
		}
	}

	total := int64(len(rows))

	return responses.NewPaginatedResponse(userResponses, total, pagination.PageSize, pagination.Page).JSON(c)
}

// GetMarinaUsersList gets all users associated with a marina through user_marinas
//
//	@Summary		Get marina users
//	@Description	Get users assigned to a marina (through user_marinas table)
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	path		string	true	"Marina ID"
//	@Param			isCustomer	query		bool	false	"Filter by customer status. If not provided, returns all users"	default()
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Success		200			{object}	responses.UserListResponse "List of users assigned to the marina"
//	@Failure		400			{object}	responses.Error "Bad request"
//	@Failure		500			{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/marina/{marinaId}/assigned [get]
func (g *UserHandler) GetMarinaUsersList(c echo.Context) error {
	// Parse marina ID
	marinaIDStr := c.Param("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Parse isCustomer parameter
	isCustomerStr := c.QueryParam("isCustomer")
	var isCustomer *bool
	if isCustomerStr != "" {
		value := isCustomerStr == "true"
		isCustomer = &value
	}

	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()

	// Get paginated marina users list
	params := db.GetMarinaUsersListPaginatedParams{
		MarinaID:   marinaID,
		IsCustomer: isCustomer,
		Limit:      pagination.PageSize,
		Offset:     (pagination.Page - 1) * pagination.PageSize,
	}
	userRows, err := queries.GetMarinaUsersListPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Get total count for pagination
	allUsers, err := queries.GetMarinaUsersList(c.Request().Context(), db.GetMarinaUsersListParams{
		MarinaID:   marinaID,
		IsCustomer: isCustomer,
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allUsers))

	// Filter by customer_id: if isCustomer==true, customer_id must not be nil; if isCustomer==false, customer_id must be nil
	filteredUserRows := make([]db.GetMarinaUsersListPaginatedRow, 0, len(userRows))
	if isCustomer != nil {
		for _, user := range userRows {
			if *isCustomer {
				if user.CustomerID != nil {
					filteredUserRows = append(filteredUserRows, user)
				}
			} else {
				if user.CustomerID == nil {
					filteredUserRows = append(filteredUserRows, user)
				}
			}
		}
	} else {
		filteredUserRows = userRows
	}

	// Create user responses with server instance for DME client
	userResponses := make([]responses.UserResponse, len(filteredUserRows))
	for i, user := range filteredUserRows {
		response := responses.NewUserResponseFromMarinaListRow(user, g.server)
		if response != nil {
			userResponses[i] = *response
		}
	}
	return responses.NewPaginatedResponse(userResponses, total, pagination.PageSize, pagination.Page).JSON(c)
}

// AssignUserToMarinaHandler assigns a user to a marina (creates a user_marinas record)
//
//	@Summary		Assign user to marina
//	@Description	Create an association between a user and a marina
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			body		body		requests.AssignUserToMarinaRequest	true	"User and marina IDs"
//	@Success		204			{object}	nil "No content"
//	@Failure		400			{object}	responses.Error "Bad request"
//	@Failure		500			{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/marina/assign [post]
func (g *UserHandler) AssignUserToMarinaHandler(c echo.Context) error {
	// Parse and validate request
	req := new(requests.AssignUserToMarinaRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := g.server.DB.Queries()
	user, err := queries.GetUserByID(c.Request().Context(), req.UserID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	if *user.IsCustomer && req.CustomerID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Customer User must have a customer ID").JSON(c)
	}

	params := db.AssignUserToMarinaParams{
		UserID:   req.UserID,
		MarinaID: req.MarinaID,
	}
	if *user.IsCustomer && req.CustomerID != nil {
		params.CustomerID = req.CustomerID
	}
	err = queries.AssignUserToMarina(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Only upsert customer settings if this is a customer user with a customer ID
	if *user.IsCustomer && req.CustomerID != nil {
		queries.UpsertCustomerSettings(c.Request().Context(), db.UpsertCustomerSettingsParams{
			MarinaID:     req.MarinaID,
			CustomerID:   *req.CustomerID,
			EnablePortal: utils.Pointer(true),
		})
	}

	return c.NoContent(http.StatusNoContent)
}

// UnassignUserFromMarinaHandler removes a user from a marina (deletes a user_marinas record)
//
//	@Summary		Unassign user from marina
//	@Description	Remove an association between a user and a marina
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			body		body		requests.AssignUserToMarinaRequest	true	"User and marina IDs"
//	@Success		204			{object}	nil "No content"
//	@Failure		400			{object}	responses.Error "Bad request"
//	@Failure		500			{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/marina/unassign [post]
func (g *UserHandler) UnassignUserFromMarinaHandler(c echo.Context) error {
	// Parse and validate request
	req := new(requests.AssignUserToMarinaRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := g.server.DB.Queries()
	ctx := c.Request().Context()

	// Get the user to check their current active marina
	user, err := queries.GetUserByID(ctx, req.UserID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// If the user's current active marina is the one being unassigned
	if user.MarinaID == req.MarinaID {
		// Get all marinas the user is assigned to (including the one being unassigned)
		marinas, err := queries.GetUserMarinasList(ctx, req.UserID)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		if len(marinas) <= 1 {
			return responses.NewErrorResponse(http.StatusBadRequest, "Cannot unassign the last marina from the user. Deactivate the user instead.").JSON(c)
		}
		var newActiveMarinaID uuid.UUID
		for _, m := range marinas {
			if m.ID != req.MarinaID {
				// Check if the marina is active
				if m.IsActive != nil && *m.IsActive {
					newActiveMarinaID = m.ID
					break
				}
			}
		}
		// If another active marina is found, switch to it
		if newActiveMarinaID != uuid.Nil {
			updateParams := db.UpdateUserParams{
				ID:                  user.ID,
				FirstName:           user.FirstName,
				LastName:            user.LastName,
				Email:               user.Email,
				EmailVerified:       user.EmailVerified,
				Phone:               user.Phone,
				Title:               user.Title,
				Image:               user.Image,
				PasswordHash:        user.PasswordHash,
				LastLogin:           user.LastLogin,
				FailedLoginAttempts: user.FailedLoginAttempts,
				LockedUntil:         user.LockedUntil,
				LastPasswordReset:   user.LastPasswordReset,
				MarinaID:            newActiveMarinaID,
				RoleID:              user.RoleID,
				IsSuperuser:         user.IsSuperuser,
				IsActive:            user.IsActive,
				UserAnalytics:       user.UserAnalytics,
			}
			_, err := queries.UpdateUser(ctx, updateParams)
			if err != nil {
				return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
			}
		} else {
			return responses.NewErrorResponse(http.StatusBadRequest, "Cannot unassign: the user has no other active marinas to switch to").JSON(c)
		}
	}

	params := db.UnassignUserFromMarinaParams{
		UserID:   req.UserID,
		MarinaID: req.MarinaID,
	}

	err = queries.UnassignUserFromMarina(ctx, params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return c.NoContent(http.StatusNoContent)
}

// UpdateUserHandler updates an existing user with partial fields
//
//	@Summary		Update user
//	@Description	Update a user with partial fields (only provided fields will be updated)
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			userId	path		string						true	"User ID"
//	@Param			user	body		requests.UpdateUserRequest	true	"User information to update"
//	@Success		200		{object}	responses.UserResponseWrapper "Updated user"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		404		{object}	responses.Error "User not found"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/{userId} [put]
func (g *UserHandler) UpdateUserHandler(c echo.Context) error {
	// Parse and validate user ID from path directly without binding
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		g.server.Logger.Zap.Error("Error parsing userId", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid user ID format").JSON(c)
	}

	// Log request content type and some request info for debugging
	g.server.Logger.Zap.Info("Update request headers",
		"Content-Type", c.Request().Header.Get("Content-Type"),
		"Content-Length", c.Request().ContentLength,
		"Method", c.Request().Method,
		"URL", c.Request().URL.String())

	// Get current user data to update only changed fields
	queries := g.server.DB.Queries()
	currentUser, err := queries.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, err).JSON(c)
	}

	// Check if this is a multipart form (which would include a file upload)
	contentType := c.Request().Header.Get("Content-Type")
	isMultipart := strings.HasPrefix(contentType, "multipart/form-data")

	// Build update params with current values that will be overridden if provided
	updateParams := db.UpdateUserParams{
		ID:                  userID,
		FirstName:           currentUser.FirstName,
		LastName:            currentUser.LastName,
		Email:               currentUser.Email,
		EmailVerified:       currentUser.EmailVerified,
		Phone:               currentUser.Phone,
		Title:               currentUser.Title,
		Image:               currentUser.Image,
		PasswordHash:        currentUser.PasswordHash,
		LastLogin:           currentUser.LastLogin,
		FailedLoginAttempts: currentUser.FailedLoginAttempts,
		LockedUntil:         currentUser.LockedUntil,
		LastPasswordReset:   currentUser.LastPasswordReset,
		MarinaID:            currentUser.MarinaID,
		RoleID:              currentUser.RoleID,
		IsSuperuser:         currentUser.IsSuperuser,
		IsActive:            currentUser.IsActive,
		UserAnalytics:       currentUser.UserAnalytics,
	}

	// Handle image upload if this is a multipart request
	if isMultipart {
		// Check if there's an image file in the form
		file, header, err := c.Request().FormFile("image")
		if err == nil {
			defer file.Close()

			// Upload the image to S3
			imagePath, err := g.server.ImageService.UploadImage(c.Request().Context(), file, header, s3.UserImageType)
			if err != nil {
				g.server.Logger.Zap.Error("Error uploading user image to S3", err)
				return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading image: "+err.Error()).JSON(c)
			}

			// Update the image path
			updateParams.Image = &imagePath
		}
		// Check if they want to change the marina, validate if the user got access to the new marina
		if marinaIDStr := c.FormValue("marinaId"); marinaIDStr != "" {
			if marinaID, err := uuid.Parse(marinaIDStr); err == nil {
				marina, err := queries.GetMarinaByID(c.Request().Context(), marinaID)
				if err != nil {
					return responses.NewErrorResponse(http.StatusBadRequest, "Marina not found").JSON(c)
				}
				organization, err := queries.GetOrganizationByID(c.Request().Context(), marina.OrganizationID)
				if err != nil {
					return responses.NewErrorResponse(http.StatusBadRequest, "Organization not found").JSON(c)
				}
				if organization.ID != currentUser.OrganizationID {
					return responses.NewErrorResponse(http.StatusBadRequest, "User does not have access to this marina").JSON(c)
				}
				_, err = queries.UserCanAccessMarina(c.Request().Context(), db.UserCanAccessMarinaParams{
					UserID:   userID,
					MarinaID: marinaID,
				})
				if err != nil {
					return responses.NewErrorResponse(http.StatusBadRequest, "User does not have access to this marina").JSON(c)
				}
			}
		}
		// Process other form fields regardless of whether an image was uploaded
		if firstName := c.FormValue("firstName"); firstName != "" {
			updateParams.FirstName = firstName
		}
		if lastName := c.FormValue("lastName"); lastName != "" {
			updateParams.LastName = lastName
		}
		if email := c.FormValue("email"); email != "" {
			updateParams.Email = utils.LowerCase(email)
		}
		if phone := c.FormValue("phone"); phone != "" {
			updateParams.Phone = &phone
		}
		if title := c.FormValue("title"); title != "" {
			updateParams.Title = &title
		}
		if password := c.FormValue("password"); password != "" {
			// Use helper function to hash password and update timestamp
			passwordHash, lastPasswordReset, err := utils.UpdatePasswordFields(password)
			if err != nil {
				return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password update").JSON(c)
			}
			updateParams.PasswordHash = utils.Pointer(passwordHash)
			updateParams.LastPasswordReset = lastPasswordReset
		}
		// Handle boolean and UUID fields
		if isActive := c.FormValue("isActive"); isActive != "" {
			active := isActive == "true"
			updateParams.IsActive = &active
		}
		if isSuperuser := c.FormValue("isSuperuser"); isSuperuser != "" {
			superuser := isSuperuser == "true"
			updateParams.IsSuperuser = &superuser
		}
		if marinaIDStr := c.FormValue("marinaId"); marinaIDStr != "" {
			if marinaID, err := uuid.Parse(marinaIDStr); err == nil {
				updateParams.MarinaID = marinaID
			}
		}
		if roleIDStr := c.FormValue("roleId"); roleIDStr != "" {
			if roleID, err := uuid.Parse(roleIDStr); err == nil {
				updateParams.RoleID = roleID
			}
		}
		// Add UserAnalytics handling for multipart form
		if userAnalytics := c.FormValue("userAnalytics"); userAnalytics != "" {
			analytics := userAnalytics == "true"
			updateParams.UserAnalytics = &analytics
		}
	} else {
		// Parse and validate the JSON request body
		req := new(requests.UpdateUserRequest)
		if err := c.Bind(req); err != nil {
			g.server.Logger.Zap.Error("Error binding request body", err)
			return responses.NewErrorResponse(http.StatusBadRequest, "Error parsing request: "+err.Error()).JSON(c)
		}
		if err := c.Validate(req); err != nil {
			g.server.Logger.Zap.Error("Error validating request body", err)
			return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
		}

		// Check if they want to change the marina, validate if the user got access to the new marina
		if req.MarinaID != nil {
			marina, err := queries.GetMarinaByID(c.Request().Context(), *req.MarinaID)
			if err != nil {
				return responses.NewErrorResponse(http.StatusBadRequest, "Marina not found").JSON(c)
			}
			organization, err := queries.GetOrganizationByID(c.Request().Context(), marina.OrganizationID)
			if err != nil {
				return responses.NewErrorResponse(http.StatusBadRequest, "Organization not found").JSON(c)
			}
			if organization.ID != currentUser.OrganizationID {
				return responses.NewErrorResponse(http.StatusBadRequest, "User does not have access to this marina").JSON(c)
			}
			_, err = queries.UserCanAccessMarina(c.Request().Context(), db.UserCanAccessMarinaParams{
				UserID:   userID,
				MarinaID: *req.MarinaID,
			})
			if err != nil {
				return responses.NewErrorResponse(http.StatusBadRequest, "User does not have access to this marina").JSON(c)
			}
		}

		// Update only the fields that were provided in the request
		if req.FirstName != nil {
			updateParams.FirstName = *req.FirstName
		}
		if req.LastName != nil {
			updateParams.LastName = *req.LastName
		}
		if req.Email != nil {
			updateParams.Email = utils.LowerCase(*req.Email)
		}
		if req.Phone != nil {
			updateParams.Phone = req.Phone
		}
		if req.Title != nil {
			updateParams.Title = req.Title
		}
		if req.Image != nil {
			updateParams.Image = req.Image
		}
		if req.Password != nil {
			// Use helper function to hash password and update timestamp
			passwordHash, lastPasswordReset, err := utils.UpdatePasswordFields(*req.Password)
			if err != nil {
				return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password update").JSON(c)
			}
			updateParams.PasswordHash = utils.Pointer(passwordHash)
			updateParams.LastPasswordReset = lastPasswordReset
		}
		if req.MarinaID != nil {
			updateParams.MarinaID = *req.MarinaID
		}
		if req.RoleID != nil {
			updateParams.RoleID = *req.RoleID
		}
		if req.IsSuperuser != nil {
			updateParams.IsSuperuser = req.IsSuperuser
		}
		if req.IsActive != nil {
			updateParams.IsActive = req.IsActive
		}
		if req.UserAnalytics != nil {
			updateParams.UserAnalytics = req.UserAnalytics
		}

	}

	// Perform update
	updatedUser, err := queries.UpdateUser(c.Request().Context(), updateParams)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.NewUserResponseSuccess(updatedUser)
	return response.JSON(c)
}

// ResetPassword
//
//	@Summary		Reset user password
//	@Description	Reset authenticated user's password and check against previous passwords
//	@ID				user-reset-password
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			params	body		requests.ResetPasswordRequest	true	"Password reset info"
//	@Success		200		{object}	responses.BaseResponse			"Password reset success"
//	@Failure		400		{object}	responses.Error					"Validation error"
//	@Failure		401		{object}	responses.Error					"Authentication error"
//	@Failure		500		{object}	responses.Error					"Server error"
//	@Security		ApiKeyAuth
//	@Router			/user/reset-password [post]
func (g *UserHandler) ResetPassword(c echo.Context) error {
	logger := g.server.Logger
	queries := g.server.DB.Queries()

	resetRequest := new(requests.ResetPasswordRequest)

	if err := c.Bind(resetRequest); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(resetRequest); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Get user from database
	user, err := queries.GetUserByID(c.Request().Context(), resetRequest.UserID)
	if err != nil {
		logger.Zap.Info("password reset failed: user not found", c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusUnauthorized, "User not found").JSON(c)
	}

	// Verify current password
	if err := utils.VerifyPassword(*user.PasswordHash, resetRequest.OldPassword); err != nil {
		logger.Zap.Info("password reset failed: invalid current password", c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusUnauthorized, "Invalid current password").JSON(c)
	}

	// Get password history (last 4 passwords)
	historyLimit := int32(4)
	passwordHistory, err := queries.GetPasswordHistoryByUser(c.Request().Context(), db.GetPasswordHistoryByUserParams{
		UserID: user.ID,
		Limit:  historyLimit,
	})
	if err != nil {
		logger.Zap.Error("failed to get password history", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password reset").JSON(c)
	}

	// Create a slice of password hashes from the history
	historyHashes := make([]string, 0, len(passwordHistory)+1)
	historyHashes = append(historyHashes, *user.PasswordHash) // Add current password to history
	historyHashes = append(historyHashes, passwordHistory...) // Add previous password hashes

	// Check if new password matches any of the last 4 passwords
	isReused, err := utils.IsPasswordInHistory(resetRequest.NewPassword, historyHashes)
	if err != nil {
		logger.Zap.Error("failed to check password history", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password reset").JSON(c)
	}

	if isReused {
		return responses.NewErrorResponse(http.StatusBadRequest, utils.ErrPasswordReused).JSON(c)
	}

	// Hash the new password
	newPasswordHash, lastPasswordReset, err := utils.UpdatePasswordFields(resetRequest.NewPassword)
	if err != nil {
		logger.Zap.Error("failed to hash new password", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password reset").JSON(c)
	}

	// Add old password to history first
	_, err = queries.AddPasswordToHistory(c.Request().Context(), db.AddPasswordToHistoryParams{
		UserID:       user.ID,
		PasswordHash: *user.PasswordHash, // Store the old password that's being replaced
	})
	if err != nil {
		logger.Zap.Error("failed to update password history", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating password history").JSON(c)
	}

	// Update user record with new password
	updateParams := db.UpdateUserParams{
		ID:                  user.ID,
		FirstName:           user.FirstName,
		LastName:            user.LastName,
		Email:               user.Email,
		EmailVerified:       user.EmailVerified,
		Phone:               user.Phone,
		Title:               user.Title,
		Image:               user.Image,
		PasswordHash:        &newPasswordHash,
		LastLogin:           user.LastLogin,
		FailedLoginAttempts: user.FailedLoginAttempts,
		LockedUntil:         user.LockedUntil,
		LastPasswordReset:   lastPasswordReset,
		MarinaID:            user.MarinaID,
		RoleID:              user.RoleID,
		IsSuperuser:         user.IsSuperuser,
		IsActive:            user.IsActive,
		UserAnalytics:       user.UserAnalytics,
	}

	_, err = queries.UpdateUser(c.Request().Context(), updateParams)
	if err != nil {
		logger.Zap.Error("failed to update user password", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating password").JSON(c)
	}

	// Cleanup old passwords if we have more than the limit
	err = queries.CleanupOldPasswords(c.Request().Context(), db.CleanupOldPasswordsParams{
		UserID: user.ID,
		Offset: historyLimit,
	})
	if err != nil {
		logger.Zap.Error("failed to cleanup old passwords", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error cleaning up password history").JSON(c)
	}

	return responses.NewMessageResponse(http.StatusOK, "Password updated successfully").JSON(c)
}

// RecoverPassword
//
//	@Summary		Complete password recovery
//	@Description	Verify token and set new password
//	@ID				user-recover-password
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			params	body		requests.CompletePasswordRecoveryRequest	true	"Recovery token and new password"
//	@Success		200		{object}	responses.BaseResponse			"Password updated successfully"
//	@Failure		400		{object}	responses.Error					"Invalid token or validation error"
//	@Failure		500		{object}	responses.Error					"Server error"
//	@Router			/user/recover-password [post]
func (g *UserHandler) RecoverPassword(c echo.Context) error {
	logger := g.server.Logger
	queries := g.server.DB.Queries()

	recoverRequest := new(requests.CompletePasswordRecoveryRequest)

	if err := c.Bind(recoverRequest); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(recoverRequest); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Look up the token
	recoveryRecord, err := queries.GetPasswordRecoveryToken(c.Request().Context(), db.GetPasswordRecoveryTokenParams{
		Token: recoverRequest.Token,
		Email: recoverRequest.Email,
	})
	if err != nil {
		logger.Zap.Info("password recovery failed: invalid or expired token", c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid or expired recovery token").JSON(c)
	}

	// Get the user from the recovery record
	user, err := queries.GetUserByID(c.Request().Context(), recoveryRecord.UserID)
	if err != nil {
		logger.Zap.Error("password recovery failed: user not found", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password recovery").JSON(c)
	}

	// Check password history
	historyLimit := int32(4)
	passwordHistory, err := queries.GetPasswordHistoryByUser(c.Request().Context(), db.GetPasswordHistoryByUserParams{
		UserID: user.ID,
		Limit:  historyLimit,
	})
	if err != nil {
		logger.Zap.Error("failed to get password history", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password recovery").JSON(c)
	}

	// Create a slice of password hashes from the history
	historyHashes := make([]string, 0, len(passwordHistory)+1)
	historyHashes = append(historyHashes, *user.PasswordHash) // Add current password to history
	historyHashes = append(historyHashes, passwordHistory...) // Add previous password hashes

	// Check if new password matches any of the last 4 passwords
	isReused, err := utils.IsPasswordInHistory(recoverRequest.NewPassword, historyHashes)
	if err != nil {
		logger.Zap.Error("failed to check password history", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password recovery").JSON(c)
	}

	if isReused {
		return responses.NewErrorResponse(http.StatusBadRequest, utils.ErrPasswordReused).JSON(c)
	}

	// Hash the new password
	newPasswordHash, lastPasswordReset, err := utils.UpdatePasswordFields(recoverRequest.NewPassword)
	if err != nil {
		logger.Zap.Error("failed to hash new password", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password recovery").JSON(c)
	}

	// Add old password to history first
	_, err = queries.AddPasswordToHistory(c.Request().Context(), db.AddPasswordToHistoryParams{
		UserID:       user.ID,
		PasswordHash: *user.PasswordHash, // Store the old password that's being replaced
	})
	if err != nil {
		logger.Zap.Error("failed to update password history", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating password history").JSON(c)
	}

	// Update user record with new password
	updateParams := db.UpdateUserParams{
		ID:                  user.ID,
		FirstName:           user.FirstName,
		LastName:            user.LastName,
		Email:               user.Email,
		EmailVerified:       user.EmailVerified,
		Phone:               user.Phone,
		Title:               user.Title,
		Image:               user.Image,
		PasswordHash:        &newPasswordHash,
		LastLogin:           user.LastLogin,
		FailedLoginAttempts: user.FailedLoginAttempts,
		LockedUntil:         user.LockedUntil,
		LastPasswordReset:   lastPasswordReset,
		MarinaID:            user.MarinaID,
		RoleID:              user.RoleID,
		IsSuperuser:         user.IsSuperuser,
		IsActive:            user.IsActive,
		UserAnalytics:       user.UserAnalytics,
	}

	_, err = queries.UpdateUser(c.Request().Context(), updateParams)
	if err != nil {
		logger.Zap.Error("failed to update user password", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating password").JSON(c)
	}

	// Cleanup old passwords if we have more than the limit
	err = queries.CleanupOldPasswords(c.Request().Context(), db.CleanupOldPasswordsParams{
		UserID: user.ID,
		Offset: historyLimit,
	})
	if err != nil {
		logger.Zap.Error("failed to cleanup old passwords", err, c.Response().Header().Get(echo.HeaderXRequestID))
		// Continue even if cleanup fails
	}

	// Mark the token as used
	err = queries.MarkTokenAsUsed(c.Request().Context(), db.MarkTokenAsUsedParams{
		Token: recoverRequest.Token,
		Email: recoverRequest.Email,
	})
	if err != nil {
		logger.Zap.Error("failed to mark token as used", err, c.Response().Header().Get(echo.HeaderXRequestID))
		// Continue even if marking token as used fails
	}

	return responses.NewMessageResponse(http.StatusOK, "Password has been successfully updated").JSON(c)
}

// ForgotPassword
//
//	@Summary		Forgot password
//	@Description	Initiate password recovery process
//	@ID				user-forgot-password
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			params	body		requests.ForgotPasswordRequest	true	"Email for password recovery"
//	@Success		200		{object}	responses.BaseResponse			"Password recovery email sent"
//	@Failure		400		{object}	responses.Error					"Validation error"
//	@Failure		404		{object}	responses.Error					"User not found"
//	@Failure		500		{object}	responses.Error					"Server error"
//	@Router			/user/forgot-password [post]
func (g *UserHandler) ForgotPassword(c echo.Context) error {
	logger := g.server.Logger
	queries := g.server.DB.Queries()

	forgotRequest := new(requests.ForgotPasswordRequest)

	if err := c.Bind(forgotRequest); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(forgotRequest); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Find user by email
	user, err := queries.GetUserByEmail(c.Request().Context(), forgotRequest.Email)
	if err != nil {
		// Don't reveal whether the user exists or not for security
		logger.Zap.Info("password recovery requested for non-existent email", c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewMessageResponse(http.StatusOK, "If your email is registered, you will receive password recovery instructions").JSON(c)
	}

	// Generate a random token
	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		logger.Zap.Error("failed to generate recovery token", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password recovery").JSON(c)
	}

	// Set expiration time (24 hours from now)
	expiresAt := utils.PgTimeNow()
	expiresAt.Time = time.Now().UTC().Add(24 * time.Hour)

	// Store token in database
	_, err = queries.CreatePasswordRecoveryToken(c.Request().Context(), db.CreatePasswordRecoveryTokenParams{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		ExpiresAt: expiresAt,
	})
	if err != nil {
		logger.Zap.Error("failed to create password recovery token", err, c.Response().Header().Get(echo.HeaderXRequestID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password recovery").JSON(c)
	}

	// Initialize the SendGrid client if we have API key
	// Create a SendGrid client
	sgClient := g.server.SendGrid

	// Build the reset URL
	baseURL := g.server.Config.App.FrontendBaseURL // Default URL
	resetURL := fmt.Sprintf("%s/reset-password?token=%s&email=%s",
		baseURL,
		token,
		url.QueryEscape(user.Email))

	// Create template data
	templateData := sendgrid.PasswordResetTemplateData{
		UserName:        user.FirstName + " " + user.LastName,
		ResetURL:        resetURL,
		TermsConditions: baseURL + "/terms-conditions",
	}

	// Send email using specialized password reset method
	taskID, resultChan, err := sgClient.SendPasswordResetEmail(
		[]string{user.Email},
		"Reset Your Password",
		templateData,
	)

	if err != nil {
		logger.Zap.Errorw("Failed to send password reset email", "error", err)
	} else {
		logger.Zap.Infow("Password reset email queued",
			"email", user.Email,
			"task_id", taskID.String())

		// Log the email attempt (non-blocking)
		go func() {
			result := <-resultChan
			if result.Status == sendgrid.StatusSent {
				logger.Zap.Infow("Password reset email sent successfully",
					"email", user.Email,
					"task_id", result.ID.String())
			} else {
				logger.Zap.Errorw("Failed to send password reset email",
					"email", user.Email,
					"task_id", result.ID.String(),
					"error", result.Error)
			}
		}()
	}

	return responses.NewMessageResponse(http.StatusOK, "If your email is registered, you will receive password recovery instructions").JSON(c)
}

// CreateCustomerUserHandler creates a new customer user
//
//	@Summary		Create customer user
//	@Description	Create a new customer user and assign them to a marina and customer
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			user	body		requests.CreateCustomerUserRequest	true	"User information"
//	@Success		201		{object}	responses.UserResponseWrapper "Created user"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/customer-portal [post]
func (g *UserHandler) CreateCustomerUserHandler(c echo.Context) error {
	// Parse and validate the request body
	req := new(requests.CreateCustomerUserRequest)
	logger := g.server.Logger
	cfg := g.server.Config
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := g.server.DB.Queries()

	// Check if the role is a customer role
	role, err := queries.GetRoleByID(c.Request().Context(), req.RoleID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Role not found").JSON(c)
	}

	if !*role.IsCustomerRole {
		return responses.NewErrorResponse(http.StatusBadRequest, "Role is not a customer role").JSON(c)
	}

	// Check if the marina and organization exist and linked
	marina, err := queries.GetMarinaByID(c.Request().Context(), req.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Marina not found").JSON(c)
	}

	organization, err := queries.GetOrganizationByID(c.Request().Context(), req.OrganizationID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Organization not found").JSON(c)
	}

	if marina.OrganizationID != organization.ID {
		return responses.NewErrorResponse(http.StatusBadRequest, "Marina and organization are not linked").JSON(c)
	}

	// Check if the email is already taken
	email := utils.LowerCase(req.Email)
	userByEmail, err := queries.GetUserByEmail(c.Request().Context(), email)
	if err == nil {
		// Email exists, check if user is already assigned to this marina
		canAccess, err := queries.UserCanAccessMarina(c.Request().Context(), db.UserCanAccessMarinaParams{
			UserID:   userByEmail.ID,
			MarinaID: req.MarinaID,
		})
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		if canAccess {
			return responses.NewErrorResponse(http.StatusBadRequest, "Email already taken for this marina").JSON(c)
		}
		// Assign the existing user to the marina with CustomerID
		assignUserToMarina := db.AssignUserToMarinaParams{
			UserID:     userByEmail.ID,
			MarinaID:   req.MarinaID,
			CustomerID: req.CustomerID,
		}
		err = queries.AssignUserToMarina(c.Request().Context(), assignUserToMarina)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		// Upsert customer settings as in the original logic
		_, err = queries.UpsertCustomerSettings(c.Request().Context(), db.UpsertCustomerSettingsParams{
			MarinaID:     req.MarinaID,
			CustomerID:   *req.CustomerID,
			EnablePortal: utils.Pointer(true),
		})
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		// Send assigned_to_marina email
		orgLogo := ""
		if organization.Image != nil {
			orgLogo = *organization.Image
		}
		assignedData := sendgrid.AssignedToMarinaTemplateData{
			CustomerLogo:    orgLogo,
			BusinessName:    marina.Name,
			UserName:        userByEmail.FirstName,
			HomeURL:         cfg.App.HomeURL(),
			TermsConditions: cfg.App.TermsConditionsURL(),
		}
		_, _, _ = g.server.SendGrid.SendAssignedToMarinaEmail(
			[]string{userByEmail.Email},
			"You have been assigned to a new marina",
			assignedData,
		)
		return responses.NewMessageResponse(http.StatusOK, "User assigned to marina").JSON(c)
	}
	username := req.Username
	if username == "" {
		username = utils.GenerateUsername(req.FirstName)
	}

	// Check if the username is already taken
	_, err = queries.GetUserByUsername(c.Request().Context(), username)
	if err == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Username already taken").JSON(c)
	}

	// Set default values for nullable fields if not provided
	isActive := true
	isSuperuser := false
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	params := db.CreateCustomerUserParams{
		Username:       username,
		FirstName:      req.FirstName,
		LastName:       req.LastName,
		Email:          email,
		Phone:          req.Phone,
		Title:          req.Title,
		Image:          req.Image,
		OrganizationID: req.OrganizationID,
		MarinaID:       req.MarinaID,
		RoleID:         req.RoleID,
		CustomerID:     req.CustomerID,
		IsCustomer:     utils.Pointer(true),
		IsSuperuser:    &isSuperuser,
		IsActive:       &isActive,
		UserAnalytics:  utils.Pointer(true),
	}

	user, err := queries.CreateCustomerUser(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	assignUserToMarina := db.AssignUserToMarinaParams{
		UserID:     user.ID,
		MarinaID:   req.MarinaID,
		CustomerID: req.CustomerID,
	}

	err = queries.AssignUserToMarina(c.Request().Context(), assignUserToMarina)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Update customer settings
	_, err = queries.UpsertCustomerSettings(c.Request().Context(), db.UpsertCustomerSettingsParams{
		MarinaID:     req.MarinaID,
		CustomerID:   *req.CustomerID,
		EnablePortal: utils.Pointer(true),
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	_, err = queries.CreateInvite(c.Request().Context(), db.CreateInviteParams{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		ExpiresAt: utils.PgTimeNowAdd(240 * time.Hour),
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	inviteURL := fmt.Sprintf("%s/%s?token=%s&email=%s",
		cfg.App.FrontendBaseURL,
		cfg.App.InvitationCustomerRoute,
		token,
		url.QueryEscape(user.Email))
	termsConditionsURL := fmt.Sprintf("%s/%s",
		cfg.App.FrontendBaseURL,
		cfg.App.TermsConditionsRoute,
	)
	templateData := sendgrid.InviteCustomerTemplateData{
		UserName:        user.FirstName,
		InviteURL:       inviteURL,
		TermsConditions: termsConditionsURL,
	}
	taskID, resultChan, err := g.server.SendGrid.SendInviteCustomerEmail(
		[]string{user.Email},
		"DockMaster Customer Portal Invite",
		templateData,
	)
	if err != nil {
		logger.Zap.Errorw("Failed to send invite email", "error", err)
	} else {
		logger.Zap.Infow("Invite email queued",
			"email", user.Email,
			"task_id", taskID.String())

		// // Create notification for marina staff about new user invitation
		// marinaUsers, err := queries.GetUsersByMarina(c.Request().Context(), db.GetUsersByMarinaParams{
		// 	MarinaID:   req.MarinaID,
		// 	IsCustomer: utils.Pointer(false), // Get marina staff, not customers
		// })
		// if err != nil {
		// 	logger.Zap.Warnw("Failed to get marina users for invite notification", "marina_id", req.MarinaID, "error", err)
		// } else {
		// 	// Create notifications for marina staff about the new invite
		// 	for _, userRow := range marinaUsers {
		// 		// Only notify active users, and don't notify the user who just created the invite
		// 		if userRow.IsActive != nil && *userRow.IsActive && userRow.ID != user.ID {
		// 			notificationErr := g.notificationService.CreateInviteNotification(
		// 				c.Request().Context(),
		// 				userRow.ID,
		// 				userRow.OrganizationID,
		// 				userRow.MarinaID,
		// 				user.FirstName+" "+user.LastName, // Invited user's name
		// 			)
		// 			if notificationErr != nil {
		// 				logger.Zap.Warnw("Failed to create invite notification for marina user",
		// 					"user_id", userRow.ID,
		// 					"error", notificationErr)
		// 			}
		// 		}
		// 	}
		// }

		// Log the email attempt (non-blocking)
		go func() {
			result := <-resultChan
			if result.Status == sendgrid.StatusSent {
				logger.Zap.Infow("Invite email sent successfully",
					"email", user.Email,
					"task_id", result.ID.String())
			} else {
				logger.Zap.Errorw("Failed to send invite email",
					"email", user.Email,
					"task_id", result.ID.String(),
					"error", result.Error)
			}
		}()
	}

	response := responses.NewUserResponseSuccess(user)
	return c.JSON(http.StatusCreated, response)
}

// CreateUserWithInvitationHandler creates a new user with invitation
//
//	@Summary		Create user with invitation
//	@Description	Create a new user and send them an invitation email
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			user	body		requests.CreateUserWithInvitationRequest	true	"User information"
//	@Success		201		{object}	responses.UserResponseWrapper "Created user"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/user/invite [post]
func (g *UserHandler) CreateUserWithInvitationHandler(c echo.Context) error {
	// Parse and validate the request body
	req := new(requests.CreateUserWithInvitationRequest)
	logger := g.server.Logger
	cfg := g.server.Config
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := g.server.DB.Queries()

	// Check if the role exists
	_, err := queries.GetRoleByID(c.Request().Context(), req.RoleID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Role not found").JSON(c)
	}

	// Check if the marina and organization exist and linked
	marina, err := queries.GetMarinaByID(c.Request().Context(), req.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Marina not found").JSON(c)
	}

	organization, err := queries.GetOrganizationByID(c.Request().Context(), req.OrganizationID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Organization not found").JSON(c)
	}

	if marina.OrganizationID != organization.ID {
		return responses.NewErrorResponse(http.StatusBadRequest, "Marina and organization are not linked").JSON(c)
	}

	role, err := queries.GetRoleByID(c.Request().Context(), req.RoleID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Role not found").JSON(c)
	}

	// Check if the email is already taken
	email := utils.LowerCase(req.Email)
	userByEmail, err := queries.GetUserByEmail(c.Request().Context(), email)
	if err == nil {
		// Email exists, check if user is already assigned to this marina
		canAccess, err := queries.UserCanAccessMarina(c.Request().Context(), db.UserCanAccessMarinaParams{
			UserID:   userByEmail.ID,
			MarinaID: req.MarinaID,
		})
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		if canAccess {
			return responses.NewErrorResponse(http.StatusBadRequest, "Email already taken for this marina").JSON(c)
		}
		// Assign the existing user to the marina
		assignUserToMarina := db.AssignUserToMarinaParams{
			UserID:   userByEmail.ID,
			MarinaID: req.MarinaID,
		}
		err = queries.AssignUserToMarina(c.Request().Context(), assignUserToMarina)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		// Send assigned_to_marina email
		orgLogo := ""
		if organization.Image != nil {
			orgLogo = *organization.Image
		}
		assignedData := sendgrid.AssignedToMarinaTemplateData{
			CustomerLogo:    orgLogo,
			BusinessName:    marina.Name,
			UserName:        userByEmail.FirstName,
			HomeURL:         cfg.App.HomeURL(),
			TermsConditions: cfg.App.TermsConditionsURL(),
		}
		_, _, _ = g.server.SendGrid.SendAssignedToMarinaEmail(
			[]string{userByEmail.Email},
			"You have been assigned to a new marina",
			assignedData,
		)
		return responses.NewMessageResponse(http.StatusOK, "User assigned to marina").JSON(c)
	}
	username := req.Username
	if username == "" {
		username = utils.GenerateUsername(req.FirstName)
	}

	// Check if the username is already taken
	_, err = queries.GetUserByUsername(c.Request().Context(), username)
	if err == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Username already taken").JSON(c)
	}

	// Set default values for nullable fields if not provided
	failedLoginAttempts := int32(0)
	isActive := true
	isSuperuser := false
	if role.Type == "internal" {
		isSuperuser = true
	}
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Create user without password hash - they'll set it via invitation
	params := db.CreateUserParams{
		Username:            username,
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		Email:               email,
		Phone:               req.Phone,
		Title:               req.Title,
		Image:               req.Image,
		PasswordHash:        nil, // No password hash - user will set via invitation
		FailedLoginAttempts: &failedLoginAttempts,
		LastPasswordReset:   utils.PgTimeNow(),
		OrganizationID:      req.OrganizationID,
		MarinaID:            req.MarinaID,
		RoleID:              req.RoleID,
		IsSuperuser:         &isSuperuser,
		IsActive:            &isActive,
		UserAnalytics:       utils.Pointer(true),
	}

	user, err := queries.CreateUser(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Assign user to marina
	assignUserToMarina := db.AssignUserToMarinaParams{
		UserID:   user.ID,
		MarinaID: req.MarinaID,
	}

	err = queries.AssignUserToMarina(c.Request().Context(), assignUserToMarina)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Generate invitation token
	token, err := utils.GenerateRandomToken(32)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Create invitation record
	_, err = queries.CreateInvite(c.Request().Context(), db.CreateInviteParams{
		UserID:    user.ID,
		Email:     user.Email,
		Token:     token,
		ExpiresAt: utils.PgTimeNowAdd(240 * time.Hour), // 10 days
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Build invitation URL
	inviteURL := fmt.Sprintf("%s/%s?token=%s&email=%s",
		cfg.App.FrontendBaseURL,
		cfg.App.InvitationRoute,
		token,
		url.QueryEscape(user.Email))
	termsConditionsURL := fmt.Sprintf("%s/%s",
		cfg.App.FrontendBaseURL,
		cfg.App.TermsConditionsRoute,
	)

	// Prepare email template data
	templateData := sendgrid.InviteTemplateData{
		UserName:        user.FirstName,
		InviteURL:       inviteURL,
		TermsConditions: termsConditionsURL,
	}

	// Send invitation email
	taskID, resultChan, err := g.server.SendGrid.SendInviteEmail(
		[]string{user.Email},
		"DockMaster Platform Invite",
		templateData,
	)
	if err != nil {
		logger.Zap.Errorw("Failed to send invite email", "error", err)
	} else {
		logger.Zap.Infow("Invite email queued",
			"email", user.Email,
			"task_id", taskID.String())

		// // Create notification for marina staff about new user invitation
		// marinaUsers, err := queries.GetUsersByMarina(c.Request().Context(), db.GetUsersByMarinaParams{
		// 	MarinaID:   req.MarinaID,
		// 	IsCustomer: utils.Pointer(false), // Get marina staff, not customers
		// })
		// if err != nil {
		// 	logger.Zap.Warnw("Failed to get marina users for invite notification", "marina_id", req.MarinaID, "error", err)
		// } else {
		// 	// Create notifications for marina staff about the new invite
		// 	for _, userRow := range marinaUsers {
		// 		// Only notify active users, and don't notify the user who just created the invite
		// 		if userRow.IsActive != nil && *userRow.IsActive && userRow.ID != user.ID {
		// 			notificationErr := g.notificationService.CreateInviteNotification(
		// 				c.Request().Context(),
		// 				userRow.ID,
		// 				userRow.OrganizationID,
		// 				userRow.MarinaID,
		// 				user.FirstName+" "+user.LastName, // Invited user's name
		// 			)
		// 			if notificationErr != nil {
		// 				logger.Zap.Warnw("Failed to create invite notification for marina user",
		// 					"user_id", userRow.ID,
		// 					"error", notificationErr)
		// 			}
		// 		}
		// 	}
		// }

		// Log the email attempt (non-blocking)
		go func() {
			result := <-resultChan
			if result.Status == sendgrid.StatusSent {
				logger.Zap.Infow("Invite email sent successfully",
					"email", user.Email,
					"task_id", result.ID.String())
			} else {
				logger.Zap.Errorw("Failed to send invite email",
					"email", user.Email,
					"task_id", result.ID.String(),
					"error", result.Error)
			}
		}()
	}

	response := responses.NewUserResponseSuccess(user)
	return c.JSON(http.StatusCreated, response)
}

// GetUsersByCustomerIDHandler gets users by customer ID
//
//	@Summary		Get users by customer ID
//	@Description	Get all users in a specific customer
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			customerId	path		string	true	"Customer ID"
//	@Param			marinaId	query		string	true	"Marina ID"
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Success		200			{object}	responses.UserListResponse "Paginated list of users in the marina for this  customer"
//	@Failure		400			{object}	responses.Error "Bad request"
//	@Failure		500			{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//		@Router			/user/customer-portal/{customerId} [get]
func (g *UserHandler) GetUsersByCustomerIDHandler(c echo.Context) error {
	// Parse marina ID
	customerIDStr := c.Param("customerId")
	if customerIDStr == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Customer ID is required").JSON(c)
	}
	customerID := &customerIDStr
	marinaIDStr := c.QueryParam("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Marina ID (UUID) is required").JSON(c)
	}

	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()

	// Get paginated users by marina
	params := db.GetCustomerMarinaUsersPaginatedParams{
		MarinaID:   marinaID,
		CustomerID: customerID,
		Limit:      pagination.PageSize,
		Offset:     (pagination.Page - 1) * pagination.PageSize,
	}
	users, err := queries.GetCustomerMarinaUsersPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Get total count for pagination
	allUsers, err := queries.CountCustomerMarinaUsers(c.Request().Context(), db.CountCustomerMarinaUsersParams{
		MarinaID:   marinaID,
		CustomerID: customerID,
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := allUsers

	return responses.NewUsersPaginatedResponse(users, total, pagination.PageSize, pagination.Page).JSON(c)
}

// GetUsersNotAssignedToMarinaHandler gets users not assigned to a specific marina
//
//	@Summary		Get users not assigned to a marina
//	@Description	Get users who are not assigned to a specific marina (through user_marinas table)
//	@Tags		User
//	@Accept		json
//	@Produce	json
//	@Param		marinaId	path	string	true	"Marina ID"
//	@Param		page		query	int	false	"Page number" default(1)
//	@Param		pageSize	query	int	false	"Page size" default(10)
//	@Success	200	{object} responses.UserListResponse "List of users not assigned to the marina"
//	@Failure	400	{object} responses.Error "Bad request"
//	@Failure	500	{object} responses.Error "Server error"
//	@Security	ApiKeyAuth
//
//	@Router		/user/marina/{marinaId}/not-assigned [get]
func (g *UserHandler) GetUsersNotAssignedToMarinaHandler(c echo.Context) error {
	// Parse marina ID
	marinaIDStr := c.Param("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID").JSON(c)
	}

	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()
	params := db.GetUsersNotAssignedToMarinaPaginatedParams{
		MarinaID: marinaID,
		Limit:    pagination.PageSize,
		Offset:   (pagination.Page - 1) * pagination.PageSize,
	}
	users, err := queries.GetUsersNotAssignedToMarinaPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total, err := queries.CountUsersNotAssignedToMarina(c.Request().Context(), marinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	return responses.NewUsersPaginatedResponse(users, total, pagination.PageSize, pagination.Page).JSON(c)
}
