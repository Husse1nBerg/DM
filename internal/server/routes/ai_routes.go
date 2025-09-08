package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterAIRoutes registers all AI system routes
func RegisterAIRoutes(server *s.Server, permissionProtected *echo.Group) {
	aiHandler := h.NewAIHandler(server)

	// Message routes
	ai := permissionProtected.Group("/ai")
	ai.POST("/compose-message", aiHandler.RewriteHandler)
	ai.POST("/detect-form-fields", aiHandler.DetectFormFieldsHandler)
}
