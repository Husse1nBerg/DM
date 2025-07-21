package routes

import (
	"github.com/dockworks/dm-web-backend/internal/config"
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterEmailRoutes registers all email sending routes
func RegisterEmailRoutes(server *s.Server, config *config.Config, permissionProtected *echo.Group) {
	emailHandler := h.NewEmailHandler(server, config)

	// Email routes
	emails := permissionProtected.Group("/email")
	emails.POST("/send-html", emailHandler.SendHTMLEmail)
	emails.POST("/send-template", emailHandler.SendTemplateEmail)
}
