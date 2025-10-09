package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterInvoiceRoutes registers all invoice-related routes
func RegisterInvoiceRoutes(server *s.Server, permissionProtected *echo.Group) {
	invoiceHandler := h.NewInvoiceHandler(server)

	// Invoice routes
	invoices := permissionProtected.Group("/invoices")
	invoices.GET("/customer", invoiceHandler.GetCustomerInvoices)
	// invoices.POST("/retrieve", invoiceHandler.GetInvoicesByIDs)
	invoices.POST("/pay", invoiceHandler.InitiatePayment)
	invoices.POST("/batch/submit", invoiceHandler.SubmitBatch)
}
