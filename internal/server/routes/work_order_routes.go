package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterWorkOrderRoutes registers all work order-related routes
func RegisterWorkOrderRoutes(server *s.Server, permissionProtected *echo.Group) {
	workOrderHandler := h.NewWorkOrderHandler(server)

	// Work Order routes
	workOrders := permissionProtected.Group("/work-orders")
	workOrders.GET("/list", workOrderHandler.ListWorkOrdersByPage)
	workOrders.GET("/retrieve", workOrderHandler.RetrieveWorkOrder)
	workOrders.GET("/search", workOrderHandler.SearchWorkOrders)
	workOrders.GET("/customer", workOrderHandler.ListWorkOrdersForCustomer)
	workOrders.GET("/operations", workOrderHandler.RetrieveWorkOrderOperations)
	workOrders.GET("/completed", workOrderHandler.RetrieveCompletedWorkOrders)
	workOrders.POST("/update", workOrderHandler.UpdateWorkOrder)
	workOrders.POST("/create", workOrderHandler.CreateWorkOrder)
	workOrders.POST("/create-from-estimate", workOrderHandler.CreateWorkOrderFromEstimate)
	workOrders.POST("/delete-operation", workOrderHandler.DeleteWorkOrderOperation)
} 