package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterPlanRoutes registers all plan management routes (notes/messages, storage, document plans)
func RegisterPlanRoutes(server *s.Server, protected *echo.Group) {
	planHandler := h.NewPlanHandler(server)

	// Plan routes
	plans := protected.Group("/plans")

	// Notes and Messages Plans
	notesMessagesPlans := plans.Group("/notes-messages")
	notesMessagesPlans.GET("", planHandler.ListNotesMessagesPlans)
	notesMessagesPlans.GET("/:planId", planHandler.GetNotesMessagesPlan)

	// Storage Plans
	storagePlans := plans.Group("/storage")
	storagePlans.GET("", planHandler.ListStoragePlans)
	storagePlans.GET("/:planId", planHandler.GetStoragePlan)

	// Document Plans
	documentPlans := plans.Group("/document")
	documentPlans.GET("", planHandler.ListDocumentPlans)
	documentPlans.GET("/:planId", planHandler.GetDocumentPlan)
}
