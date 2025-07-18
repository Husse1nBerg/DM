package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterOrganizationRoutes registers all organization-related routes
func RegisterOrganizationRoutes(server *s.Server, permissionProtected *echo.Group) {
	organizationHandler := h.NewOrganizationHandler(server)

	// Organization routes
	organizations := permissionProtected.Group("/organizations")
	organizations.POST("", organizationHandler.CreateOrganization)
	organizations.GET("", organizationHandler.GetOrganizationsPaginated)
	organizations.GET("/by-email", organizationHandler.GetOrganizationByEmail)
	organizations.GET("/:id", organizationHandler.GetOrganizationByID)
	organizations.GET("/:id/with-address", organizationHandler.GetOrganizationWithAddress)
	organizations.PUT("/:id", organizationHandler.UpdateOrganization)
	organizations.PUT("/:id/with-address", organizationHandler.UpdateOrgAddress)
	organizations.DELETE("/:id", organizationHandler.DeleteOrganization)
} 