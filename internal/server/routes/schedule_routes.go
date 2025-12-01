package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterScheduleRoutes registers all schedule-related routes
func RegisterScheduleRoutes(server *s.Server, permissionProtected *echo.Group) {
	scheduleHandler := h.NewScheduleHandler(server)

	// Schedule routes
	schedule := permissionProtected.Group("/service/schedule")
	schedule.GET("/retrieve", scheduleHandler.RetrieveSchedule)
	schedule.GET("/retrieve-for-manager", scheduleHandler.RetrieveScheduleForManager)
	schedule.GET("/retrieve-for-tech", scheduleHandler.RetrieveScheduleForTech)
	schedule.GET("/retrieve-for-work-order", scheduleHandler.RetrieveScheduleForWorkOrder)
	schedule.GET("/work-order-schedule", scheduleHandler.RetrieveWorkOrderSchedule)
	schedule.GET("/operation-schedule", scheduleHandler.RetrieveOperationSchedule)
	schedule.GET("/labels", scheduleHandler.RetrieveScheduleLabels)
	schedule.POST("/update", scheduleHandler.UpdateSchedule)
	schedule.POST("/resolve-merge-conflict", scheduleHandler.ResolveMergeConflict)
}

