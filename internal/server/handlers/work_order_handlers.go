package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
)

// WorkOrderHandler handles work order-related requests
type WorkOrderHandler struct {
	server *s.Server
}

// NewWorkOrderHandler creates a new work order handler
func NewWorkOrderHandler(server *s.Server) *WorkOrderHandler {
	return &WorkOrderHandler{
		server: server,
	}
}

// @Summary List work orders by page
// @Description Retrieves a paginated list of work orders
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param page query int true "Page number" minimum(1)
// @Param pageSize query int true "Page size" minimum(1) maximum(100)
// @Success 200 {object} responses.WorkOrderListResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/list [get]
func (h *WorkOrderHandler) ListWorkOrdersByPage(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.WorkOrderListRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	dmeResponse, err := h.server.DME.ListWorkOrders(ctx, req.Page, req.PageSize, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list work orders",
			zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("pageSize", req.PageSize))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertWorkOrderList(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve work order by ID
// @Description Retrieves a work order by its ID
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param WorkOrderId query string true "Work Order ID"
// @Success 200 {object} responses.WorkOrderResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/retrieve [get]
func (h *WorkOrderHandler) RetrieveWorkOrder(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.WorkOrderRetrieveRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	dmeResponse, err := h.server.DME.WorkOrderRetrieve(ctx, req.Id, req.Detail, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve work order",
			zap.Error(err),
			zap.String("workOrderId", req.Id),
			zap.Bool("detail", req.Detail))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertWorkOrder(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Search work orders
// @Description Searches for work orders based on search string
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param SearchString query string true "Search string"
// @Param DirectHit query bool false "Direct hit search" default(false)
// @Success 200 {object} responses.WorkOrderSearchResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/search [get]
func (h *WorkOrderHandler) SearchWorkOrders(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.WorkOrderSearchRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	dmeResponse, err := h.server.DME.WorkOrderSearch(ctx, req.SearchString, req.DirectHit, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to search work orders",
			zap.Error(err),
			zap.String("searchString", req.SearchString),
			zap.String("directHit", req.DirectHit))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertWorkOrderSearch(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Create work order
// @Description Creates a new work order
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param workOrder body requests.WorkOrderCreateRequest true "Work Order information"
// @Success 200 {object} responses.WorkOrderResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/create [post]
func (h *WorkOrderHandler) CreateWorkOrder(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.WorkOrderCreateRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Ensure no id is provided in the payload (defensive, in case client sends extra fields)
	var raw map[string]interface{}
	if err := c.Bind(&raw); err == nil {
		if _, hasID := raw["id"]; hasID {
			return responses.NewErrorResponse(http.StatusBadRequest, "'id' field must not be provided when creating a work order").JSON(c)
		}
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	// Convert the operation codes to the right format
	operationCodes := make([]map[string]interface{}, len(req.OperationCodes))
	for i, op := range req.OperationCodes {
		// Convert struct to map
		opBytes, err := json.Marshal(op)
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Failed to marshal operation code").JSON(c)
		}

		var opMap map[string]interface{}
		if err := json.Unmarshal(opBytes, &opMap); err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Failed to unmarshal operation code").JSON(c)
		}

		operationCodes[i] = opMap
	}

	// Create a map with all the work order data
	workOrderData := map[string]interface{}{
		"woId":            req.WoId,
		"clerkId":         req.ClerkId,
		"custId":          req.CustId,
		"boatId":          req.BoatId,
		"boatName":        req.BoatName,
		"customerPhone":   req.CustomerPhone,
		"customerEmail":   req.CustomerEmail,
		"comments":        req.Comments,
		"locationCode":    req.LocationCode,
		"estCompDate":     req.EstCompDate,
		"estStartDate":    req.EstStartDate,
		"custPromiseDate": req.CustPromiseDate,
		"categoryCode":    req.CategoryCode,
		"title":           req.Title,
		"operationCodes":  operationCodes,
	}

	dmeResponse, err := h.server.DME.CreateWorkOrder(ctx, workOrderData, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to create work order",
			zap.Error(err),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertWorkOrder(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Update work order
// @Description Updates an existing work order
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param workOrder body requests.WorkOrderUpdateRequest true "Work Order information"
// @Success 200 {object} responses.WorkOrderResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/update [post]
func (h *WorkOrderHandler) UpdateWorkOrder(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.WorkOrderUpdateRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	// Convert the operation codes to the right format
	operationCodes := make([]map[string]interface{}, len(req.OperationCodes))
	for i, op := range req.OperationCodes {
		// Convert struct to map
		opBytes, err := json.Marshal(op)
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Failed to marshal operation code").JSON(c)
		}

		var opMap map[string]interface{}
		if err := json.Unmarshal(opBytes, &opMap); err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Failed to unmarshal operation code").JSON(c)
		}

		operationCodes[i] = opMap
	}

	// Create a map with all the work order data
	// Directly map all fields without conditionals, just like in CreateWorkOrder
	workOrderData := map[string]interface{}{
		"woId":            req.WoId,
		"clerkId":         req.ClerkId,
		"custId":          req.CustId,
		"boatId":          req.BoatId,
		"boatName":        req.BoatName,
		"customerPhone":   req.CustomerPhone,
		"customerEmail":   req.CustomerEmail,
		"comments":        req.Comments,
		"locationCode":    req.LocationCode,
		"estCompDate":     req.EstCompDate,
		"estStartDate":    req.EstStartDate,
		"custPromiseDate": req.CustPromiseDate,
		"categoryCode":    req.CategoryCode,
		"title":           req.Title,
		"operationCodes":  operationCodes,
	}

	dmeResponse, err := h.server.DME.UpdateWorkOrder(ctx, workOrderData, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to update work order",
			zap.Error(err),
			zap.String("workOrderId", req.WoId),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertWorkOrder(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary List work orders for customer
// @Description Retrieves a list of work orders for a specific customer
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param custId query string true "Customer ID"
// @Param status query string false "Status (O for Open, C for Closed, blank for All)"
// @Param locationCodeList query string false "Comma delimited list of location codes"
// @Success 200 {object} responses.WorkOrderShortListResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/customer [get]
func (h *WorkOrderHandler) ListWorkOrdersForCustomer(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.WorkOrdersForCustomerRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	dmeResponse, err := h.server.DME.ListWorkOrdersForCustomer(ctx, req.CustId, req.Status, req.LocationCodeList, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list work orders for customer",
			zap.Error(err),
			zap.String("customerId", req.CustId),
			zap.String("status", req.Status),
			zap.String("locationCodeList", req.LocationCodeList))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertWorkOrderShortList(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve work order operations
// @Description Retrieves available operation codes for work orders
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Success 200 {object} responses.WorkOrderOperationsResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/operations [get]
func (h *WorkOrderHandler) RetrieveWorkOrderOperations(c echo.Context) error {
	ctx := c.Request().Context()

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	// Use default pagination (page 0, pageSize 100) for backwards compatibility
	dmeResponse, err := h.server.DME.RetrieveWorkOrderOperations(ctx, 0, 100, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve work order operations",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertWorkOrderOperations(dmeResponse.Content)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve all work order operations
// @Description Retrieves a list of all Operation Codes (not filtered by USE.ONLINE)
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param request body requests.RetrieveAllOperationsRequest true "Pagination parameters"
// @Success 200 {object} responses.WorkOrderAllOperationsResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/operations/all [post]
func (h *WorkOrderHandler) RetrieveAllWorkOrderOperations(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.RetrieveAllOperationsRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	dmeResponse, err := h.server.DME.RetrieveAllWorkOrderOperations(ctx, req.Page, req.PageSize, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve all work order operations",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertWorkOrderAllOperations(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve completed work orders
// @Description Retrieves work orders completed on a specific date
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param CompleteDate query string true "Completion date (format: YYYY-MM-DD)"
// @Success 200 {object} responses.WorkOrderCompletedResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/completed [get]
func (h *WorkOrderHandler) RetrieveCompletedWorkOrders(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.WorkOrderCompletedRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	dmeResponse, err := h.server.DME.RetrieveCompletedWorkOrders(ctx, req.CompleteDate, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve completed work orders",
			zap.Error(err),
			zap.String("completeDate", req.CompleteDate))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertCompletedWorkOrders(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Create work order from estimate
// @Description Creates a new work order from an existing estimate
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param workOrder body requests.WorkOrderCreateFromEstimateRequest true "Estimate information"
// @Success 200 {object} responses.WorkOrderCreateResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/create-from-estimate [post]
func (h *WorkOrderHandler) CreateWorkOrderFromEstimate(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.WorkOrderCreateFromEstimateRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	dmeResponse, err := h.server.DME.CreateWorkOrderFromEstimate(ctx, req.EstimateId, req.WithDetail, req.WithUnapprovedOps, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to create work order from estimate",
			zap.Error(err),
			zap.String("estimateId", req.EstimateId))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Return the response directly
	return c.JSON(http.StatusOK, dmeResponse)
}

// @Summary Delete work order operation
// @Description Deletes an operation from a work order
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param WorkOrder query string true "Work Order ID"
// @Param Operation query string true "Operation Code to delete"
// @Success 200 {object} responses.WorkOrderDeleteOperationResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/delete-operation [post]
func (h *WorkOrderHandler) DeleteWorkOrderOperation(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.WorkOrderDeleteOperationRequest)

	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	// Check if systemID is nil before dereferencing
	if systemID == nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Marina system ID is not configured").JSON(c)
	}

	dmeResponse, err := h.server.DME.DeleteWorkOrderOperation(ctx, req.WorkOrder, req.Operation, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to delete work order operation",
			zap.Error(err),
			zap.String("workOrderId", req.WorkOrder),
			zap.String("operationCode", req.Operation))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Return the response directly
	return c.JSON(http.StatusOK, dmeResponse)
}

// @Summary List work order sublets
// @Description Retrieves sublet purchase orders for work orders
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Success 200 {object} responses.WorkOrderSubletsResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/sublets [get]
func (h *WorkOrderHandler) ListWorkOrderSublets(c echo.Context) error {
	ctx := c.Request().Context()

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	// Pass empty strings for optional parameters
	dmeResponse, err := h.server.DME.ListWorkOrderSublets(ctx, "", "", "", orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list work order sublets",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert the typed response to the expected format
	var genericResponse []interface{}
	for _, sublet := range dmeResponse {
		genericResponse = append(genericResponse, sublet)
	}
	
	response := responses.ConvertWorkOrderSublets(genericResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve work order group descriptions
// @Description Retrieves a list of work order group descriptions
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Success 200 {object} responses.WorkOrderGroupDescriptionsResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/group-descriptions [get]
func (h *WorkOrderHandler) RetrieveWorkOrderGroupDescriptions(c echo.Context) error {
	ctx := c.Request().Context()

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	// Get optional WorkOrderId query parameter
	workOrderID := c.QueryParam("WorkOrderId")

	dmeResponse, err := h.server.DME.RetrieveWorkOrderGroupDescriptions(ctx, workOrderID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve work order group descriptions",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert the typed response to the expected format
	var genericResponse []interface{}
	for _, desc := range dmeResponse {
		genericResponse = append(genericResponse, desc)
	}
	
	response := responses.ConvertWorkOrderGroupDescriptions(genericResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Submit work order part entry
// @Description Submits a part entry for a work order
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param partEntry body requests.WorkOrderPartEntryRequest true "Part entry information"
// @Success 200 {object} responses.WorkOrderPartEntryResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/submit-part-entry [post]
func (h *WorkOrderHandler) SubmitWorkOrderPartEntry(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.WorkOrderPartEntryRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	// Convert request to map for DME API
	partEntryData := map[string]interface{}{
		"workOrderId": req.WorkOrderID,
		"partNumber":  req.PartNumber,
		"quantity":    req.Quantity,
		"unitPrice":   req.UnitPrice,
		"description": req.Description,
		"operationId": req.OperationID,
	}

	dmeResponse, err := h.server.DME.SubmitWorkOrderPartEntry(ctx, partEntryData, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to submit work order part entry",
			zap.Error(err),
			zap.String("workOrderId", req.WorkOrderID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertWorkOrderPartEntry(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Submit work order time entry
// @Description Submits a labor time entry for a technician
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param timeEntry body requests.WorkOrderTimeEntryRequest true "Time entry information"
// @Success 200 {object} responses.WorkOrderTimeEntryResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/submit-time-entry [post]
func (h *WorkOrderHandler) SubmitWorkOrderTimeEntry(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.WorkOrderTimeEntryRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	// Convert request to map for DME API
	timeEntryData := map[string]interface{}{
		"TechId":            req.TechnicianID,
		"WorkOrderId":       req.WorkOrderID,
		"OpCode":            req.OperationID,
		"Date":              req.Date,
		"StartTime":         req.StartTime,
		"StopTime":          req.StopTime,
		"Comments":          req.Comments,
	}
	
	// Add optional fields if provided
	if req.IsApproved != nil {
		timeEntryData["IsApproved"] = *req.IsApproved
	}
	if req.FlagLaborFinished != nil {
		timeEntryData["FlagLaborFinished"] = *req.FlagLaborFinished
	}
	if req.TimeEntryUID != "" {
		timeEntryData["TimeEntryUId"] = req.TimeEntryUID
	}

	dmeResponse, err := h.server.DME.SubmitWorkOrderTimeEntry(ctx, timeEntryData, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to submit work order time entry",
			zap.Error(err),
			zap.String("workOrderId", req.WorkOrderID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertWorkOrderTimeEntry(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary List work order time entries
// @Description Retrieves time entries for work orders
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param AsOfDate query string true "As of date"
// @Param page query int true "Page number" minimum(1)
// @Param pageSize query int true "Page size" minimum(1) maximum(100)
// @Success 200 {object} responses.WorkOrderTimeEntriesResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/time-entries [get]
func (h *WorkOrderHandler) ListWorkOrderTimeEntries(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.WorkOrderListTimeEntryRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	// Call with new signature: startDate, endDate, page, pageSize, listName, detail
	dmeResponse, err := h.server.DME.ListWorkOrderTimeEntries(ctx, req.AsOfDate, "", req.Page, req.PageSize, "", false, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list work order time entries",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert typed response to interface for the response converter
	var genericResponse interface{} = dmeResponse
	response := responses.ConvertWorkOrderTimeEntries(&genericResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary List new or changed work orders
// @Description Searches for work orders created or changed as of a date
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param AsOfDate query string true "As of date"
// @Param page query int true "Page number" minimum(1)
// @Param pageSize query int true "Page size" minimum(1) maximum(100)
// @Success 200 {object} responses.WorkOrderListNewOrChangedResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/new-or-changed [get]
func (h *WorkOrderHandler) ListNewOrChangedWorkOrders(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.WorkOrderListNewOrChangedRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	// Pass empty string for optional listName parameter
	dmeResponse, err := h.server.DME.ListNewOrChangedWorkOrders(ctx, req.AsOfDate, req.Page, req.PageSize, "", orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list new or changed work orders",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertWorkOrderListNewOrChanged(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve work orders list
// @Description Retrieves a list of work orders with detail or summary information
// @Tags WorkOrders
// @Accept json
// @Produce json
// @Param listRequest body requests.WorkOrderRetrieveListRequest true "List request parameters"
// @Success 200 {object} responses.WorkOrderListResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /work-orders/retrieve-list [post]
func (h *WorkOrderHandler) RetrieveWorkOrdersList(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.WorkOrderRetrieveListRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	// Convert request to map for DME API
	listRequestData := map[string]interface{}{
		"withDetail": req.WithDetail,
		"status":     req.Status,
	}

	dmeResponse, err := h.server.DME.RetrieveWorkOrdersList(ctx, listRequestData, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve work orders list",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertWorkOrderList(dmeResponse)
	return c.JSON(http.StatusOK, response)
}