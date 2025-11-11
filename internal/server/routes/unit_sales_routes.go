package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterUnitSalesRoutes registers Unit Sales related routes
func RegisterUnitSalesRoutes(server *s.Server, permissionProtected *echo.Group) {
	unitSalesHandler := h.NewUnitSalesHandler(server)

	unitSales := permissionProtected.Group("/unit-sales")
	unitSales.GET("/customer-contracts", unitSalesHandler.RetrieveCustomerContractsHandler)
}
