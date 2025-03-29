package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type UserHandler struct {
	server *s.Server
}

func NewUserHandler(server *s.Server) *UserHandler {
	return &UserHandler{server: server}
}

// ListUsersHandler lists all existing users
//
//	@Summary		List users
//	@Description	get users
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"
//	@Param			pageSize	query		int		false	"Page size"
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

	// Hash the password
	passwordHash, err := utils.HashPassword(req.Password)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing password").JSON(c)
	}

	// Set default values for nullable fields if not provided
	failedLoginAttempts := int32(0)
	isActive := true
	isSuperuser := false
	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.IsSuperuser != nil {
		isSuperuser = *req.IsSuperuser
	}

	params := db.CreateUserParams{
		Username:            req.Username,
		FirstName:           req.FirstName,
		LastName:            req.LastName,
		Email:               req.Email,
		Phone:               req.Phone,
		Title:               req.Title,
		Image:               req.Image,
		PasswordHash:        passwordHash,
		FailedLoginAttempts: &failedLoginAttempts,
		LastPasswordReset:   utils.PgTimeNow(),
		OrganizationID:      req.OrganizationID,
		MarinaID:            req.MarinaID,
		RoleID:              req.RoleID,
		IsSuperuser:         &isSuperuser,
		IsActive:            &isActive,
	}

	user, err := queries.CreateUser(c.Request().Context(), params)
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
//	@Param			page	query		int		false	"Page number"
//	@Param			pageSize	query		int		false	"Page size"
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
//	@Param			page			query		int		false	"Page number"
//	@Param			pageSize		query		int		false	"Page size"
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
//	@Param			page		query		int		false	"Page number"
//	@Param			pageSize	query		int		false	"Page size"
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

	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()

	// Get paginated users by marina
	params := db.GetUsersByMarinaPaginatedParams{
		MarinaID: marinaID,
		Limit:    pagination.PageSize,
		Offset:   (pagination.Page - 1) * pagination.PageSize,
	}
	users, err := queries.GetUsersByMarinaPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Get total count for pagination
	allUsers, err := queries.GetUsersByMarina(c.Request().Context(), marinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allUsers))

	return responses.NewUsersPaginatedResponse(users, total, pagination.PageSize, pagination.Page).JSON(c)
}

// GetMarinaUsersList gets all users associated with a marina through user_marinas
//
//	@Summary		Get marina users
//	@Description	Get users assigned to a marina (through user_marinas table)
//	@Tags			User
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	path		string	true	"Marina ID"
//	@Param			page		query		int		false	"Page number"
//	@Param			pageSize	query		int		false	"Page size"
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

	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()

	// Get paginated marina users list
	params := db.GetMarinaUsersListPaginatedParams{
		MarinaID: marinaID,
		Limit:    pagination.PageSize,
		Offset:   (pagination.Page - 1) * pagination.PageSize,
	}
	users, err := queries.GetMarinaUsersListPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Get total count for pagination
	allUsers, err := queries.GetMarinaUsersList(c.Request().Context(), marinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allUsers))

	return responses.NewUsersPaginatedResponse(users, total, pagination.PageSize, pagination.Page).JSON(c)
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

	params := db.AssignUserToMarinaParams{
		UserID:   req.UserID,
		MarinaID: req.MarinaID,
	}

	queries := g.server.DB.Queries()
	err := queries.AssignUserToMarina(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
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

	params := db.UnassignUserFromMarinaParams{
		UserID:   req.UserID,
		MarinaID: req.MarinaID,
	}

	queries := g.server.DB.Queries()
	err := queries.UnassignUserFromMarina(c.Request().Context(), params)
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

	// Parse and validate the request body
	req := new(requests.UpdateUserRequest)
	if err := c.Bind(req); err != nil {
		g.server.Logger.Zap.Error("Error binding request body", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Error parsing request: "+err.Error()).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		g.server.Logger.Zap.Error("Error validating request body", err)
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Get current user data to update only changed fields
	queries := g.server.DB.Queries()
	currentUser, err := queries.GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, err).JSON(c)
	}

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
	}

	// Update only the fields that were provided in the request
	if req.FirstName != nil {
		updateParams.FirstName = *req.FirstName
	}
	if req.LastName != nil {
		updateParams.LastName = *req.LastName
	}
	if req.Email != nil {
		updateParams.Email = *req.Email
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
		updateParams.PasswordHash = passwordHash
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

	// Perform update
	updatedUser, err := queries.UpdateUser(c.Request().Context(), updateParams)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.NewUserResponseSuccess(updatedUser)
	return response.JSON(c)
}
