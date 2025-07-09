package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

type AdminHandler struct {
	server *s.Server
}

func NewAdminHandler(server *s.Server) *AdminHandler {
	return &AdminHandler{
		server: server,
	}
}

// GetUsersByMarinaHandler gets users by marina ID
//
//	@Summary		Get users by marina
//	@Description	Get all users in a specific marina
//	@Tags			Admin
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
//	@Router			/admin/user/marina/{marinaId} [get]
func (g *AdminHandler) GetUsersByMarinaHandler(c echo.Context) error {
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

	params := db.ListUserMarinasAssignmentsPaginatedAdminOnlyParams{
		MarinaID: marinaID,
		Limit:    pagination.PageSize,
		Offset:   (pagination.Page - 1) * pagination.PageSize,
	}
	rows, err := queries.ListUserMarinasAssignmentsPaginatedAdminOnly(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Map to response DTOs
	userResponses := make([]responses.UserResponse, len(rows))
	for i, row := range rows {
		response := responses.NewUserResponseFromUserMarinasAssignmentRowAdminOnly(row, g.server)
		if response != nil {
			userResponses[i] = *response
		}
	}

	total, err := queries.CountUserMarinasAssignmentsPaginatedAdminOnly(c.Request().Context(), params.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewPaginatedResponse(userResponses, total, pagination.PageSize, pagination.Page).JSON(c)
}

// ListRolesHandler lists all existing roles
//
//	@Summary		List roles
//	@Description	Get all roles with pagination
//	@Tags			Admin
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Success		200	{array}	responses.RoleResponse "Paginated list of roles"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/admin/role/list [get]
func (g *AdminHandler) ListRolesHandler(c echo.Context) error {
	// Parse pagination params
	pagination := new(requests.PaginationQuery)
	if err := c.Bind(pagination); err != nil {
		pagination.Page = 1
		pagination.PageSize = 10
	}

	queries := g.server.DB.Queries()

	// Get paginated roles
	params := db.GetAllRolesByTypesPaginatedParams{
		Column1: []string{"marina", "customer", "internal"},
		Limit:   pagination.PageSize,
		Offset:  (pagination.Page - 1) * pagination.PageSize,
	}
	roles, err := queries.GetAllRolesByTypesPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	total := int64(len(roles))

	return responses.NewRolesPaginatedResponse(roles, total, pagination.PageSize, pagination.Page).JSON(c)
}
