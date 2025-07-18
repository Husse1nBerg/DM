package routes

import (
	"github.com/dockworks/dm-web-backend/internal/config"
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterSMSRoutes registers all SMS sending routes
func RegisterSMSRoutes(server *s.Server, config *config.Config, permissionProtected *echo.Group) {
	smsHandler := h.NewSMSHandler(server, config)

	// SMS routes
	sms := permissionProtected.Group("/sms")
	sms.POST("/send", smsHandler.SendSMS)
}
