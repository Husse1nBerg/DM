package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterGeneralRoutes registers all DME General API related routes
func RegisterGeneralRoutes(server *s.Server, permissionProtected *echo.Group) {
	generalHandler := h.NewGeneralHandler(server)

	// General routes group
	general := permissionProtected.Group("/general")

	// Clerk routes
	general.GET("/clerks", generalHandler.RetrieveClerkHandler)
	general.GET("/clerks/list", generalHandler.ListClerksHandler)

	// Location routes
	general.GET("/locations", generalHandler.ListLocationsHandler)
}
