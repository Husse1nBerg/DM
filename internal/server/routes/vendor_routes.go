package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterVendorRoutes registers all DME Vendor API related routes
func RegisterVendorRoutes(server *s.Server, permissionProtected *echo.Group) {
	vendorHandler := h.NewVendorHandler(server)

	// Vendor routes group
	vendors := permissionProtected.Group("/vendors")

	// Vendor routes
	vendors.GET("", vendorHandler.RetrieveVendorHandler)
	vendors.GET("/list", vendorHandler.ListVendorsHandler)
	vendors.GET("/search", vendorHandler.SearchVendorsHandler)
}

