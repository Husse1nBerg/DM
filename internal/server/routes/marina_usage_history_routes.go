package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterMarinaUsageHistoryRoutes registers all marina usage history tracking routes
func RegisterMarinaUsageHistoryRoutes(server *s.Server, protected *echo.Group) {
	marinaUsageHistoryHandler := h.NewMarinaUsageHistoryHandler(server)

	// Marina Usage History routes
	marinaUsageHistory := protected.Group("/marina-usage-history")
	marinaUsageHistory.GET("/:marinaId/date-range", marinaUsageHistoryHandler.GetMarinaUsageHistoryByDateRange)
	marinaUsageHistory.GET("/:marinaId/latest", marinaUsageHistoryHandler.GetLatestMarinaUsageHistory)
	marinaUsageHistory.GET("/:marinaId/month", marinaUsageHistoryHandler.GetMarinaUsageHistoryByMonth)
	marinaUsageHistory.GET("/all", marinaUsageHistoryHandler.GetAllMarinaUsageHistory)
	marinaUsageHistory.GET("/usage/:id", marinaUsageHistoryHandler.GetMarinaUsageHistoryByID)
	marinaUsageHistory.GET("/:marinaId", marinaUsageHistoryHandler.GetMarinaUsageHistoryByMarinaID)
}
