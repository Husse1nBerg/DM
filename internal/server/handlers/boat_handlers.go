package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"strings"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/token"
)

// BoatHandler handles boat-related requests
type BoatHandler struct {
	server *s.Server
}

// NewBoatHandler creates a new boat handler
func NewBoatHandler(server *s.Server) *BoatHandler {
	return &BoatHandler{
		server: server,
	}
}

// @Summary List boats by page
// @Description Retrieves a paginated list of boats (minimal fields)
// @Tags Boats
// @Accept json
// @Produce json
// @Param page query int true "Page number" minimum(1)
// @Param pageSize query int true "Page size" minimum(1) maximum(100)
// @Success 200 {object} responses.BoatListMinimalResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/list [get]
func (h *BoatHandler) ListBoatsByPage(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.BoatsListRequest)
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

	dmeResponse, err := h.server.DME.BoatsList(ctx, req.Page, req.PageSize, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list boats",
			zap.Error(err),
			zap.Int("page", req.Page),
			zap.Int("pageSize", req.PageSize))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert response to API response format
	response := responses.ConvertBoatListMinimal(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve boat by ID
// @Description Retrieves a boat by its ID
// @Tags Boats
// @Accept json
// @Produce json
// @Param BoatId query string true "Boat ID"
// @Success 200 {object} responses.BoatResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/retrieve [get]
func (h *BoatHandler) RetrieveBoat(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.BoatRetrieveRequest)
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

	dmeResponse, err := h.server.DME.RetrieveBoatByID(ctx, req.BoatID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve boat",
			zap.Error(err),
			zap.String("boatId", req.BoatID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Deduplicate and merge DME attachments with local metadata
	mergedMap := make(map[string]map[string]interface{})
	for _, att := range dmeResponse.Attachments {
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
	// Convert mergedMap to a slice and to a struct with the public field
	type AttachmentWithPublic struct {
		FileName    string  `json:"fileName"`
		Description string  `json:"description"`
		S3Path      string  `json:"s3Path"`
		FileType    *string `json:"fileType"`
		FromDMWeb   *bool   `json:"fromDMWeb"`
		Public      bool    `json:"public"`
	}
	mergedAttachments := make([]AttachmentWithPublic, 0, len(mergedMap))
	for _, v := range mergedMap {
		mergedAttachments = append(mergedAttachments, AttachmentWithPublic{
			FileName:    v["fileName"].(string),
			Description: v["description"].(string),
			S3Path:      v["s3Path"].(string),
			FileType:    v["fileType"].(*string),
			FromDMWeb:   v["fromDMWeb"].(*bool),
			Public:      v["public"].(bool),
		})
	}
	// Set the merged attachments inside the boat object
	dmeResponse.Attachments = nil // clear original
	response := responses.ConvertBoat(dmeResponse)

	// Marshal the DME boat to a map
	boatBytes, _ := json.Marshal(response.Data)
	var boatMap map[string]interface{}
	json.Unmarshal(boatBytes, &boatMap)

	// Set the merged attachments
	boatMap["attachments"] = mergedAttachments

	return c.JSON(http.StatusOK, map[string]interface{}{
		"data": boatMap,
	})
}

// @Summary Retrieve boats for customer
// @Description Retrieves boats associated with a customer
// @Tags Boats
// @Accept json
// @Produce json
// @Param CustomerId query string true "Customer ID"
// @Success 200 {object} responses.BoatSearchResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/customer [get]
func (h *BoatHandler) RetrieveBoatsForCustomer(c echo.Context) error {
	ctx := c.Request().Context()
	customerID := c.QueryParam("CustomerId")
	if customerID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Customer ID is required").JSON(c)
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

	// Get boats for the customer
	boats, err := h.server.DME.RetrieveBoatsForCustomer(ctx, customerID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve boats for customer",
			zap.Error(err),
			zap.String("customerId", customerID))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Create a response with the list of boats
	response := responses.ConvertBoatList(boats, 1, 1, len(boats))
	return c.JSON(http.StatusOK, response)
}

// @Summary Search boats
// @Description Searches for boats based on search string
// @Tags Boats
// @Accept json
// @Produce json
// @Param SearchString query string true "Search string"
// @Param DirectHit query bool false "Direct hit search" default(false)
// @Success 200 {object} responses.BoatSearchResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/search [get]
func (h *BoatHandler) SearchBoats(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.BoatSearchRequest)
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

	dmeResponse, err := h.server.DME.SearchBoats(ctx, req.SearchString, req.DirectHit, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to search boats",
			zap.Error(err),
			zap.String("searchString", req.SearchString),
			zap.Bool("directHit", req.DirectHit))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Convert DME response to API response
	response := responses.ConvertBoatSearch(&dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Update boat
// @Description Updates a boat's information
// @Tags Boats
// @Accept json
// @Produce json
// @Param boat body dme.BoatUpdate true "Boat information"
// @Success 200 {object} responses.BoatResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/update [post]
func (h *BoatHandler) UpdateBoat(c echo.Context) error {
	ctx := c.Request().Context()
	var reqStruct requests.BoatUpdateRequest
	if err := c.Bind(&reqStruct); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := c.Validate(&reqStruct); err != nil {
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

	// Fetch the existing boat from DME
	existingBoat, err := h.server.DME.RetrieveBoatByID(ctx, reqStruct.ID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("RetrieveBoatByID error in UpdateBoat", zap.Error(err), zap.String("boatId", reqStruct.ID))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to fetch existing boat: "+err.Error()).JSON(c)
	}

	// Overlay the incoming changes onto the existing boat
	// Only update fields that are present in the request
	if reqStruct.Name != nil {
		existingBoat.Name = *reqStruct.Name
	}
	if reqStruct.Registration != nil {
		existingBoat.Registration = *reqStruct.Registration
	}
	if reqStruct.Year != nil {
		existingBoat.Year = *reqStruct.Year
	}
	if reqStruct.Make != nil {
		existingBoat.Make = *reqStruct.Make
	}
	if reqStruct.Model != nil {
		existingBoat.Model = *reqStruct.Model
	}
	if reqStruct.HIN != nil {
		existingBoat.HIN = *reqStruct.HIN
	}
	if reqStruct.LOA != nil {
		existingBoat.LOA = *reqStruct.LOA
	}
	if reqStruct.LWL != nil {
		existingBoat.LWL = *reqStruct.LWL
	}
	if reqStruct.Draft != nil {
		existingBoat.Draft = *reqStruct.Draft
	}
	if reqStruct.Beam != nil {
		existingBoat.Beam = *reqStruct.Beam
	}
	if reqStruct.Height != nil {
		existingBoat.Height = *reqStruct.Height
	}
	if reqStruct.Color != nil {
		existingBoat.Color = *reqStruct.Color
	}
	if reqStruct.TrailerMake != nil {
		existingBoat.TrailerMake = *reqStruct.TrailerMake
	}
	if reqStruct.TrailerModel != nil {
		existingBoat.TrailerModel = *reqStruct.TrailerModel
	}
	if reqStruct.TrailerSerial != nil {
		existingBoat.TrailerSerial = *reqStruct.TrailerSerial
	}
	if reqStruct.TrailerRegistration != nil {
		existingBoat.TrailerRegistration = *reqStruct.TrailerRegistration
	}
	if reqStruct.TrailerLocation != nil {
		existingBoat.TrailerLocation = *reqStruct.TrailerLocation
	}
	if reqStruct.SummerSlip != nil {
		existingBoat.SummerSlip = *reqStruct.SummerSlip
	}
	if reqStruct.WinterSlip != nil {
		existingBoat.WinterSlip = *reqStruct.WinterSlip
	}
	if reqStruct.InsuranceCompany != nil {
		existingBoat.InsuranceCompany = *reqStruct.InsuranceCompany
	}
	if reqStruct.InsuranceExpDate != nil {
		existingBoat.InsuranceExpDate = *reqStruct.InsuranceExpDate
	}
	if reqStruct.SlipID != nil {
		existingBoat.SlipID = *reqStruct.SlipID
	}
	if reqStruct.Comments != nil {
		existingBoat.Comments = *reqStruct.Comments
	}
	if reqStruct.IntegrationID != nil {
		existingBoat.IntegrationID = *reqStruct.IntegrationID
	}
	if reqStruct.OwnerIntegrationID != nil {
		existingBoat.OwnerIntegrationID = *reqStruct.OwnerIntegrationID
	}
	if reqStruct.LastModified != nil {
		existingBoat.LastModified = *reqStruct.LastModified
	}
	if reqStruct.ContractStartDate != nil {
		existingBoat.ContractStartDate = *reqStruct.ContractStartDate
	}
	if reqStruct.ContractEndDate != nil {
		existingBoat.ContractEndDate = *reqStruct.ContractEndDate
	}
	if reqStruct.TransomType != nil {
		existingBoat.TransomType = *reqStruct.TransomType
	}
	if reqStruct.TransomHeight != nil {
		existingBoat.TransomHeight = *reqStruct.TransomHeight
	}
	if reqStruct.TransomMaterial != nil {
		existingBoat.TransomMaterial = *reqStruct.TransomMaterial
	}
	if reqStruct.TransomCondition != nil {
		existingBoat.TransomCondition = *reqStruct.TransomCondition
	}
	if reqStruct.Access != nil {
		existingBoat.Access = *reqStruct.Access
	}
	if reqStruct.Attachments != nil {
		// Convert []AttachmentWithPublic to []dme.Attachment
		attachments := make([]dme.Attachment, 0, len(reqStruct.Attachments))
		for _, att := range reqStruct.Attachments {
			attachments = append(attachments, att.Attachment)
		}
		existingBoat.Attachments = attachments
		// Update public status in dme_attachment_metadata for each attachment
		for _, att := range reqStruct.Attachments {
			if att.S3Path == "" {
				continue
			}
			if att.FromDMWeb != nil && *att.FromDMWeb {
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
	existingBoat.DoNotLaunch = *reqStruct.DoNotLaunch
	if reqStruct.Motors != nil {
		existingBoat.Motors = reqStruct.Motors
	}
	if reqStruct.Drives != nil {
		existingBoat.Drives = reqStruct.Drives
	}
	if reqStruct.Generators != nil {
		existingBoat.Generators = reqStruct.Generators
	}
	// BillingCodes are intentionally set to empty array to clear existing billing configuration
	existingBoat.BillingCodes = []dme.BillingCode{}
	// BillingCodes are intentionally NOT updated to preserve existing billing configuration
	// if reqStruct.BillingCodes != nil {
	//	existingBoat.BillingCodes = reqStruct.BillingCodes
	// }

	if reqStruct.BoatDescriptionCodes != nil {
		existingBoat.BoatDescriptionCodes = reqStruct.BoatDescriptionCodes
	}
	if reqStruct.CustomInformation != nil {
		existingBoat.CustomInformation = reqStruct.CustomInformation
	}
	if reqStruct.OperationsHistory != nil {
		existingBoat.OperationsHistory = reqStruct.OperationsHistory
	}
	if (reqStruct.Slip != dme.Slip{}) {
		existingBoat.Slip = reqStruct.Slip
	}

	// Debug: Log the merged boat object before sending to DME
	h.server.Logger.DesugarZap.Debug("Merged boat object before DME.UpdateBoat", zap.Any("boatUpdate", existingBoat))

	// Convert to BoatUpdate for DME
	boatUpdate := dme.BoatUpdate{
		ID:                   existingBoat.ID,
		Name:                 existingBoat.Name,
		Registration:         existingBoat.Registration,
		Year:                 existingBoat.Year,
		Make:                 existingBoat.Make,
		Model:                existingBoat.Model,
		HIN:                  existingBoat.HIN,
		LOA:                  existingBoat.LOA,
		LWL:                  existingBoat.LWL,
		Draft:                existingBoat.Draft,
		Beam:                 existingBoat.Beam,
		Height:               existingBoat.Height,
		Color:                existingBoat.Color,
		TrailerMake:          existingBoat.TrailerMake,
		TrailerModel:         existingBoat.TrailerModel,
		TrailerSerial:        existingBoat.TrailerSerial,
		TrailerRegistration:  existingBoat.TrailerRegistration,
		TrailerLocation:      existingBoat.TrailerLocation,
		SummerSlip:           existingBoat.SummerSlip,
		WinterSlip:           existingBoat.WinterSlip,
		InsuranceCompany:     existingBoat.InsuranceCompany,
		InsuranceExpDate:     existingBoat.InsuranceExpDate,
		SlipID:               existingBoat.SlipID,
		Slip:                 existingBoat.Slip,
		Motors:               existingBoat.Motors,
		Drives:               existingBoat.Drives,
		Generators:           existingBoat.Generators,
		DoNotLaunch:          existingBoat.DoNotLaunch,
		BillingCodes:         existingBoat.BillingCodes,
		BoatDescriptionCodes: existingBoat.BoatDescriptionCodes,
		CustomInformation:    existingBoat.CustomInformation,
		OperationsHistory:    existingBoat.OperationsHistory,
		IntegrationID:        existingBoat.IntegrationID,
		OwnerIntegrationID:   existingBoat.OwnerIntegrationID,
		LastModified:         existingBoat.LastModified,
		Comments:             existingBoat.Comments,
		ContractStartDate:    existingBoat.ContractStartDate,
		ContractEndDate:      existingBoat.ContractEndDate,
		TransomType:          existingBoat.TransomType,
		TransomTypeDesc:      existingBoat.TransomTypeDesc,
		TransomHeight:        existingBoat.TransomHeight,
		TransomMaterial:      existingBoat.TransomMaterial,
		TransomCondition:     existingBoat.TransomCondition,
		Access:               existingBoat.Access,
		Attachments:          existingBoat.Attachments,
	}
	dmeResponse, err := h.server.DME.UpdateBoat(ctx, &boatUpdate, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("DME.UpdateBoat error in UpdateBoat", zap.Error(err), zap.Any("boatUpdate", boatUpdate))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertBoat(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Create boat
// @Description Creates a new boat
// @Tags Boats
// @Accept json
// @Produce json
// @Param boat body requests.BoatCreateRequest true "Boat information"
// @Success 200 {object} responses.BoatResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/create [post]
func (h *BoatHandler) CreateBoat(c echo.Context) error {
	ctx := c.Request().Context()
	var req requests.BoatCreateRequest
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
			return responses.NewErrorResponse(http.StatusBadRequest, "'id' field must not be provided when creating a boat").JSON(c)
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

	// Convert request to dme.BoatCreate with all fields properly mapped
	boat := &dme.BoatCreate{
		Name:                 req.Name,
		OwnerID:              req.OwnerID,
		Registration:         req.Registration,
		Year:                 req.Year,
		Make:                 req.Make,
		Model:                req.Model,
		HIN:                  req.Hin,
		LOA:                  req.LOA,
		LWL:                  req.LWL,
		Draft:                req.Draft,
		Beam:                 req.Beam,
		Height:               req.Height,
		Color:                req.Color,
		TrailerMake:          req.TrailerMake,
		TrailerModel:         req.TrailerModel,
		TrailerSerial:        req.TrailerSerial,
		TrailerRegistration:  req.TrailerRegistration,
		TrailerLocation:      req.TrailerLocation,
		SummerSlip:           req.SummerSlip,
		WinterSlip:           req.WinterSlip,
		InsuranceCompany:     req.InsuranceCompany,
		InsuranceExpDate:     req.InsuranceExpDate,
		SlipID:               req.SlipID,
		DoNotLaunch:          req.DoNotLaunch,
		BillingCodes:         req.BillingCodes,
		BoatDescriptionCodes: req.BoatDescriptionCodes,
		CustomInformation:    req.CustomInformation,
		OperationsHistory:    req.OperationsHistory,
		IntegrationID:        req.IntegrationID,
		OwnerIntegrationID:   req.OwnerIntegrationID,
		LastModified:         req.LastModified,
		Comments:             req.Comments,
	}

	dmeResponse, err := h.server.DME.CreateBoat(ctx, boat, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to create boat",
			zap.Error(err),
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertBoat(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary List boats new or changed
// @Description Retrieves boats created or changed after a specific date with pagination
// @Tags Boats
// @Accept json
// @Produce json
// @Param LastUpdate query string true "Date/Time to query from (URL encoded)"
// @Param Page query int true "Current page being requested" minimum(1)
// @Param PageSize query int true "Number of records per page" minimum(1) maximum(100)
// @Param ListName query string false "Optional: Name of list for paged data"
// @Success 200 {object} dme.BoatList
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/list-new-or-changed [get]
func (h *BoatHandler) ListBoatsNewOrChanged(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.BoatListNewOrChangedRequest)
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

	dmeResponse, err := h.server.DME.BoatsListNewOrChanged(ctx, req.LastUpdate, req.Page, req.PageSize, req.ListName, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list new or changed boats",
			zap.Error(err),
			zap.String("lastUpdate", req.LastUpdate),
			zap.Int("page", req.Page),
			zap.Int("pageSize", req.PageSize))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return c.JSON(http.StatusOK, dmeResponse)
}

// @Summary Retrieve boats with filters
// @Description Retrieves boats with optional filters (CustomerId, LastUpdateDate, HasInsurance)
// @Tags Boats
// @Accept json
// @Produce json
// @Param CustomerId query string false "Optional: Retrieves boats for a particular customer"
// @Param LastUpdateDate query string false "Optional: Retrieve boats modified on or after this date"
// @Param HasInsurance query boolean false "Optional: Filter by insurance status" default(false)
// @Success 200 {array} dme.Boat
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /boats/retrieve-boats [get]
func (h *BoatHandler) RetrieveBoats(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.BoatRetrieveListRequest)
	if err := c.Bind(req); err != nil {
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

	dmeResponse, err := h.server.DME.RetrieveBoatsFiltered(ctx, req.CustomerID, req.LastUpdateDate, req.HasInsurance, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve boats with filters",
			zap.Error(err),
			zap.String("customerId", req.CustomerID),
			zap.String("lastUpdateDate", req.LastUpdateDate),
			zap.Bool("hasInsurance", req.HasInsurance))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return c.JSON(http.StatusOK, dmeResponse)
}
