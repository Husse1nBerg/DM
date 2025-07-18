package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterAdminRoutes registers all admin-related routes
func RegisterAdminRoutes(server *s.Server, permissionProtected *echo.Group) {
	adminHandler := h.NewAdminHandler(server)

	// Admin routes
	admin := permissionProtected.Group("/admin")
	admin.GET("/user/marina/:marinaId", adminHandler.GetUsersByMarinaHandler)
	admin.GET("/role/list", adminHandler.ListRolesHandler)
} 