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

	// Read optional query parameters for filtering
	estimateID := c.QueryParam("estimateId")
	opcode := c.QueryParam("opcode")
	vendorID := c.QueryParam("vendorID")

	dmeResponse, err := h.server.DME.ListEstimateSublets(ctx, estimateID, opcode, vendorID, orgID, *systemID)
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

	// Check if estimate has any approved operations
	hasApprovedOperations := false
	if len(req.OperationCodes) > 0 {
		for _, opCode := range req.OperationCodes {
			if opCode.Approved {
				hasApprovedOperations = true
				break
			}
		}
	}

	estimateID := ""
	if dmeResponse != nil {
		estimateID = dmeResponse.ID
	}
	h.server.Logger.DesugarZap.Info("Estimate creation notification check",
		zap.String("estimateId", estimateID),
		zap.Bool("hasApprovedOperations", hasApprovedOperations),
		zap.Int("operationCodesCount", len(req.OperationCodes)),
		zap.String("customerId", req.CustId))

	// If NOT approved, send notification to customer users
	if !hasApprovedOperations && dmeResponse != nil && dmeResponse.ID != "" {
		go func() {
			// Use a background context with timeout to avoid blocking the response
			notifCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			// Get customer users for this estimate
			customerID := &req.CustId
			customerUsers, err := h.server.DB.Queries().GetMarinaCustomerUsersByCustomerID(notifCtx, db.GetMarinaCustomerUsersByCustomerIDParams{
				MarinaID:   user.MarinaID,
				CustomerID: customerID,
			})

			if err != nil {
				h.server.Logger.DesugarZap.Error("Failed to get customer users for notification",
					zap.Error(err),
					zap.String("estimateId", dmeResponse.ID),
					zap.String("customerId", req.CustId),
					zap.String("marinaId", user.MarinaID.String()))
				return
			}

			if len(customerUsers) == 0 {
				h.server.Logger.DesugarZap.Warn("No customer users found for estimate notification",
					zap.String("estimateId", dmeResponse.ID),
					zap.String("customerId", req.CustId),
					zap.String("marinaId", user.MarinaID.String()))
				return
			}

			h.server.Logger.DesugarZap.Info("Found customer users for estimate notification",
				zap.String("estimateId", dmeResponse.ID),
				zap.String("customerId", req.CustId),
				zap.Int("customerUsersCount", len(customerUsers)))

			// Initialize notification service
			notificationService := notifications.NewNotificationService(
				h.server.DB.Queries(),
				h.server.Redis,
				h.server.Logger,
				h.server.SendGrid,
				h.server.Config,
			)

			// Get estimate details for notification
			estimateDetail, err := h.server.DME.EstimateRetrieve(notifCtx, dmeResponse.ID, false, orgID, *systemID)
			estimateTitle := req.Title
			if err == nil && estimateDetail != nil && estimateDetail.Title != "" {
				estimateTitle = estimateDetail.Title
			}

			// Prepare notification requests
			var notificationRequests []notifications.SmartNotificationRequest

			for _, customerUser := range customerUsers {
				// Check if user is active
				if customerUser.IsActive == nil || !*customerUser.IsActive {
					continue
				}

				// Ensure customer user has "service" notification preference
				// Check if preference exists, if not create it with defaults
				_, err := h.server.DB.Queries().GetNotificationPreference(notifCtx, db.GetNotificationPreferenceParams{
					UserID:           customerUser.ID,
					NotificationType: "service",
				})
				if err != nil {
					// Preference doesn't exist, create it with defaults for customer users
					enabled := true
					deliveryMethod := "all"
					_, createErr := h.server.DB.Queries().CreateNotificationPreference(notifCtx, db.CreateNotificationPreferenceParams{
						UserID:           customerUser.ID,
						NotificationType: "service",
						Enabled:          &enabled,
						DeliveryMethod:   &deliveryMethod,
					})
					if createErr != nil {
						h.server.Logger.DesugarZap.Warn("Failed to create service notification preference for customer user",
							zap.String("user_id", customerUser.ID.String()),
							zap.Error(createErr))
						// Continue anyway - preference might exist now or will be created later
					} else {
						h.server.Logger.DesugarZap.Debug("Created service notification preference for customer user",
							zap.String("user_id", customerUser.ID.String()))
					}
				}

				// Prepare notification data
				notificationData := map[string]interface{}{
					"estimateId": dmeResponse.ID,
					"title":      estimateTitle,
					"link":       fmt.Sprintf("/customerPortal/estimates/%s", dmeResponse.ID),
				}

				notificationReq := notifications.SmartNotificationRequest{
					UserID:         customerUser.ID,
					OrganizationID: orgID,
					MarinaID:       user.MarinaID,
					Type:           "service",
					Title:          "New Estimate Requires Approval",
					Content:        fmt.Sprintf("A new estimate has been created and requires your approval: %s", estimateTitle),
					Data:           notificationData,
					Priority:       nil, // Use default priority
				}

				notificationRequests = append(notificationRequests, notificationReq)
			}

			// Send bulk notifications
			if len(notificationRequests) > 0 {
				results, err := notificationService.SendBulkSmartNotifications(notifCtx, notificationRequests)
				if err != nil {
					h.server.Logger.DesugarZap.Error("Failed to send bulk customer notifications",
						zap.Error(err),
						zap.String("estimateId", dmeResponse.ID),
						zap.Int("recipientCount", len(notificationRequests)))
				} else {
					successCount := 0
					for _, result := range results {
						if result.SystemDelivered || result.EmailDelivered {
							successCount++
						}
					}
					h.server.Logger.DesugarZap.Info("Estimate creation notifications sent to customers",
						zap.String("estimateId", dmeResponse.ID),
						zap.Int("totalRecipients", len(notificationRequests)),
						zap.Int("successfulNotifications", successCount))
				}
			}
		}()
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

				// Ensure internal user has "service" notification preference
				// Check if preference exists, if not create it with defaults
				_, err = h.server.DB.Queries().GetNotificationPreference(notifCtx, db.GetNotificationPreferenceParams{
					UserID:           internalUser.ID,
					NotificationType: "service",
				})
				if err != nil {
					// Preference doesn't exist, create it with defaults for internal users
					enabled := true
					deliveryMethod := "all"
					_, createErr := h.server.DB.Queries().CreateNotificationPreference(notifCtx, db.CreateNotificationPreferenceParams{
						UserID:           internalUser.ID,
						NotificationType: "service",
						Enabled:          &enabled,
						DeliveryMethod:   &deliveryMethod,
					})
					if createErr != nil {
						h.server.Logger.DesugarZap.Warn("Failed to create service notification preference for internal user",
							zap.String("user_id", internalUser.ID.String()),
							zap.Error(createErr))
						// Continue anyway - preference might exist now or will be created later
					} else {
						h.server.Logger.DesugarZap.Debug("Created service notification preference for internal user",
							zap.String("user_id", internalUser.ID.String()))
					}
				}

				// Prepare notification data
				notificationData := map[string]interface{}{
					"estimateId":               req.EstId,
					"customerName":             customerName,
					"approvedOperationsCount":  approvedCount,
					"approvedOperationsAmount": approvedTotalAmount,
					"totalEstimateAmount":      totalEstimateAmount,
					"approvalTimestamp":        time.Now().Format(time.RFC3339),
					"link":                     fmt.Sprintf("/service/estimates/%s", req.EstId),
				}

				notificationReq := notifications.SmartNotificationRequest{
					UserID:         internalUser.ID,
					OrganizationID: orgID,
					MarinaID:       user.MarinaID,
					Type:           "service",
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

// @Summary Submit estimate sublet entry
// @Description Submits a sublet entry for an estimate
// @Tags Estimates
// @Accept json
// @Produce json
// @Param request body requests.SubmitEstimateSubletEntryRequest true "Sublet entry data"
// @Success 200 {object} responses.SubmitEstimateSubletEntryResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/sublet [post]
func (h *EstimateHandler) SubmitEstimateSubletEntry(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.SubmitEstimateSubletEntryRequest
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

	// Convert request to map for DME API
	subletEntry := map[string]interface{}{
		"workOrderId":         req.WorkOrderId,
		"opCode":              req.OpCode,
		"vendorId":            req.VendorId,
		"purchaseDate":        req.PurchaseDate,
		"partsPrice":          req.PartsPrice,
		"partsCost":           req.PartsCost,
		"laborPrice":          req.LaborPrice,
		"laborCost":           req.LaborCost,
		"description":         req.Description,
		"subletDiscount":      req.SubletDiscount,
		"subletLaborDiscount": req.SubletLaborDiscount,
		"locationCode":        req.LocationCode,
		"department":          req.Department,
	}

	dmeResponse, err := h.server.DME.SubmitEstimateSubletEntry(ctx, subletEntry, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to submit estimate sublet entry",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertEstimateSubmitSubletResult(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Submit estimate part entry
// @Description Submits a part entry for an estimate
// @Tags Estimates
// @Accept json
// @Produce json
// @Param request body requests.SubmitEstimatePartEntryRequest true "Part entry data"
// @Success 200 {object} responses.SubmitEstimatePartEntryResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/part [post]
func (h *EstimateHandler) SubmitEstimatePartEntry(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.SubmitEstimatePartEntryRequest
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

	// Convert request to map for DME API
	partEntry := map[string]interface{}{
		"estimateId":   req.EstimateId,
		"opCode":       req.OpCode,
		"partNumber":   req.PartNumber,
		"quantity":     req.Quantity,
		"unitPrice":    req.UnitPrice,
		"description":  req.Description,
		"locationCode": req.LocationCode,
	}

	dmeResponse, err := h.server.DME.SubmitEstimatePartEntry(ctx, partEntry, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to submit estimate part entry",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertEstimateSubmitPartResult(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve estimate parts
// @Description Retrieves a list of part entries for a specific estimate
// @Tags Estimates
// @Accept json
// @Produce json
// @Param estimatesId query string true "Estimate ID"
// @Param opcode query string false "Operation code filter"
// @Success 200 {object} responses.EstimatePartsResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/parts [get]
func (h *EstimateHandler) RetrieveEstimateParts(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.RetrieveEstimatePartsRequest)
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

	// If a specific opcode is provided, fetch parts for that opcode only
	if req.Opcode != "" {
		dmeResponse, err := h.server.DME.RetrieveEstimateParts(ctx, req.EstimatesId, req.Opcode, orgID, *systemID)
		if err != nil {
			h.server.Logger.DesugarZap.Error("Failed to retrieve estimate parts",
				zap.Error(err),
				zap.String("estimateId", req.EstimatesId),
				zap.String("opcode", req.Opcode))
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		response := responses.ConvertEstimateParts(dmeResponse)
		return c.JSON(http.StatusOK, response)
	}

	// No opcode specified - fetch estimate details to get all operations
	estimate, err := h.server.DME.EstimateRetrieve(ctx, req.EstimatesId, true, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve estimate details",
			zap.Error(err),
			zap.String("estimateId", req.EstimatesId))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve estimate details").JSON(c)
	}

	// Extract operation codes from estimate
	var allParts []dme.EstimateDetailPartEntry
	if estimate.Operations != nil {
		for _, operation := range estimate.Operations {
			if operation.Opcode == "" {
				continue
			}
			
			parts, err := h.server.DME.RetrieveEstimateParts(ctx, req.EstimatesId, operation.Opcode, orgID, *systemID)
			if err != nil {
				h.server.Logger.DesugarZap.Warn("Failed to retrieve parts for operation",
					zap.Error(err),
					zap.String("estimateId", req.EstimatesId),
					zap.String("opcode", operation.Opcode))
				continue // Skip this operation but continue with others
			}
			allParts = append(allParts, parts...)
		}
	}

	response := responses.ConvertEstimateParts(allParts)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve estimate labor
// @Description Retrieves labor entries for a specific estimate
// @Tags Estimates
// @Accept json
// @Produce json
// @Param estimatesId query string true "Estimate ID"
// @Param opcode query string false "Operation code filter"
// @Success 200 {object} responses.EstimateLaborResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/labor-entries [get]
func (h *EstimateHandler) RetrieveEstimateLabor(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.RetrieveEstimateLaborRequest)
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

	// If a specific opcode is provided, fetch labor for that opcode only
	if req.Opcode != "" {
		dmeResponse, err := h.server.DME.RetrieveEstimateLabor(ctx, req.EstimatesId, req.Opcode, orgID, *systemID)
		if err != nil {
			h.server.Logger.DesugarZap.Error("Failed to retrieve estimate labor",
				zap.Error(err),
				zap.String("estimateId", req.EstimatesId),
				zap.String("opcode", req.Opcode))
			return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
		}
		response := responses.ConvertEstimateLabor(dmeResponse)
		return c.JSON(http.StatusOK, response)
	}

	// No opcode specified - fetch estimate details to get all operations
	estimate, err := h.server.DME.EstimateRetrieve(ctx, req.EstimatesId, true, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve estimate details",
			zap.Error(err),
			zap.String("estimateId", req.EstimatesId))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve estimate details").JSON(c)
	}

	// Extract operation codes from estimate
	var allLabor []dme.LaborEntry
	if estimate.Operations != nil {
		for _, operation := range estimate.Operations {
			if operation.Opcode == "" {
				continue
			}
			
			labor, err := h.server.DME.RetrieveEstimateLabor(ctx, req.EstimatesId, operation.Opcode, orgID, *systemID)
			if err != nil {
				h.server.Logger.DesugarZap.Warn("Failed to retrieve labor for operation",
					zap.Error(err),
					zap.String("estimateId", req.EstimatesId),
					zap.String("opcode", operation.Opcode))
				continue // Skip this operation but continue with others
			}
			allLabor = append(allLabor, labor...)
		}
	}

	response := responses.ConvertEstimateLabor(allLabor)
	return c.JSON(http.StatusOK, response)
}

// @Summary Submit estimate labor entry
// @Description Submits a labor entry for an estimate
// @Tags Estimates
// @Accept json
// @Produce json
// @Param request body requests.SubmitEstimateLaborEntryRequest true "Labor entry data"
// @Success 200 {object} responses.SubmitEstimateLaborEntryResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /estimates/labor [post]
func (h *EstimateHandler) SubmitEstimateLaborEntry(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.SubmitEstimateLaborEntryRequest
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

	// Convert request to map for DME API
	laborEntry := map[string]interface{}{
		"workOrderId": req.EstimateId,
		"opCode":      req.OpCode,
		"techId":      req.TechId,
		"date":        req.Date,
	}

	// Add optional fields if provided
	if req.StartTime != "" {
		laborEntry["startTime"] = req.StartTime
	}
	if req.StopTime != "" {
		laborEntry["stopTime"] = req.StopTime
	}
	if req.Hours > 0 {
		laborEntry["hours"] = req.Hours
	}
	if req.Comments != "" {
		laborEntry["comments"] = req.Comments
	}
	if req.Department != "" {
		laborEntry["department"] = req.Department
	}
	if req.IsApproved != nil {
		laborEntry["isApproved"] = *req.IsApproved
	}
	if req.FlagLaborFinished != nil {
		laborEntry["flagLaborFinished"] = *req.FlagLaborFinished
	}

	dmeResponse, err := h.server.DME.SubmitEstimateLaborEntry(ctx, laborEntry, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to submit estimate labor entry",
			zap.Error(err),
			zap.String("estimateId", req.EstimateId))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertEstimateSubmitLaborResult(dmeResponse)
	return c.JSON(http.StatusOK, response)
}