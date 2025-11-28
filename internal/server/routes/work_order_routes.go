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
	workOrders.POST("/operations/all", workOrderHandler.RetrieveAllWorkOrderOperations)
	workOrders.POST("/operations/search", workOrderHandler.SearchAllWorkOrderOperations)
	workOrders.GET("/completed", workOrderHandler.RetrieveCompletedWorkOrders)
	workOrders.GET("/sublets", workOrderHandler.ListWorkOrderSublets)
	workOrders.GET("/group-descriptions", workOrderHandler.RetrieveWorkOrderGroupDescriptions)
	workOrders.GET("/time-entries", workOrderHandler.ListWorkOrderTimeEntries)
	workOrders.GET("/new-or-changed", workOrderHandler.ListNewOrChangedWorkOrders)
	workOrders.POST("/update", workOrderHandler.UpdateWorkOrder)
	workOrders.POST("/create", workOrderHandler.CreateWorkOrder)
	workOrders.POST("/create-from-estimate", workOrderHandler.CreateWorkOrderFromEstimate)
	workOrders.POST("/delete-operation", workOrderHandler.DeleteWorkOrderOperation)
	workOrders.POST("/submit-part-entry", workOrderHandler.SubmitWorkOrderPartEntry)
	workOrders.POST("/submit-time-entry", workOrderHandler.SubmitWorkOrderTimeEntry)
	workOrders.POST("/sublet", workOrderHandler.SubmitWorkOrderSubletEntry)
	workOrders.POST("/retrieve-list", workOrderHandler.RetrieveWorkOrdersList)
	workOrders.GET("/labor-detail", workOrderHandler.RetrieveWorkOrderLaborDetail)
	workOrders.GET("/parts", workOrderHandler.RetrieveWorkOrderParts)
} 