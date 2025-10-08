package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterGalleryRoutes registers all gallery-related routes (marina and vessel galleries)
func RegisterGalleryRoutes(server *s.Server, permissionProtected *echo.Group) {
	galleryHandler := h.NewGalleryHandler(server)

	// Gallery routes
	gallery := permissionProtected.Group("/gallery")

	// Marina gallery routes
	gallery.POST("/marina", galleryHandler.CreateMarinaGalleryItem)
	gallery.GET("/marina/:marinaId", galleryHandler.GetMarinaGallery)
	gallery.GET("/marina/item/:id", galleryHandler.GetMarinaGalleryItem)
	gallery.PUT("/marina/item/:id", galleryHandler.UpdateMarinaGalleryItem)
	gallery.DELETE("/marina/item/:id", galleryHandler.DeleteMarinaGalleryItem)

	// Vessel/Boat gallery routes
	gallery.POST("/boat", galleryHandler.CreateVesselGalleryItem)
	gallery.GET("/boat/:boatId", galleryHandler.GetVesselGallery)
	gallery.GET("/boat/item/:id", galleryHandler.GetVesselGalleryItem)
	gallery.PUT("/boat/item/:id", galleryHandler.UpdateVesselGalleryItem)
	gallery.PUT("/boat/item/:id/public", galleryHandler.UpdateVesselGalleryItemPublic)
	gallery.DELETE("/boat/item/:id", galleryHandler.DeleteVesselGalleryItem)
}
