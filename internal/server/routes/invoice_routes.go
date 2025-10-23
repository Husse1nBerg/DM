package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterInvoiceRoutes registers all invoice-related routes
func RegisterInvoiceRoutes(server *s.Server, base *echo.Group, permissionProtected *echo.Group) {
	invoiceHandler := h.NewInvoiceHandler(server)

	// Public invoice routes (no authentication required)
	publicInvoices := base.Group("/invoices")
	publicInvoices.GET("/customer", invoiceHandler.GetCustomerInvoices)

	// Protected invoice routes (authentication required)
	protectedInvoices := permissionProtected.Group("/invoices")
	protectedInvoices.POST("/batch/submit", invoiceHandler.SubmitBatch)
}
