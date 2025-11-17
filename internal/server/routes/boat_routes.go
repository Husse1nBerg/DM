package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterBoatRoutes registers all boat/vessel management routes
func RegisterBoatRoutes(server *s.Server, permissionProtected *echo.Group) {
	boatHandler := h.NewBoatHandler(server)

	// Boat routes
	boats := permissionProtected.Group("/boats")
	boats.GET("/list", boatHandler.ListBoatsByPage)
	boats.GET("/list-new-or-changed", boatHandler.ListBoatsNewOrChanged)
	boats.GET("/retrieve", boatHandler.RetrieveBoat)
	boats.GET("/retrieve-boats", boatHandler.RetrieveBoats)
	boats.GET("/customer", boatHandler.RetrieveBoatsForCustomer)
	boats.GET("/search", boatHandler.SearchBoats)
	boats.POST("/update", boatHandler.UpdateBoat)
	boats.POST("/create", boatHandler.CreateBoat)
}
