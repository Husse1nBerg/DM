package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterRoleRoutes registers all role-related routes
func RegisterRoleRoutes(server *s.Server, permissionProtected *echo.Group) {
	roleHandler := h.NewRoleHandler(server)

	// Role routes
	roles := permissionProtected.Group("/role")
	roles.GET("/list", roleHandler.ListRolesHandler)
	roles.POST("", roleHandler.CreateRoleHandler)
	roles.GET("/:roleId", roleHandler.GetRoleHandler)
	roles.PUT("/:roleId", roleHandler.UpdateRoleHandler)
	roles.DELETE("/:roleId", roleHandler.DeleteRoleHandler)
	roles.GET("/name", roleHandler.GetRoleByNameHandler)
}
