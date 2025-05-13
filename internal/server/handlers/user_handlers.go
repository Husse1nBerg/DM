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
	"github.com/dockworks/dm-web-backend/pkg/models"
	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
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

	// Convert permissions and modules to bytes
	var permissionsBytes, modulesBytes []byte

	// Use provided permissions or default from role
	if req.Permissions != nil {
		var err error
		permissionsBytes, err = req.Permissions.ToBytes()
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid permissions format").JSON(c)
		}
	} else {
		// Set default permissions based on the role
		role, err := queries.GetRoleByID(c.Request().Context(), req.RoleID)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		permissionsBytes = role.Permissions
	}

	// Use provided modules or default read-only modules
	if req.Modules != nil {
		var err error
		modulesBytes, err = req.Modules.ToBytes()
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid modules format").JSON(c)
		}
	} else {
		// Set default read-only modules if not provided
		defaultModules := models.ReadOnlyModules()
		modulesBytes, _ = defaultModules.ToBytes()
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
		Modules:             modulesBytes,
		Permissions:         permissionsBytes,
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
		Modules:             currentUser.Modules,
		Permissions:         currentUser.Permissions,
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

		// Process other form fields regardless of whether an image was uploaded
		if firstName := c.FormValue("firstName"); firstName != "" {
			updateParams.FirstName = firstName
		}
		if lastName := c.FormValue("lastName"); lastName != "" {
			updateParams.LastName = lastName
		}
		if email := c.FormValue("email"); email != "" {
			updateParams.Email = email
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
			updateParams.PasswordHash = passwordHash
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

		// Update permissions if provided
		if req.Permissions != nil {
			permissionsBytes, err := req.Permissions.ToBytes()
			if err != nil {
				return responses.NewErrorResponse(http.StatusBadRequest, "Invalid permissions format").JSON(c)
			}
			updateParams.Permissions = permissionsBytes
		}

		// Update modules if provided
		if req.Modules != nil {
			modulesBytes, err := req.Modules.ToBytes()
			if err != nil {
				return responses.NewErrorResponse(http.StatusBadRequest, "Invalid modules format").JSON(c)
			}
			updateParams.Modules = modulesBytes
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
	if err := utils.VerifyPassword(user.PasswordHash, resetRequest.OldPassword); err != nil {
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
	historyHashes = append(historyHashes, user.PasswordHash)  // Add current password to history
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
		PasswordHash: user.PasswordHash, // Store the old password that's being replaced
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
		PasswordHash:        newPasswordHash,
		LastLogin:           user.LastLogin,
		FailedLoginAttempts: user.FailedLoginAttempts,
		LockedUntil:         user.LockedUntil,
		LastPasswordReset:   lastPasswordReset,
		MarinaID:            user.MarinaID,
		RoleID:              user.RoleID,
		IsSuperuser:         user.IsSuperuser,
		IsActive:            user.IsActive,
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
	historyHashes = append(historyHashes, user.PasswordHash)  // Add current password to history
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
		PasswordHash: user.PasswordHash, // Store the old password that's being replaced
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
		PasswordHash:        newPasswordHash,
		LastLogin:           user.LastLogin,
		FailedLoginAttempts: user.FailedLoginAttempts,
		LockedUntil:         user.LockedUntil,
		LastPasswordReset:   lastPasswordReset,
		MarinaID:            user.MarinaID,
		RoleID:              user.RoleID,
		IsSuperuser:         user.IsSuperuser,
		IsActive:            user.IsActive,
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
	baseURL := "https://app.dockmaster.com" // Default URL
	resetURL := fmt.Sprintf("%s/reset-password?token=%s&email=%s",
		baseURL,
		token,
		url.QueryEscape(user.Email))

	// Create template data
	templateData := sendgrid.PasswordResetTemplateData{
		FirstName:   user.FirstName,
		ResetURL:    resetURL,
		Token:       token,
		Email:       user.Email,
		ExpiresIn:   "24 hours",
		CompanyName: "Dockmaster",
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
