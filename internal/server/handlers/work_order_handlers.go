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
	marinaIDStr := claims.MarinaId
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

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
	marinaIDStr := claims.MarinaId
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

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
	marinaIDStr := claims.MarinaId
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

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
	marinaIDStr := claims.MarinaId
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

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
	marinaIDStr := claims.MarinaId
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

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
	marinaIDStr := claims.MarinaId
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

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
