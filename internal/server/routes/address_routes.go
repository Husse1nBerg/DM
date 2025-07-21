package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterAddressRoutes registers all address management routes
func RegisterAddressRoutes(server *s.Server, permissionProtected *echo.Group) {
	addressHandler := h.NewAddressHandler(server)

	// Address routes
	addresses := permissionProtected.Group("/addresses")
	addresses.POST("", addressHandler.CreateAddress)
	addresses.GET("/:id", addressHandler.GetAddressById)
	addresses.PUT("/:id", addressHandler.UpdateAddress)
}
