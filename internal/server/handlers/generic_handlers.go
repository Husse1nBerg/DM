package handlers

import (
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/labstack/echo/v4"
)

type GenericHandler struct {
	server *s.Server
}

func NewGenericHandler(server *s.Server) *GenericHandler {
	return &GenericHandler{server: server}
}

// HealthResponse is purely for Swagger documentation
type HealthResponse struct {
	Data map[string]string `json:"data" example:"{\"status\":\"ok\",\"database\":\"connected\"}"`
}

// ProjectDetailsResponse is purely for Swagger documentation
type ProjectDetailsResponse struct {
	Data map[string]string `json:"data" example:"{\"name\":\"Marina Management System\",\"version\":\"1.0.0\"}"`
}

// healthHandler checks the health of the server
//
//	@Summary		Health check
//	@Description	Checks the health of the server
//	@Tags			Generic
//	@Accept			json
//	@Produce		json
//	@Success		200	{object} HealthResponse "Health status information"
//	@Router			/health [get]
func (g *GenericHandler) HealthHandler(c echo.Context) error {
	return responses.NewSuccessResponse(g.server.DB.Health()).JSON(c)
}

// ProjectDetailsHandler returns information about the project
//
//	@Summary		Project details
//	@Description	Returns information about the Marina Management System project
//	@Tags			Generic
//	@Accept			json
//	@Produce		json
//	@Success		200	{object} ProjectDetailsResponse "Project details information"
//	@Router			/project-details [get]
func (g *GenericHandler) ProjectDetailsHandler(c echo.Context) error {
	details := map[string]string{
		"name":        "Marina Management System",
		"version":     "1.0.0",
		"description": "A multi-tenant platform for marina management",
		"tech_stack":  "Go, Echo, PostgreSQL, SQLC",
	}
	return responses.NewSuccessResponse(details).JSON(c)
}
