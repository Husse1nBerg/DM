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

	// Add debug middleware
	ai.Use(func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			server.Logger.Zap.Info("\n\nAI Route Debug - Request received",
				"path", c.Path(),
				"method", c.Request().Method,
				"content_type", c.Request().Header.Get("Content-Type"),
				"content_length", c.Request().ContentLength)
			return next(c)
		}
	})

	ai.POST("/compose-message", aiHandler.RewriteHandler)
	ai.POST("/detect-form-fields", aiHandler.DetectFormFieldsHandler)
}
