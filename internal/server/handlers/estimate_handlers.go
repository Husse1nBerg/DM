package handlers

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
)

// EstimateHandler handles estimate-related requests
type EstimateHandler struct {
	server *s.Server
}

// NewEstimateHandler creates a new estimate handler
func NewEstimateHandler(server *s.Server) *EstimateHandler {
	return &EstimateHandler{
		server: server,
	}
}

// @Summary List estimates for customer
// @Description Retrieves a list of basic estimate information for a specific customer
// @Tags Estimates
// @Accept json
// @Produce json
// @Param custId query string true "Customer ID"
// @Param status query string false "Status filter"
// @Param locationCodeList query string false "Location code list"
// @Success 200 {object} responses.EstimateShortListResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/customer [get]
func (h *EstimateHandler) ListEstimatesForCustomer(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.EstimatesForCustomerRequest)
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

	dmeResponse, err := h.server.DME.ListEstimatesForCustomer(ctx, req.CustId, req.Status, req.LocationCodeList, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list estimates for customer",
			zap.Error(err),
			zap.String("customerId", req.CustId))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertEstimateShortList(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve estimate
// @Description Retrieves a single estimate with detail or summary information
// @Tags Estimates
// @Accept json
// @Produce json
// @Param Id query string true "Estimate ID"
// @Param Detail query bool false "Include detail information" default(true)
// @Success 200 {object} responses.EstimateResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/retrieve [get]
func (h *EstimateHandler) RetrieveEstimate(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.EstimateRetrieveRequest)
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

	dmeResponse, err := h.server.DME.EstimateRetrieve(ctx, req.Id, req.Detail, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve estimate",
			zap.Error(err),
			zap.String("estimateId", req.Id))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertEstimate(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary List estimate sublets
// @Description Retrieves sublet purchase orders for estimates
// @Tags Estimates
// @Accept json
// @Produce json
// @Success 200 {object} responses.EstimateSubletsResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/sublets [get]
func (h *EstimateHandler) ListEstimateSublets(c echo.Context) error {
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

	dmeResponse, err := h.server.DME.ListEstimateSublets(ctx, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list estimate sublets",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertEstimateSublets(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Search estimates
// @Description Search for estimates based on provided criteria
// @Tags Estimates
// @Accept json
// @Produce json
// @Param SearchString query string true "Search string"
// @Param DirectHit query string false "Direct hit flag" Enums(true, false)
// @Success 200 {object} responses.EstimateSearchResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/search [get]
func (h *EstimateHandler) SearchEstimates(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.EstimateSearchRequest)
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

	dmeResponse, err := h.server.DME.EstimateSearch(ctx, req.SearchString, req.DirectHit, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to search estimates",
			zap.Error(err),
			zap.String("searchString", req.SearchString))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertEstimateSearch(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Delete estimate operation
// @Description Delete an operation from an estimate
// @Tags Estimates
// @Accept json
// @Produce json
// @Param WorkOrder query string true "Estimate ID"
// @Param Operation query string true "Operation code to delete"
// @Success 200 {object} responses.EstimateDeleteOperationResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/delete-operation [post]
func (h *EstimateHandler) DeleteEstimateOperation(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.EstimateDeleteOperationRequest)
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

	dmeResponse, err := h.server.DME.DeleteEstimateOperation(ctx, req.WorkOrder, req.Operation, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to delete estimate operation",
			zap.Error(err),
			zap.String("estimateId", req.WorkOrder),
			zap.String("operationCode", req.Operation))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return c.JSON(http.StatusOK, dmeResponse)
}

// @Summary Retrieve estimates list
// @Description Retrieves a list of estimates with detail or summary information
// @Tags Estimates
// @Accept json
// @Produce json
// @Param listRequest body requests.EstimateRetrieveListRequest true "List request parameters"
// @Success 200 {object} responses.EstimateListResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/retrieve-list [post]
func (h *EstimateHandler) RetrieveEstimatesList(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.EstimateRetrieveListRequest
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

	dmeResponse, err := h.server.DME.RetrieveEstimatesList(ctx, listRequestData, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve estimates list",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertEstimateList(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Update estimate
// @Description Create a new or update an existing estimate
// @Tags Estimates
// @Accept json
// @Produce json
// @Param estimate body requests.EstimateUpdateRequest true "Estimate information"
// @Success 200 {object} responses.EstimateUpdateResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/update [post]
func (h *EstimateHandler) UpdateEstimate(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.EstimateUpdateRequest
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
	estimateData := map[string]interface{}{
		"estId":           req.EstId,
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
		"operationCodes":  req.OperationCodes,
		"attachments":     req.Attachments,
	}

	dmeResponse, err := h.server.DME.UpdateEstimate(ctx, estimateData, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to update estimate",
			zap.Error(err),
			zap.String("estimateId", req.EstId))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertEstimateUpdate(dmeResponse)
	return c.JSON(http.StatusOK, response)
}
