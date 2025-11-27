package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterCustomerRoutes registers all customer-related routes
func RegisterCustomerRoutes(server *s.Server, base *echo.Group, permissionProtected *echo.Group) {
	customerHandler := h.NewCustomerHandler(server)

	// Customer intake (public endpoint)
	base.POST("/customer-intake", customerHandler.CustomerIntake)

	// Protected customer routes
	customers := permissionProtected.Group("/customers")
	customers.GET("/list", customerHandler.ListCustomersByPage)
	customers.GET("/list-short", customerHandler.ListCustomersShortByPage)
	customers.GET("/retrieve", customerHandler.RetrieveCustomer)
	customers.GET("/retrieve-customers", customerHandler.RetrieveCustomers)
	customers.GET("/retrieve-customers-paginated", customerHandler.RetrieveCustomersPaginated)
	customers.GET("/search", customerHandler.SearchCustomers)
	customers.POST("/update", customerHandler.UpdateCustomer)
	customers.POST("/create", customerHandler.CreateCustomer)
	customers.GET("/settings", customerHandler.GetCustomerSettings)
	customers.POST("/settings", customerHandler.UpdateCustomerSettings)
} 