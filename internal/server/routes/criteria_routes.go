package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterCriteriaRoutes registers all criteria-related routes
func RegisterCriteriaRoutes(server *s.Server, permissionProtected *echo.Group) {
	criteriaHandler := h.NewCriteriaHandler(server)

	// Criteria routes
	criteria := permissionProtected.Group("/criteria")
	criteria.GET("", criteriaHandler.ListCriteria)
	criteria.GET("/paginated", criteriaHandler.ListCriteriaPaginated)
	criteria.POST("", criteriaHandler.CreateCriteria)
	criteria.GET("/search", criteriaHandler.SearchCriteria)
	criteria.GET("/search/paginated", criteriaHandler.SearchCriteriaPaginated)
	criteria.GET("/:criteriaId", criteriaHandler.GetCriteria)
	criteria.PUT("/:criteriaId", criteriaHandler.UpdateCriteria)
	criteria.DELETE("/:criteriaId", criteriaHandler.DeleteCriteria)
	criteria.POST("/:criteriaId/duplicate", criteriaHandler.DuplicateCriteria)
}
