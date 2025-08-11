package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterDocumentRoutes registers all document management routes
func RegisterDocumentRoutes(server *s.Server, base *echo.Group, permissionProtected *echo.Group) {
	documentHandler := h.NewDocumentHandler(server)

	// Public Document routes for customer
	publicDocuments := base.Group("/public/documents")
	publicDocuments.POST("/customer", documentHandler.CustomerUploadDocumentPublic)
	publicDocuments.GET("/customer", documentHandler.CustomerGetDocumentsByEntityPublic)

	// Document routes (protected)
	documents := permissionProtected.Group("/documents")

	// Customer document routes
	documents.POST("/customer", documentHandler.CustomerUploadDocument)
	documents.GET("/customer", documentHandler.CustomerGetDocumentsByEntity)

	// Boat document routes
	documents.POST("/boat", documentHandler.BoatUploadDocument)
	documents.GET("/boat", documentHandler.BoatGetDocumentsByEntity)

	// User document routes
	documents.POST("/user", documentHandler.UserUploadDocument)
	documents.GET("/user", documentHandler.UserGetDocumentsByEntity)

	// General document routes
	documents.GET("/:id", documentHandler.GetDocument)
	documents.DELETE("/:id", documentHandler.DeleteDocument)
}
