package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

// InventoryHandler handles inventory-related requests
type InventoryHandler struct {
	server *s.Server
}

// NewInventoryHandler creates a new inventory handler
func NewInventoryHandler(server *s.Server) *InventoryHandler {
	return &InventoryHandler{
		server: server,
	}
}

// RetrieveFuel godoc
// @Summary Retrieve fuel inventory
// @Description Retrieves one or all fuel inventory records from DockMaster
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Param fuelId query string false "Fuel ID (optional - if empty, retrieves all)"
// @Success 200 {array} responses.FuelInventoryResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/fuel [get]
func (h *InventoryHandler) RetrieveFuel(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.RetrieveFuelRequest)
	
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

	dmeResponse, err := h.server.DME.RetrieveFuel(ctx, req.FuelID, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve fuel inventory",
			zap.Error(err),
			zap.String("fuelId", req.FuelID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	responseData := make([]responses.FuelInventoryResponse, len(dmeResponse))
	for i, fuel := range dmeResponse {
		responseData[i] = responses.NewFuelInventoryResponse(fuel)
	}

	return responses.NewSuccessResponse(responseData).JSON(c)
}

// RetrieveOnlineBillcodeList godoc
// @Summary Retrieve online billing codes
// @Description Retrieves a list of billing codes that have been selected for use online
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Success 200 {array} responses.OnlineBillcodeResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/online-billcodes [get]
func (h *InventoryHandler) RetrieveOnlineBillcodeList(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.RetrieveOnlineBillcodeListRequest)
	
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

	dmeResponse, err := h.server.DME.RetrieveOnlineBillcodeList(ctx, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve online billcode list", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	responseData := make([]responses.OnlineBillcodeResponse, len(dmeResponse))
	for i, billcode := range dmeResponse {
		responseData[i] = responses.NewOnlineBillcodeResponse(billcode)
	}

	return responses.NewSuccessResponse(responseData).JSON(c)
}

// RetrieveOnlinePartsList godoc
// @Summary Retrieve online parts list
// @Description Retrieves a list of inventory records that have been selected for use online
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Success 200 {array} responses.OnlinePartResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/online-parts [get]
func (h *InventoryHandler) RetrieveOnlinePartsList(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.RetrieveOnlinePartsListRequest)
	
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

	dmeResponse, err := h.server.DME.RetrieveOnlinePartsList(ctx, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve online parts list", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	responseData := make([]responses.OnlinePartResponse, len(dmeResponse))
	for i, part := range dmeResponse {
		responseData[i] = responses.NewOnlinePartResponse(part)
	}

	return responses.NewSuccessResponse(responseData).JSON(c)
}

// RetrieveQtyInfo godoc
// @Summary Retrieve quantity information for a part
// @Description Retrieves quantity information for a specific part at a specific location
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Param partNumber query string true "Part Number"
// @Param locationCode query string true "Location Code"
// @Success 200 {object} responses.PartQtyInfoResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/qty-info [get]
func (h *InventoryHandler) RetrieveQtyInfo(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.RetrieveQtyInfoRequest)
	
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

	dmeResponse, err := h.server.DME.RetrieveQtyInfo(ctx, req.PartNumber, req.LocationCode, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve quantity info",
			zap.Error(err),
			zap.String("partNumber", req.PartNumber),
			zap.String("locationCode", req.LocationCode))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	response := responses.NewPartQtyInfoResponse(*dmeResponse)

	return responses.NewSuccessResponse(response).JSON(c)
}

// RetrievePartsKit godoc
// @Summary Retrieve a parts kit
// @Description Retrieves a parts kit record by ID
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Param kitId query string true "Kit ID"
// @Success 200 {object} responses.PartsKitResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/parts-kits [get]
func (h *InventoryHandler) RetrievePartsKit(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.RetrievePartsKitRequest)
	
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

	dmeResponse, err := h.server.DME.RetrievePartsKit(ctx, req.KitID, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve parts kit",
			zap.Error(err),
			zap.String("kitId", req.KitID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	response := responses.NewPartsKitResponse(*dmeResponse)

	return responses.NewSuccessResponse(response).JSON(c)
}

// ListPartsKits godoc
// @Summary List all parts kits
// @Description Retrieves a list of all parts kits
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Success 200 {array} responses.PartsKitResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/parts-kits/list [get]
func (h *InventoryHandler) ListPartsKits(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ListPartsKitsRequest)
	
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

	dmeResponse, err := h.server.DME.ListPartsKits(ctx, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list parts kits", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	responseData := make([]responses.PartsKitResponse, len(dmeResponse))
	for i, kit := range dmeResponse {
		responseData[i] = responses.NewPartsKitResponse(kit)
	}

	return responses.NewSuccessResponse(responseData).JSON(c)
}

// RetrievePurchaseOrder godoc
// @Summary Retrieve a purchase order
// @Description Retrieves a purchase order by its ID
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Param poId query string true "Purchase Order ID"
// @Success 200 {object} responses.PurchaseOrderResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/purchase-orders [get]
func (h *InventoryHandler) RetrievePurchaseOrder(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.RetrievePurchaseOrderRequest)
	
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

	dmeResponse, err := h.server.DME.RetrievePurchaseOrder(ctx, req.POID, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve purchase order",
			zap.Error(err),
			zap.String("poId", req.POID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	response := responses.NewPurchaseOrderResponse(*dmeResponse)

	return responses.NewSuccessResponse(response).JSON(c)
}

// ListPurchaseOrders godoc
// @Summary List purchase orders
// @Description Retrieves a list of purchase orders by single date or date range
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Param startDate query string true "Start Date (YYYY-MM-DD)"
// @Param endDate query string false "End Date (YYYY-MM-DD) - optional for date range"
// @Success 200 {array} responses.PurchaseOrderResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/purchase-orders/list [get]
func (h *InventoryHandler) ListPurchaseOrders(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ListPurchaseOrdersRequest)
	
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

	dmeResponse, err := h.server.DME.ListPurchaseOrders(ctx, req.StartDate, req.EndDate, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list purchase orders",
			zap.Error(err),
			zap.String("startDate", req.StartDate),
			zap.String("endDate", req.EndDate))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	responseData := make([]responses.PurchaseOrderResponse, len(dmeResponse))
	for i, po := range dmeResponse {
		responseData[i] = responses.NewPurchaseOrderResponse(po)
	}

	return responses.NewSuccessResponse(responseData).JSON(c)
}

// RetrieveSpecialOrder godoc
// @Summary Retrieve a special order
// @Description Retrieves a special order by ID
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Param orderId query string true "Special Order ID"
// @Success 200 {object} responses.SpecialOrderResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/special-orders [get]
func (h *InventoryHandler) RetrieveSpecialOrder(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.RetrieveSpecialOrderRequest)
	
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

	dmeResponse, err := h.server.DME.RetrieveSpecialOrder(ctx, req.OrderID, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve special order",
			zap.Error(err),
			zap.String("orderId", req.OrderID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	response := responses.NewSpecialOrderResponse(*dmeResponse)

	return responses.NewSuccessResponse(response).JSON(c)
}

// ListCustomerSpecialOrders godoc
// @Summary List customer special orders
// @Description Retrieves a list of special orders for a specific customer
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Param customerId query string true "Customer ID"
// @Success 200 {array} responses.SpecialOrderResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/special-orders/customer [get]
func (h *InventoryHandler) ListCustomerSpecialOrders(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ListCustomerSpecialOrdersRequest)
	
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

	dmeResponse, err := h.server.DME.ListCustomerSpecialOrders(ctx, req.CustomerID, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list customer special orders",
			zap.Error(err),
			zap.String("customerId", req.CustomerID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	responseData := make([]responses.SpecialOrderResponse, len(dmeResponse))
	for i, order := range dmeResponse {
		responseData[i] = responses.NewSpecialOrderResponse(order)
	}

	return responses.NewSuccessResponse(responseData).JSON(c)
}

// ListReceivedSpecialOrders godoc
// @Summary List received special orders
// @Description Retrieves a list of received special orders
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Success 200 {array} responses.SpecialOrderResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/special-orders/received [get]
func (h *InventoryHandler) ListReceivedSpecialOrders(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ListReceivedSpecialOrdersRequest)
	
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

	dmeResponse, err := h.server.DME.ListReceivedSpecialOrders(ctx, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list received special orders", zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	responseData := make([]responses.SpecialOrderResponse, len(dmeResponse))
	for i, order := range dmeResponse {
		responseData[i] = responses.NewSpecialOrderResponse(order)
	}

	return responses.NewSuccessResponse(responseData).JSON(c)
}

// SearchInventory godoc
// @Summary Search inventory
// @Description Searches the full inventory for an item
// @Tags Inventory
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Param searchString query string true "Search String"
// @Param directHit query boolean false "Direct Hit - only return if single match found" default(false)
// @Success 200 {array} responses.InventorySearchResultResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/search [get]
func (h *InventoryHandler) SearchInventory(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.SearchInventoryRequest)
	
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

	dmeResponse, err := h.server.DME.SearchInventory(ctx, req.SearchString, req.DirectHit, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to search inventory",
			zap.Error(err),
			zap.String("searchString", req.SearchString),
			zap.Bool("directHit", req.DirectHit))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	responseData := make([]responses.InventorySearchResultResponse, len(dmeResponse))
	for i, result := range dmeResponse {
		responseData[i] = responses.NewInventorySearchResultResponse(result)
	}

	return responses.NewSuccessResponse(responseData).JSON(c)
}

// FindParts godoc
// @Summary Find parts by part numbers
// @Description Attempts to find parts using various part numbers
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body requests.FindPartsRequest true "Find Parts Request"
// @Success 200 {array} responses.InventoryPartResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/find-parts [post]
func (h *InventoryHandler) FindParts(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.FindPartsRequest)
	
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

	dmeResponse, err := h.server.DME.FindParts(ctx, req.PartNumbers, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to find parts",
			zap.Error(err),
			zap.Any("partNumbers", req.PartNumbers))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	responseData := make([]responses.InventoryPartResponse, len(dmeResponse))
	for i, part := range dmeResponse {
		responseData[i] = responses.NewInventoryPartResponse(part)
	}

	return responses.NewSuccessResponse(responseData).JSON(c)
}

// RetrieveInventory godoc
// @Summary Retrieve inventory records
// @Description Retrieves inventory records from full inventory using query parameters
// @Tags Inventory
// @Accept json
// @Produce json
// @Param request body requests.RetrieveInventoryRequest true "Retrieve Inventory Request"
// @Success 200 {array} responses.InventoryPartResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /inventory/retrieve [post]
func (h *InventoryHandler) RetrieveInventory(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.RetrieveInventoryRequest)
	
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

	// Build the query object for Dockmaster API
	query := dme.RetrieveInventoryQuery{
		LocationCode:     req.LocationCode,
		LastModifiedDate: req.LastModifiedDate,
		OnHandOnly:       req.OnHandOnly,
		VendorID:         req.VendorID,
		ItemIds:          req.ItemIds,
	}

	dmeResponse, err := h.server.DME.RetrieveInventory(ctx, query, orgID, req.SystemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve inventory",
			zap.Error(err),
			zap.Any("query", query))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert to response format
	responseData := make([]responses.InventoryPartResponse, len(dmeResponse))
	for i, part := range dmeResponse {
		responseData[i] = responses.NewInventoryPartResponse(part)
	}

	return responses.NewSuccessResponse(responseData).JSON(c)
}
