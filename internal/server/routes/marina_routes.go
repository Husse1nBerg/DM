package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterMarinaRoutes registers all marina-related routes
func RegisterMarinaRoutes(server *s.Server, permissionProtected *echo.Group) {
	marinaHandler := h.NewMarinaHandler(server)
	contactHandler := h.NewContactHandler(server)

	// Marina routes
	marinas := permissionProtected.Group("/marinas")
	marinas.POST("", marinaHandler.CreateMarina)
	marinas.GET("", marinaHandler.GetMarinasPaginated)
	marinas.GET("/over-current-limit", marinaHandler.GetMarinasOverCurrentLimit)
	marinas.GET("/by-email", marinaHandler.GetMarinaByEmail)
	marinas.GET("/organization/:organizationId", marinaHandler.GetMarinasByOrganization)
	marinas.GET("/user/:userId", marinaHandler.GetUserMarinas)
	marinas.GET("/user", marinaHandler.GetMyUserMarinas)
	marinas.GET("/:id", marinaHandler.GetMarinaByID)
	marinas.GET("/:id/with-address", marinaHandler.GetMarinaWithAddress)
	marinas.PUT("/:id", marinaHandler.UpdateMarina)
	marinas.PUT("/:id/with-address", marinaHandler.UpdateMarinaWithAddress)
	marinas.DELETE("/:id", marinaHandler.DeleteMarina)

	// Marina-specific contact routes
	marinas.GET("/:id/contacts", contactHandler.ListContacts)
	marinas.POST("/:id/contacts", contactHandler.CreateContact)
	marinas.PUT("/:id/contacts/:contactId", contactHandler.UpdateContact)
	marinas.DELETE("/:id/contacts/:contactId", contactHandler.DeleteContact)
}
