package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterServiceRoutes registers all general service-related routes
func RegisterServiceRoutes(server *s.Server, permissionProtected *echo.Group) {
	serviceHandler := h.NewServiceHandler(server)

	// Service routes
	service := permissionProtected.Group("/service")
	service.GET("/opcodes/new-or-changed", serviceHandler.ListNewOrChangedOpCodes)
	service.GET("/work-order-category-codes", serviceHandler.ListWOCategoryCodes)
	service.GET("/operation-category-codes", serviceHandler.ListOPCategoryCodes)
	service.GET("/operation-descriptions", serviceHandler.RetrieveOperationDescriptions)
	service.GET("/technicians", serviceHandler.ListTechnicians)
}
