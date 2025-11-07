package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterInvoiceRoutes registers all invoice-related routes
func RegisterInvoiceRoutes(server *s.Server, base *echo.Group, protected *echo.Group, permissionProtected *echo.Group) {
	invoiceHandler := h.NewInvoiceHandler(server)

	// Protected invoice routes (authentication required)
	protectedInvoices := permissionProtected.Group("/invoices")
	protectedInvoices.POST("/batch/submit", invoiceHandler.SubmitBatch)
	protectedInvoices.GET("/next-reference", invoiceHandler.GetNextReference)

	// Authorization-flexible: allow Bearer or payment token
	invoices := protected.Group("/invoices")
	invoices.GET("/customer", invoiceHandler.GetCustomerInvoices)
}
