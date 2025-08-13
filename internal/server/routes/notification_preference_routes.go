package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

func RegisterNotificationPreferenceRoutes(server *s.Server, permissionProtected *echo.Group) {
	notificationPreferenceHandler := h.NewNotificationPreferenceHandler(server)

	// Notification preference routes
	notificationPreferences := permissionProtected.Group("/notification-preference")
	notificationPreferences.GET("/list", notificationPreferenceHandler.ListNotificationPreferencesHandler)
	notificationPreferences.PUT("", notificationPreferenceHandler.UpdateNotificationPreferenceHandler)
}
