package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterPaymentTaxRoutes registers all payment tax configuration routes
func RegisterPaymentTaxRoutes(s *s.Server, protected *echo.Group) {
	paymentTaxHandler := h.NewPaymentTaxHandler(s)

	// Payment tax routes group
	paymentTax := protected.Group("/payment-tax")

	// CRUD operations
	paymentTax.POST("", paymentTaxHandler.CreatePaymentTax)
	paymentTax.GET("", paymentTaxHandler.ListPaymentTax)
	paymentTax.GET("/active", paymentTaxHandler.GetActiveMarinaPaymentTax)
	paymentTax.GET("/:id", paymentTaxHandler.GetPaymentTax)
	paymentTax.PUT("/:id", paymentTaxHandler.UpdatePaymentTax)
	paymentTax.DELETE("/:id", paymentTaxHandler.DeletePaymentTax)

	// Actions
	paymentTax.POST("/:id/deactivate", paymentTaxHandler.DeactivatePaymentTax)
	paymentTax.POST("/calculate", paymentTaxHandler.CalculateFees)
}

