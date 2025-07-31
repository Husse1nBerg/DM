package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RoleHandler struct {
	server *s.Server
}

func (h *RoleHandler) getUserInfoFromContext(c echo.Context) (marinaID uuid.UUID, err error) {
	userToken := c.Get("user").(*jwt.Token)
	if userToken == nil {
		return uuid.Nil, responses.NewErrorResponse(http.StatusUnauthorized, "Authentication required").JSON(c)
	}

	claims, ok := userToken.Claims.(*token.JwtCustomClaims)
	if !ok {
		return uuid.Nil, responses.NewErrorResponse(http.StatusUnauthorized, "Invalid token claims").JSON(c)
	}

	if claims.ID == uuid.Nil {
		return uuid.Nil, responses.NewErrorResponse(http.StatusUnauthorized, "Invalid user ID").JSON(c)
	}

	// Fetch the user's current marina_id from the database
	queries := h.server.DB.Queries()
	user, dbErr := queries.GetUserByID(c.Request().Context(), claims.ID)
	if dbErr != nil {
		return uuid.Nil, responses.NewErrorResponse(http.StatusInternalServerError, "Failed to load user: "+dbErr.Error()).JSON(c)
	}

	if user.MarinaID == uuid.Nil {
		return uuid.Nil, responses.NewErrorResponse(http.StatusBadRequest, "User has no marina assigned").JSON(c)
	}

	marinaID = user.MarinaID
	return marinaID, nil
}

func NewRoleHandler(server *s.Server) *RoleHandler {
	return &RoleHandler{server: server}
}

// ListRolesHandler lists all existing roles
//
//	@Summary		List roles
//	@Description	Get all roles with pagination
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Success		200	{array}	responses.RoleResponse "Paginated list of roles"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/role/list [get]
func (g *RoleHandler) ListRolesHandler(c echo.Context) error {
	marinaID, err := g.getUserInfoFromContext(c)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Authentication required").JSON(c)
	}

	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	// Validate pagination params
	if pagination.Page < 1 {
		pagination.Page = 1
	}
	if pagination.PageSize < 1 || pagination.PageSize > 100 {
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()

	// Get paginated roles
	params := db.GetRolesPaginatedParams{
		MarinaID: marinaID,
		Limit:    pagination.PageSize,
		Offset:   (pagination.Page - 1) * pagination.PageSize,
	}

	roles, err := queries.GetRolesPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve roles: "+err.Error()).JSON(c)
	}

	// Get total count for pagination
	total, err := queries.CountRolesByMarina(c.Request().Context(), marinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to count roles: "+err.Error()).JSON(c)
	}

	response := responses.NewRolesPaginatedResponse(roles, total, pagination.PageSize, pagination.Page)
	return response.JSON(c)
}

// CreateRoleHandler creates a new role
//
//	@Summary		Create role
//	@Description	Create a new role
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			role	body		requests.CreateRoleRequest	true	"Role information"
//	@Success		201		{object}	responses.RoleResponse "Created role"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/role [post]
func (g *RoleHandler) CreateRoleHandler(c echo.Context) error {
	marinaID, err := g.getUserInfoFromContext(c)
	if err != nil {
		return responses.NewErrorResponse(http.StatusUnauthorized, "Authentication required").JSON(c)
	}

	// Parse and validate the request body
	req := new(requests.CreateRoleRequest)
	if err = c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err = c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Convert permissions to bytes for database storage
	var permissionsBytes []byte

	if req.Type == "" {
		req.Type = "marina"
	}
	if req.Type != "marina" && req.Type != "customer" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid role type").JSON(c)
	}

	if req.Permissions != nil {
		permissionsBytes, err = req.Permissions.ToBytes()
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error processing permissions").JSON(c)
		}
	} else {
		return responses.NewErrorResponse(http.StatusBadRequest, "Permissions are required").JSON(c)
	}

	// Set default values for nullable fields if not provided
	isActive := true
	if req.IsActive != nil {
		isActive = *req.IsActive
	}

	// Create the role
	queries := g.server.DB.Queries()
	params := db.CreateRoleParams{
		Name:           req.Name,
		Description:    req.Description,
		Permissions:    permissionsBytes,
		IsActive:       &isActive,
		IsCustomerRole: req.IsCustomerRole,
		Type:           req.Type,
		MarinaID:       marinaID,
	}

	role, err := queries.CreateRole(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewRoleResponseSuccess(role).JSON(c)
}

// GetRoleHandler gets a specific role by ID
//
//	@Summary		Get role
//	@Description	Get a specific role by ID
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			roleId	path		string	true	"Role ID"
//	@Success		200		{object}	responses.RoleResponse "Role details"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		404		{object}	responses.Error "Role not found"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/role/{roleId} [get]
func (g *RoleHandler) GetRoleHandler(c echo.Context) error {
	// Parse and validate role ID from the path
	param := new(requests.RoleIDParam)
	if err := c.Bind(param); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(param); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := g.server.DB.Queries()
	role, err := queries.GetRoleByID(c.Request().Context(), param.RoleID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, err).JSON(c)
	}

	response := responses.NewRoleResponseSuccess(role)
	response.Pretty = true
	return response.JSON(c)
}

// UpdateRoleHandler updates an existing role
//
//	@Summary		Update role
//	@Description	Update an existing role
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			roleId	path		string					true	"Role ID"
//	@Param			role	body		requests.UpdateRoleRequest	true	"Updated role information"
//	@Success		200		{object}	responses.RoleResponse "Updated role"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		404		{object}	responses.Error "Role not found"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/role/{roleId} [put]
func (g *RoleHandler) UpdateRoleHandler(c echo.Context) error {
	// Parse the role ID from the path
	idStr := c.Param("roleId")
	roleID, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid role ID format").JSON(c)
	}

	// Get the existing role
	queries := g.server.DB.Queries()
	existingRole, err := queries.GetRoleByID(c.Request().Context(), roleID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Role not found").JSON(c)
	}

	// Parse the request body
	var req requests.UpdateRoleRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	// Check if the role is assigned to any users
	if !*req.IsActive {
		usersCount, err := queries.CountUsersByRoleID(c.Request().Context(), roleID)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		if usersCount > 0 {
			return responses.NewErrorResponse(http.StatusBadRequest, "Role is assigned to users, cannot deactivate").JSON(c)
		}
	}

	// Start with existing values
	name := existingRole.Name
	description := existingRole.Description
	isActive := existingRole.IsActive
	permissionsBytes := existingRole.Permissions
	isCustomerRole := existingRole.IsCustomerRole
	marinaID := existingRole.MarinaID
	// Update fields if provided
	if req.Name != nil {
		// Check if the new name already exists for another role
		if *req.Name != existingRole.Name {
			existingRoleWithName, err := queries.GetRoleByName(c.Request().Context(), *req.Name)
			if err == nil && existingRoleWithName.ID != roleID {
				// Another role with this name exists
				return responses.NewErrorResponse(http.StatusBadRequest, "Role with this name already exists").JSON(c)
			}
		}
		name = *req.Name
	}

	if req.Description != nil {
		description = req.Description
	}

	if req.MarinaID != nil {
		marinaID = *req.MarinaID
	}

	if req.IsActive != nil {
		isActive = req.IsActive
	}

	if req.IsCustomerRole != nil {
		isCustomerRole = req.IsCustomerRole
	}

	if req.Permissions != nil {
		var convErr error
		permissionsBytes, convErr = req.Permissions.ToBytes()
		if convErr != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid permissions format").JSON(c)
		}
	}

	// Prepare update parameters
	updateParams := db.UpdateRoleParams{
		ID:             roleID,
		Name:           name,
		Description:    description,
		Permissions:    permissionsBytes,
		IsActive:       isActive,
		IsCustomerRole: isCustomerRole,
		Column8:        marinaID,
	}

	// Update the role
	updatedRole, err := queries.UpdateRole(c.Request().Context(), updateParams)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err.Error()).JSON(c)
	}

	return responses.NewRoleResponseSuccess(updatedRole).JSON(c)
}

// DeleteRoleHandler soft-deletes a role
//
//	@Summary		Delete role
//	@Description	Soft delete a role
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			roleId	path		string	true	"Role ID"
//	@Success		200		{object}	responses.BaseResponse "Success message"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		404		{object}	responses.Error "Role not found"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/role/{roleId} [delete]
func (g *RoleHandler) DeleteRoleHandler(c echo.Context) error {
	// Parse and validate role ID from the path
	param := new(requests.RoleIDParam)
	if err := c.Bind(param); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(param); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Verify role exists
	queries := g.server.DB.Queries()
	_, err := queries.GetRoleByID(c.Request().Context(), param.RoleID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Role not found").JSON(c)
	}

	// Check if the role is assigned to any users
	usersCount, err := queries.CountUsersByRoleID(c.Request().Context(), param.RoleID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	if usersCount > 0 {
		return responses.NewErrorResponse(http.StatusBadRequest, "Role is assigned to users, cannot delete").JSON(c)
	}

	// Soft delete the role
	err = queries.SoftDeleteRole(c.Request().Context(), param.RoleID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMessageResponse(http.StatusOK, "Role successfully deleted").JSON(c)
}

// GetRoleByNameHandler gets a specific role by name
//
//	@Summary		Get role by name
//	@Description	Get a specific role by name
//	@Tags			Role
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	true	"Role name"
//	@Success		200		{object}	responses.RoleResponse "Role details"
//	@Failure		400		{object}	responses.Error "Bad request"
//	@Failure		404		{object}	responses.Error "Role not found"
//	@Failure		500		{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/role/name [get]
func (g *RoleHandler) GetRoleByNameHandler(c echo.Context) error {
	name := c.QueryParam("name")
	if name == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Role name is required").JSON(c)
	}

	queries := g.server.DB.Queries()
	role, err := queries.GetRoleByName(c.Request().Context(), name)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Role not found").JSON(c)
	}

	response := responses.NewRoleResponseSuccess(role)
	response.Pretty = true
	return response.JSON(c)
}
