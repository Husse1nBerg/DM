package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterPaymentTaxRoutes registers all payment tax configuration routes
func RegisterPaymentTaxRoutes(s *s.Server, base *echo.Group, protected *echo.Group) {
	paymentTaxHandler := h.NewPaymentTaxHandler(s)

	// Public payment tax routes (no authentication required)
	publicPaymentTax := base.Group("/payment-tax")
	publicPaymentTax.POST("/calculate", paymentTaxHandler.CalculateFees)

	// Protected payment tax routes (authentication required)
	paymentTax := protected.Group("/payment-tax")

	// CRUD operations
	paymentTax.POST("", paymentTaxHandler.CreatePaymentTax)
	paymentTax.GET("", paymentTaxHandler.ListPaymentTax)
	paymentTax.GET("/by-type", paymentTaxHandler.GetMarinaPaymentTaxByType)
	paymentTax.GET("/:id", paymentTaxHandler.GetPaymentTax)
	paymentTax.PUT("/:id", paymentTaxHandler.UpdatePaymentTax)
	paymentTax.DELETE("/:id", paymentTaxHandler.DeletePaymentTax)
}
