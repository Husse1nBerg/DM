package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/guard"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/notifications"
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

	// Process estimate level attachments with metadata
	type AttachmentWithPublic struct {
		FileName    string  `json:"fileName"`
		Description string  `json:"description"`
		S3Path      string  `json:"s3Path"`
		FileType    *string `json:"fileType"`
		FromDMWeb   *bool   `json:"fromDMWeb"`
		Public      bool    `json:"public"`
	}

	// Helper function to process attachments
	processAttachments := func(attachments []dme.Attachment) []AttachmentWithPublic {
		mergedMap := make(map[string]map[string]interface{})
		for _, att := range attachments {
			if att.S3Path == "" {
				continue
			}
			if _, exists := mergedMap[att.S3Path]; exists {
				continue // skip duplicates
			}
			meta, err := h.server.DB.Queries().GetAttachmentMetadataByS3Path(ctx, att.S3Path)
			public := false
			if err == nil {
				public = meta.Public
			} else if strings.Contains(err.Error(), "no rows") {
				_, _ = h.server.DB.Queries().CreateAttachmentMetadata(ctx, db.CreateAttachmentMetadataParams{
					S3Path: att.S3Path,
					Public: false,
				})
				public = false
			}
			merged := map[string]interface{}{
				"fileName":    att.FileName,
				"description": att.Description,
				"s3Path":      att.S3Path,
				"fileType":    att.FileType,
				"fromDMWeb":   att.FromDMWeb,
				"public":      public,
			}
			mergedMap[att.S3Path] = merged
		}

		result := make([]AttachmentWithPublic, 0, len(mergedMap))
		for _, v := range mergedMap {
			result = append(result, AttachmentWithPublic{
				FileName:    v["fileName"].(string),
				Description: v["description"].(string),
				S3Path:      v["s3Path"].(string),
				FileType:    v["fileType"].(*string),
				FromDMWeb:   v["fromDMWeb"].(*bool),
				Public:      v["public"].(bool),
			})
		}
		return result
	}

	// Store original operation attachments before processing
	operationAttachmentsMap := make(map[int][]dme.Attachment)
	for i := range dmeResponse.Operations {
		if len(dmeResponse.Operations[i].Attachments) > 0 {
			operationAttachmentsMap[i] = dmeResponse.Operations[i].Attachments
		}
	}

	// Process estimate level attachments
	estimateAttachments := processAttachments(dmeResponse.Attachments)

	// Marshal the estimate to a map to add merged attachments
	estimateBytes, _ := json.Marshal(dmeResponse)
	var estimateMap map[string]interface{}
	json.Unmarshal(estimateBytes, &estimateMap)

	// Set the merged estimate level attachments
	estimateMap["attachments"] = estimateAttachments

	// Process and set operation level attachments
	if operations, ok := estimateMap["operations"].([]interface{}); ok {
		for i, op := range operations {
			if opMap, ok := op.(map[string]interface{}); ok {
				if origAttachments, exists := operationAttachmentsMap[i]; exists {
					operationAttachments := processAttachments(origAttachments)
					opMap["attachments"] = operationAttachments
				}
			}
		}
	}

	// Return response with enriched attachments
	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": estimateMap,
	})
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

	// Pass empty strings for optional parameters
	dmeResponse, err := h.server.DME.ListEstimateSublets(ctx, "", "", "", orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list estimate sublets",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert the typed response to the expected format
	var genericResponse []interface{}
	for _, sublet := range dmeResponse {
		genericResponse = append(genericResponse, sublet)
	}
	
	response := responses.ConvertEstimateSublets(genericResponse)
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
		"detail":   req.Detail,
		"page":     req.Page,
		"pageSize": req.PageSize,
	}
	
	// Add optional fields if provided
	if req.Status != "" {
		listRequestData["status"] = req.Status
	}
	if req.LastUpdateDate != "" {
		listRequestData["lastUpdateDate"] = req.LastUpdateDate
	}
	if req.LastUpdateTime != "" {
		listRequestData["lastUpdateTime"] = req.LastUpdateTime
	}
	if len(req.WoIds) > 0 {
		listRequestData["woIds"] = req.WoIds
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

// @Summary Create estimate
// @Description Creates a new estimate
// @Tags Estimates
// @Accept json
// @Produce json
// @Param estimate body requests.EstimateCreateRequest true "Estimate information"
// @Success 200 {object} responses.EstimateUpdateResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/create [post]
func (h *EstimateHandler) CreateEstimate(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.EstimateCreateRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Ensure no id is provided in the payload (defensive, in case client sends extra fields)
	var raw map[string]interface{}
	if err := c.Bind(&raw); err == nil {
		if _, hasID := raw["estId"]; hasID {
			return responses.NewErrorResponse(http.StatusBadRequest, "'estId' field must not be provided when creating an estimate").JSON(c)
		}
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
	// Note: DME API expects "woId" for estimate ID (estimates are treated as work orders)
	estimateData := map[string]interface{}{
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

	dmeResponse, err := h.server.DME.CreateEstimate(ctx, estimateData, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to create estimate",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertEstimateUpdate(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Update estimate
// @Description Updates an existing estimate
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

	// Process attachments if provided
	var attachmentsForDME []dme.Attachment
	if len(req.Attachments) > 0 {
		// Convert []AttachmentWithPublic to []dme.Attachment
		attachmentsForDME = make([]dme.Attachment, 0, len(req.Attachments))
		for _, att := range req.Attachments {
			attachmentsForDME = append(attachmentsForDME, att.Attachment)
		}
		
		// Update public status in dme_attachment_metadata for each attachment
		for _, att := range req.Attachments {
			if att.S3Path == "" {
				continue
			}
			_, err := h.server.DB.Queries().UpdateAttachmentMetadataPublic(ctx, db.UpdateAttachmentMetadataPublicParams{
				S3Path: att.S3Path,
				Public: att.Public,
			})
			if err != nil && strings.Contains(err.Error(), "no rows") {
				_, createErr := h.server.DB.Queries().CreateAttachmentMetadata(ctx, db.CreateAttachmentMetadataParams{
					S3Path: att.S3Path,
					Public: att.Public,
				})
				if createErr != nil {
					h.server.Logger.DesugarZap.Error("Failed to create attachment metadata",
						zap.Error(createErr),
						zap.String("s3Path", att.S3Path))
				}
			} else if err != nil {
				h.server.Logger.DesugarZap.Error("Failed to update attachment metadata",
					zap.Error(err),
					zap.String("s3Path", att.S3Path))
			}
		}
	}

	// Convert request to map for DME API
	// Note: DME API expects "woId" for estimate ID (estimates are treated as work orders)
	estimateData := map[string]interface{}{
		"woId":            req.EstId,
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
	}
	
	// Add attachments to estimate data if provided
	if len(attachmentsForDME) > 0 {
		estimateData["attachments"] = attachmentsForDME
	}

	dmeResponse, err := h.server.DME.UpdateEstimate(ctx, estimateData, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to update estimate",
			zap.Error(err),
			zap.String("estimateId", req.EstId))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Check if any operations were approved and create notifications for internal marina staff
	operationsApproved := false
	approvedCount := 0
	var approvedTotalAmount float64
	
	if len(req.OperationCodes) > 0 {
		// Check if any operation codes have approved: true
		for _, opCode := range req.OperationCodes {
			if opCode.Approved {
				operationsApproved = true
				approvedCount++
				// Sum up the estimated charges for approved operations
				approvedTotalAmount += float64(opCode.EstimatedParts + opCode.EstimatedLabor + opCode.EstimatedEquipment + opCode.EstimatedSublet + opCode.EstimatedFreight + opCode.EstimatedMiscSupply + opCode.EstimatedMileage + opCode.EstimatedBillCodes)
			}
		}
	}

	// If operations were approved, create notifications for internal marina staff
	if operationsApproved {
		go func() {
			// Use a background context with timeout to avoid blocking the response
			notifCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Get estimate details for notification data
			estimateDetail, err := h.server.DME.EstimateRetrieve(notifCtx, req.EstId, true, orgID, *systemID)
			customerName := req.CustId // Fallback to customer ID
			var totalEstimateAmount float64
			
			if err == nil && estimateDetail != nil {
				if estimateDetail.CustomerName != "" {
					customerName = estimateDetail.CustomerName
				}
				// Calculate total from estimate detail
				if len(estimateDetail.Operations) > 0 {
					for _, op := range estimateDetail.Operations {
						totalEstimateAmount += float64(op.TotalCharges)
					}
				}
			} else {
				// Fallback: use approved operations total
				totalEstimateAmount = approvedTotalAmount
			}

			// Get all internal marina users (exclude customer users)
			isCustomer := false
			internalUsers, err := h.server.DB.Queries().GetUsersByMarina(notifCtx, db.GetUsersByMarinaParams{
				MarinaID:   user.MarinaID,
				IsCustomer: &isCustomer,
			})
			
			if err != nil {
				h.server.Logger.DesugarZap.Error("Failed to get internal marina users for notification",
					zap.Error(err),
					zap.String("estimateId", req.EstId),
					zap.String("marinaId", user.MarinaID.String()))
				return
			}

			// Initialize notification service
			notificationService := notifications.NewNotificationService(
				h.server.DB.Queries(),
				h.server.Redis,
				h.server.Logger,
				h.server.SendGrid,
				h.server.Config,
			)

			// Initialize permission service to check user permissions
			permissionService, err := guard.NewPermissionService(h.server.DB.Queries())
			if err != nil {
				h.server.Logger.DesugarZap.Error("Failed to initialize permission service for notifications",
					zap.Error(err))
				return
			}

			// Filter users by permissions and send notifications
			var notificationRequests []notifications.SmartNotificationRequest
			marinaIDStr := user.MarinaID.String()

			for _, internalUser := range internalUsers {
				// Check if user is active
				if internalUser.IsActive == nil || !*internalUser.IsActive {
					continue
				}

				// Check if user has EstimatesRead or WorkOrdersRead permission
				hasEstimatesRead, err := permissionService.CanRead(notifCtx, internalUser.ID.String(), marinaIDStr, "estimates")
				if err != nil {
					h.server.Logger.DesugarZap.Warn("Failed to check estimates.read permission",
						zap.Error(err),
						zap.String("userId", internalUser.ID.String()))
					continue
				}

				hasWorkOrdersRead, err := permissionService.CanRead(notifCtx, internalUser.ID.String(), marinaIDStr, "work_orders")
				if err != nil {
					h.server.Logger.DesugarZap.Warn("Failed to check work_orders.read permission",
						zap.Error(err),
						zap.String("userId", internalUser.ID.String()))
					continue
				}

				// User must have at least one of these permissions
				if !hasEstimatesRead && !hasWorkOrdersRead {
					continue
				}

				// Prepare notification data
				notificationData := map[string]interface{}{
					"estimateId":           req.EstId,
					"customerName":         customerName,
					"approvedOperationsCount": approvedCount,
					"approvedOperationsAmount": approvedTotalAmount,
					"totalEstimateAmount":  totalEstimateAmount,
					"approvalTimestamp":    time.Now().Format(time.RFC3339),
					"link":                 fmt.Sprintf("/service/estimates/%s", req.EstId),
				}

				notificationReq := notifications.SmartNotificationRequest{
					UserID:         internalUser.ID,
					OrganizationID: orgID,
					MarinaID:       user.MarinaID,
					Type:           "estimate_approved",
					Title:          "Estimate Approved",
					Content:        fmt.Sprintf("Customer %s has approved %d operation(s) for Estimate #%s", customerName, approvedCount, req.EstId),
					Data:           notificationData,
					Priority:       nil, // Use default priority
				}

				notificationRequests = append(notificationRequests, notificationReq)
			}

			// Send bulk notifications
			if len(notificationRequests) > 0 {
				results, err := notificationService.SendBulkSmartNotifications(notifCtx, notificationRequests)
				if err != nil {
					h.server.Logger.DesugarZap.Error("Failed to send bulk approval notifications",
						zap.Error(err),
						zap.String("estimateId", req.EstId),
						zap.Int("recipientCount", len(notificationRequests)))
				} else {
					successCount := 0
					for _, result := range results {
						if result.SystemDelivered || result.EmailDelivered {
							successCount++
						}
					}
					h.server.Logger.DesugarZap.Info("Estimate approval notifications sent",
						zap.String("estimateId", req.EstId),
						zap.Int("totalRecipients", len(notificationRequests)),
						zap.Int("successfulNotifications", successCount),
						zap.Int("approvedOperationsCount", approvedCount))
				}
			}
		}()
	}

	response := responses.ConvertEstimateUpdate(dmeResponse)
	return c.JSON(http.StatusOK, response)
}
