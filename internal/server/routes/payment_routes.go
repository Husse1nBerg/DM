package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterPaymentRoutes registers all payment-related routes
func RegisterPaymentRoutes(s *s.Server, base *echo.Group, protected *echo.Group) {
	paymentHandler := h.NewPaymentHandler(s)

	// Public payment routes (no authentication required)
	publicPayments := base.Group("/payments")
	publicPayments.POST("/sessions", paymentHandler.CreatePaymentSession)
	publicPayments.POST("/webhooks", paymentHandler.ProcessWebhook)

	// Protected payment routes (authentication required)
	payments := protected.Group("/payments")

	// Adyen payment flow
	payments.POST("/details", paymentHandler.HandlePaymentRedirect)

	// Payment data retrieval
	payments.GET("", paymentHandler.ListPayments)
	payments.GET("/:id", paymentHandler.GetPaymentByID)
	payments.GET("/entity", paymentHandler.GetPaymentsByEntity)
	payments.GET("/stats", paymentHandler.GetPaymentStats)
}
