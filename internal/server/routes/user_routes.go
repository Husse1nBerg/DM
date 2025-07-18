package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterUserRoutes registers all user-related routes
func RegisterUserRoutes(server *s.Server, permissionProtected *echo.Group) {
	userHandler := h.NewUserHandler(server)

	// User routes
	users := permissionProtected.Group("/user")
	users.GET("/profile", userHandler.GetMyUserHandler)
	users.GET("/list", userHandler.ListUsersHandler)
	users.POST("", userHandler.CreateUserHandler)
	users.GET("/:userId", userHandler.GetUserHandler)
	users.PUT("/:userId", userHandler.UpdateUserHandler)
	users.DELETE("/:userId", userHandler.DeleteUserHandler)
	users.POST("/reset-password", userHandler.ResetPassword)
	users.POST("/customer-portal", userHandler.CreateCustomerUserHandler)
	users.GET("/customer-portal/:customerId", userHandler.GetUsersByCustomerIDHandler)
	users.POST("/invite", userHandler.CreateUserWithInvitationHandler)

	// User by role, organization, marina
	users.GET("/role/:roleId", userHandler.GetUsersByRoleHandler)
	users.GET("/organization/:organizationId", userHandler.GetUsersByOrganizationHandler)
	users.GET("/marina/:marinaId", userHandler.GetUsersByMarinaHandler)
	users.GET("/marina/:marinaId/assigned", userHandler.GetMarinaUsersList)
	users.GET("/marina/:marinaId/not-assigned", userHandler.GetUsersNotAssignedToMarinaHandler)

	// User-marina assignments
	users.POST("/marina/assign", userHandler.AssignUserToMarinaHandler)
	users.POST("/marina/unassign", userHandler.UnassignUserFromMarinaHandler)
} 