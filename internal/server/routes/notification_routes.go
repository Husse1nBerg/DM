package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterNotificationRoutes registers all notification-related routes
// Note: Notifications are READ-ONLY API (notifications are created internally by services)
func RegisterNotificationRoutes(server *s.Server, permissionProtected *echo.Group) {
	notificationHandler := h.NewNotificationHandler(server)

	// Notification routes
	notifications := permissionProtected.Group("/notifications")
	notifications.GET("/list", notificationHandler.ListNotificationsHandler)
	notifications.GET("/unread-count", notificationHandler.GetUnreadCountHandler)
	notifications.GET("/stream", notificationHandler.NotificationStreamHandler)
	notifications.GET("/:id", notificationHandler.GetNotificationHandler)
	notifications.PUT("/:id/read", notificationHandler.MarkAsReadHandler)
	notifications.PUT("/:id/unread", notificationHandler.MarkAsUnreadHandler)
	notifications.PUT("/mark-all-read", notificationHandler.MarkAllAsReadHandler)
	notifications.GET("/type/:type", notificationHandler.GetNotificationsByTypeHandler)
	notifications.DELETE("/:id", notificationHandler.DeleteNotificationHandler)
}
