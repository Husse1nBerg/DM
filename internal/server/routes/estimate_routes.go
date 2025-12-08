package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterEstimateRoutes registers all estimate-related routes
func RegisterEstimateRoutes(server *s.Server, permissionProtected *echo.Group) {
	estimateHandler := h.NewEstimateHandler(server)

	// Estimate routes
	estimates := permissionProtected.Group("/estimates")
	estimates.GET("/customer", estimateHandler.ListEstimatesForCustomer)
	estimates.GET("/retrieve", estimateHandler.RetrieveEstimate)
	estimates.GET("/sublets", estimateHandler.ListEstimateSublets)
	estimates.GET("/search", estimateHandler.SearchEstimates)
	estimates.POST("/create", estimateHandler.CreateEstimate)
	estimates.POST("/delete-operation", estimateHandler.DeleteEstimateOperation)
	estimates.POST("/retrieve-list", estimateHandler.RetrieveEstimatesList)
	estimates.POST("/update", estimateHandler.UpdateEstimate)
	estimates.POST("/sublet", estimateHandler.SubmitEstimateSubletEntry)
	estimates.POST("/part", estimateHandler.SubmitEstimatePartEntry)
	estimates.POST("/labor", estimateHandler.SubmitEstimateLaborEntry)
	estimates.GET("/parts", estimateHandler.RetrieveEstimateParts)
	estimates.GET("/labor-entries", estimateHandler.RetrieveEstimateLabor)
	estimates.GET("/part-detail", estimateHandler.RetrieveEstimatePartDetail)
	estimates.GET("/labor-detail-records", estimateHandler.RetrieveEstimateLaborDetail)
}
