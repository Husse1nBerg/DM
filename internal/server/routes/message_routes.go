package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterMessageRoutes registers all messaging system routes
func RegisterMessageRoutes(server *s.Server, permissionProtected *echo.Group) {
	messageHandler := h.NewMessageHandler(server)

	// Message routes
	messages := permissionProtected.Group("/message")

	// Customer message routes
	messages.POST("/customer", messageHandler.CreateMessageHandler)
	messages.GET("/customer", messageHandler.ListMessagesCustomerHandler)
	messages.PUT("/customer", messageHandler.UpdateCustomerMessageHandler)
	messages.DELETE("/customer", messageHandler.DeleteCustomerMessageHandler)

	// Marina message routes
	messages.POST("/marina", messageHandler.CreateMessageMarinaHandler)
	messages.GET("/marina", messageHandler.ListMessagesMarinaHandler)
	messages.PUT("/marina", messageHandler.UpdateMarinaMessageHandler)
	messages.DELETE("/marina", messageHandler.DeleteMarinaMessageHandler)

	// General message routes
	messages.GET("/get", messageHandler.GetMessageByIDHandler)
	messages.POST("/compose-message", messageHandler.RewriteMessageHandler)
}
