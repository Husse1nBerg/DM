package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterPaymentRoutes registers all payment-related routes
func RegisterPaymentRoutes(s *s.Server, protected *echo.Group) {
	paymentHandler := h.NewPaymentHandler(s)

	// Payment routes group
	payments := protected.Group("/payments")

	// Adyen payment flow
	payments.POST("/sessions", paymentHandler.CreatePaymentSession)
	payments.POST("/details", paymentHandler.HandlePaymentRedirect)

	// Payment data retrieval
	payments.GET("", paymentHandler.ListPayments)
	payments.GET("/:id", paymentHandler.GetPaymentByID)
	payments.GET("/entity", paymentHandler.GetPaymentsByEntity)
	payments.GET("/stats", paymentHandler.GetPaymentStats)
}

// RegisterPaymentWebhookRoutes registers webhook routes (public access)
func RegisterPaymentWebhookRoutes(s *s.Server, base *echo.Group) {
	paymentHandler := h.NewPaymentHandler(s)

	// Webhook endpoint (public - no authentication required)
	base.POST("/payments/webhooks", paymentHandler.ProcessWebhook)
}
