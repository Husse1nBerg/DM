package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterAuthRoutes registers all authentication-related routes
func RegisterAuthRoutes(server *s.Server, base *echo.Group) {
	authHandler := h.NewAuthHandler(server)
	userHandler := h.NewUserHandler(server)

	// Authentication routes
	auth := base.Group("/auth")
	auth.POST("/login", authHandler.Login)
	auth.POST("/refresh", authHandler.RefreshToken)

	// Password recovery (public endpoints)
	user := base.Group("/user")
	user.POST("/forgot-password", userHandler.ForgotPassword)
	user.POST("/recover-password", userHandler.RecoverPassword)
}
