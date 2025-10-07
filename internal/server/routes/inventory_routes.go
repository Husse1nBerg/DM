package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterInventoryRoutes registers all inventory-related routes
func RegisterInventoryRoutes(server *s.Server, base *echo.Group, permissionProtected *echo.Group) {
	inventoryHandler := h.NewInventoryHandler(server)

	// Protected inventory routes
	inventory := permissionProtected.Group("/inventory")

	// Fuel inventory
	inventory.GET("/fuel", inventoryHandler.RetrieveFuel)

	// Online billcodes and parts
	inventory.GET("/online-billcodes", inventoryHandler.RetrieveOnlineBillcodeList)
	inventory.GET("/online-parts", inventoryHandler.RetrieveOnlinePartsList)

	// Quantity information
	inventory.GET("/qty-info", inventoryHandler.RetrieveQtyInfo)

	// Parts kits
	inventory.GET("/parts-kits", inventoryHandler.RetrievePartsKit)
	inventory.GET("/parts-kits/list", inventoryHandler.ListPartsKits)

	// Purchase orders
	inventory.GET("/purchase-orders", inventoryHandler.RetrievePurchaseOrder)
	inventory.GET("/purchase-orders/list", inventoryHandler.ListPurchaseOrders)

	// Special orders
	inventory.GET("/special-orders", inventoryHandler.RetrieveSpecialOrder)
	inventory.GET("/special-orders/customer", inventoryHandler.ListCustomerSpecialOrders)
	inventory.GET("/special-orders/received", inventoryHandler.ListReceivedSpecialOrders)

	// Search and retrieve
	inventory.GET("/search", inventoryHandler.SearchInventory)
	inventory.POST("/find-parts", inventoryHandler.FindParts)
	inventory.POST("/retrieve", inventoryHandler.RetrieveInventory)
}
