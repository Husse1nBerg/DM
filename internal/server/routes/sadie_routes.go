package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	pm "github.com/dockworks/dm-web-backend/internal/server/middleware"
	"github.com/labstack/echo/v4"
)

// RegisterSadieRoutes registers all SADIE-related routes
func RegisterSadieRoutes(server *s.Server, router *echo.Echo) {
	sadieHandler := h.NewSadieHandler(server)
	sadieAuthMiddleware := pm.RequireSadieAuth(server.Config)

	// Public routes (no authentication)
	router.GET("/heartbeat", sadieHandler.HeartbeatGetHandler)
	router.GET("/webhooks", sadieHandler.WebhookGetHandler)
	router.POST("/webhooks", sadieHandler.WebhookPostHandler)
	router.OPTIONS("/webhooks", sadieHandler.WebhookOptionsHandler)

	// Protected routes (require x-sadie-core-secret header)
	// Root-level routes for backward compatibility (SADIE may be configured to call these)
	sadieRootProtected := router.Group("")
	sadieRootProtected.Use(sadieAuthMiddleware)
	sadieRootProtected.GET("/getAvailableSlips", sadieHandler.GetAvailableSlipsGetHandler)
	sadieRootProtected.POST("/getAvailableSlips", sadieHandler.GetAvailableSlipsPostHandler)
	sadieRootProtected.POST("/makeReservation", sadieHandler.MakeReservationPostHandler)
	sadieRootProtected.GET("/getAssistantPhoneNumber", sadieHandler.GetAssistantPhoneNumberHandler)
	sadieRootProtected.GET("/getPhoneNumbers", sadieHandler.GetPhoneNumbersHandler)
	sadieRootProtected.POST("/updateAgentWebhook", sadieHandler.UpdateAgentWebhookHandler)
	sadieRootProtected.POST("/assignPhoneNumber", sadieHandler.AssignPhoneNumberHandler)

	// Routes under /sadie prefix (alternative path)
	sadieProtected := router.Group("/sadie")
	sadieProtected.Use(sadieAuthMiddleware)
	sadieProtected.GET("/getAvailableSlips", sadieHandler.GetAvailableSlipsGetHandler)
	sadieProtected.POST("/getAvailableSlips", sadieHandler.GetAvailableSlipsPostHandler)
	sadieProtected.POST("/makeReservation", sadieHandler.MakeReservationPostHandler)
	sadieProtected.GET("/getAssistantPhoneNumber", sadieHandler.GetAssistantPhoneNumberHandler)
	sadieProtected.GET("/getPhoneNumbers", sadieHandler.GetPhoneNumbersHandler)
	sadieProtected.POST("/updateAgentWebhook", sadieHandler.UpdateAgentWebhookHandler)
	sadieProtected.POST("/assignPhoneNumber", sadieHandler.AssignPhoneNumberHandler)
}
