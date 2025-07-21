package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterInviteRoutes registers all invitation system routes (public and protected)
func RegisterInviteRoutes(server *s.Server, base *echo.Group, permissionProtected *echo.Group) {
	inviteHandler := h.NewInviteHandler(server)

	// Public invite routes
	invite := base.Group("/invite")
	invite.GET("/confirm", inviteHandler.ConfirmToken)
	invite.POST("/accept", inviteHandler.AcceptInvitation)

	// Protected invite routes
	protectedInvite := permissionProtected.Group("/invite")
	protectedInvite.POST("/refresh", inviteHandler.RefreshInvite)
}
