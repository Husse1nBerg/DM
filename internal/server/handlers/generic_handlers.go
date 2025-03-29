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
