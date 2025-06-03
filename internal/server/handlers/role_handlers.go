package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/models"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type RoleHandler struct {
	server *s.Server
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
	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()

	// Get paginated roles
	params := db.GetAllRolesByTypesPaginatedParams{
		Column1: []string{"marina", "customer"},
		Limit:   pagination.PageSize,
		Offset:  (pagination.Page - 1) * pagination.PageSize,
	}
	roles, err := queries.GetAllRolesByTypesPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Get total count for pagination
	// allRoles, err := queries.GetAllRoles(c.Request().Context())
	allRoles, err := queries.GetAllRolesByTypes(c.Request().Context(), []string{"marina", "customer"})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allRoles))

	return responses.NewRolesPaginatedResponse(roles, total, pagination.PageSize, pagination.Page).JSON(c)
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
	// Parse and validate the request body
	req := new(requests.CreateRoleRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Convert permissions to bytes for database storage
	var permissionsBytes []byte
	var err error

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
		// Use default admin permissions
		defaultPermissions := models.DefaultAdminPermissions()
		permissionsBytes, err = defaultPermissions.ToBytes()
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating default permissions").JSON(c)
		}
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
		g.server.Logger.Zap.Error("Failed to bind request", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	// Start with existing values
	name := existingRole.Name
	description := existingRole.Description
	isActive := existingRole.IsActive
	permissionsBytes := existingRole.Permissions
	isCustomerRole := existingRole.IsCustomerRole
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
			g.server.Logger.Zap.Error("Failed to convert permissions", convErr)
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
	}

	// Log the update operation
	g.server.Logger.Zap.Info("Updating role",
		"id", roleID,
		"name", name,
		"description_provided", req.Description != nil,
		"permissions_provided", req.Permissions != nil,
		"is_active_provided", req.IsActive != nil,
		"is_customer_role_provided", req.IsCustomerRole != nil)

	// Update the role
	updatedRole, err := queries.UpdateRole(c.Request().Context(), updateParams)
	if err != nil {
		g.server.Logger.Zap.Error("Failed to update role", err)
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
