package routes

import (
	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

func RegisterPaymentRoutes(e *echo.Echo, cfg *config.Config, queries *db.Queries, logger *zap.Logger) error {
	handler, err := handlers.NewPaymentHandler(cfg, queries, logger)
	if err != nil {
		return err
	}

	// Create payment routes group
	g := e.Group("/api/v1/payments")

	// Register routes
	g.POST("/session", handler.CreatePaymentSession)
	g.GET("/:id", handler.GetPaymentStatus)
	g.POST("/webhook", handler.HandleWebhook)

	return nil
}
