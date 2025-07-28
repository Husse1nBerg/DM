package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/notifications"
	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// EsignHandler handles operations related to e-signature templates and documents
type EsignHandler struct {
	server              *s.Server
	esignService        *s3.ESignService
	notificationService *notifications.NotificationService
}

// NewEsignHandler creates a new e-signature handler
func NewEsignHandler(server *s.Server) *EsignHandler {
	notificationService := notifications.NewNotificationService(
		server.DB.Queries(),
		server.Redis,
		server.Logger,
	)
	return &EsignHandler{server: server, esignService: server.ESignService, notificationService: notificationService}
}

// getUserInfoFromContext extracts user information from JWT token
func (h *EsignHandler) getUserInfoFromContext(c echo.Context) (userID uuid.UUID, organizationID uuid.UUID, marinaID uuid.UUID, err error) {
	userToken := c.Get("user").(*jwt.Token)
	if userToken == nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, responses.NewErrorResponse(http.StatusUnauthorized, "Authentication required").JSON(c)
	}

	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID = claims.ID
	organizationID = claims.OrgId
	// Fetch the user's current marina_id from the database
	queries := h.server.DB.Queries()
	user, dbErr := queries.GetUserByID(c.Request().Context(), claims.ID)
	if dbErr != nil {
		return uuid.Nil, uuid.Nil, uuid.Nil, responses.NewErrorResponse(http.StatusInternalServerError, "Failed to load user: "+dbErr.Error()).JSON(c)
	}
	marinaID = user.MarinaID
	return userID, organizationID, marinaID, nil
}

// updateEsignUsage updates the marina's e-signature usage count
func (h *EsignHandler) updateEsignUsage(ctx context.Context, marinaID uuid.UUID) error {
	increment := int16(1)

	_, err := h.server.DB.Queries().IncrementMarinaEmailUsage(ctx, db.IncrementMarinaEmailUsageParams{
		ID:      marinaID,
		Column2: increment,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error updating marina email usage", err)
		return err
	}
	return nil
}

// checkDocumentLimit checks if the marina has reached its document usage limit
func (h *EsignHandler) checkDocumentLimit(ctx context.Context, marinaID uuid.UUID) error {
	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina", err)
		return fmt.Errorf("marina not found")
	}
	documentPlan, err := h.server.DB.Queries().GetDocumentPlanByID(ctx, marina.DocumentPlanID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching document plan", err)
		return fmt.Errorf("document plan not found")
	}
	currentUsage := int64(0)
	if marina.DocumentUsage != nil {
		currentUsage = *marina.DocumentUsage
	}
	maxLimit := documentPlan.DocumentLimit
	if maxLimit == nil {
		return nil
	}
	if currentUsage >= int64(*maxLimit) {
		h.server.Logger.Zap.Info("Limit would be exceeded",
			"currentUsage", currentUsage,
			"maxLimit", *maxLimit)

		return fmt.Errorf("limit exceeded: current usage %d has reached the %s limit of %d",
			currentUsage, "documents", *maxLimit)
	}

	return nil
}

// ====================
// E-SIGNATURE TEMPLATES
// ====================

// ListEsignTemplates retrieves all e-signature templates for the user's marina
//
//	@Summary		List e-signature templates
//	@Description	Retrieves all e-signature templates for the authenticated user's marina
//	@Tags			E-signature Templates
//	@Accept			json
//	@Produce		json
//	@Param			status	query		string	false	"Template status (optional)" Enums(draft, active, archived)
//	@Param			page	query		int	false	"Page number"	default(1)	minimum(1)
//	@Param			pageSize	query		int	false	"Page size"	default(10)	minimum(1)	maximum(100)
//	@Success		200		{object}	responses.EsignTemplateListResponse
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		401		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/templates [get]
func (h *EsignHandler) ListEsignTemplates(c echo.Context) error {
	userID, organizationID, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina by user ID", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}
	marinaID := user.MarinaID

	status := c.QueryParam("status")

	var req requests.PaginationQuery
	if err := c.Bind(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	ctx := c.Request().Context()
	var templates []db.EsignTemplate
	var total int64
	if status != "" {
		templates, err = h.server.DB.Queries().ListEsignTemplatesByMarinaStatus(ctx, db.ListEsignTemplatesByMarinaStatusParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Column3:        status,
			Limit:          req.PageSize,
			Offset:         (req.Page - 1) * req.PageSize,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error fetching filtered templates", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching templates").JSON(c)
		}
		total, err = h.server.DB.Queries().CountEsignTemplatesByMarinaStatus(ctx, db.CountEsignTemplatesByMarinaStatusParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Column3:        status,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error counting filtered templates", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error counting templates").JSON(c)
		}
	} else {
		total64, err := h.server.DB.Queries().CountEsignTemplatesByMarina(ctx, db.CountEsignTemplatesByMarinaParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error counting templates", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error counting templates").JSON(c)
		}
		total = int64(total64)
		templates, err = h.server.DB.Queries().ListEsignTemplatesByMarina(ctx, db.ListEsignTemplatesByMarinaParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Limit:          req.PageSize,
			Offset:         (req.Page - 1) * req.PageSize,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error fetching templates", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching templates").JSON(c)
		}
	}
	return responses.NewEsignTemplatesPaginatedResponse(templates, total, req.PageSize, req.Page).JSON(c)
}

// CreateEsignTemplate creates a new e-signature template
//
//	@Summary		Create e-signature template
//	@Description	Creates a new e-signature template for the authenticated user's marina
//	@Tags			E-signature Templates
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			name			formData	string	true	"Template name"
//	@Param			description		formData	string	false	"Template description"
//	@Param			type			formData	string	true	"Template type"
//	@Param			status			formData	string	false	"Template status (default: draft)"
//	@Param			file			formData	file	true	"Template file"
//	@Success		201				{object}	responses.BaseResponse{data=responses.EsignTemplateResponse}
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		401				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/templates [post]
func (h *EsignHandler) CreateEsignTemplate(c echo.Context) error {
	userID, organizationID, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina by user ID", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}
	marinaID := user.MarinaID

	// Parse form values
	name := c.FormValue("name")
	if name == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Template name is required").JSON(c)
	}

	templateType := c.FormValue("type")
	if templateType == "" {
		templateType = "template"
	}

	status := c.FormValue("status")
	if status == "" {
		status = "draft" // Default status
	}

	description := c.FormValue("description")
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}

	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		h.server.Logger.Zap.Error("Error getting file", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Template file is required").JSON(c)
	}
	defer file.Close()

	// Upload the file to S3 using document storage service
	filePath, err := h.esignService.UploadFileToS3(c.Request().Context(), file, header, "esign_template")
	if err != nil {
		h.server.Logger.Zap.Error("Error uploading file to S3", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
	}

	// Parse optional jsonData field
	var jsonDataBytes []byte
	jsonDataStr := c.FormValue("jsonData")
	if jsonDataStr != "" {
		if err := json.Unmarshal([]byte(jsonDataStr), new(map[string]interface{})); err != nil {
			h.server.Logger.Zap.Error("Invalid jsonData", err)
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid jsonData: "+err.Error()).JSON(c)
		}
		jsonDataBytes = []byte(jsonDataStr)
	}

	// Create template
	template, err := h.server.DB.Queries().CreateEsignTemplate(c.Request().Context(), db.CreateEsignTemplateParams{
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Name:           name,
		Description:    descriptionPtr,
		Type:           templateType,
		Status:         status,
		BlobUrl:        filePath,
		BlobMetadata:   nil, // Ignoring blob metadata for now as requested
		JsonData:       jsonDataBytes,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error creating e-signature template", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating template").JSON(c)
	}

	response := responses.NewEsignTemplateResponseSuccess(template)
	response.Code = http.StatusCreated
	return response.JSON(c)
}

// GetEsignTemplate retrieves an e-signature template by ID
//
//	@Summary		Get e-signature template
//	@Description	Retrieves an e-signature template by ID
//	@Tags			E-signature Templates
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Template ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse{data=responses.EsignTemplateResponse}
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		401	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/templates/{id} [get]
func (h *EsignHandler) GetEsignTemplate(c echo.Context) error {
	_, _, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	// Parse template ID
	templateIDStr := c.Param("id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid template ID format").JSON(c)
	}

	// Get template
	template, err := h.server.DB.Queries().GetEsignTemplateByID(c.Request().Context(), templateID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching e-signature template", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Template not found").JSON(c)
	}

	return responses.NewEsignTemplateResponseSuccess(template).JSON(c)
}

// UpdateEsignTemplate updates an existing e-signature template
//
//	@Summary		Update e-signature template
//	@Description	Updates an existing e-signature template
//	@Tags			E-signature Templates
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id				path		string	true	"Template ID"	Format(uuid)
//	@Param			name			formData	string	false	"Template name"
//	@Param			description		formData	string	false	"Template description"
//	@Param			type			formData	string	false	"Template type"
//	@Param			status			formData	string	false	"Template status"
//	@Param			file			formData	file	false	"Template file (optional for update)"
//	@Success		200				{object}	responses.BaseResponse{data=responses.EsignTemplateResponse}
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		401				{object}	responses.BaseResponse
//	@Failure		404				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/templates/{id} [put]
func (h *EsignHandler) UpdateEsignTemplate(c echo.Context) error {
	_, _, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	// Parse template ID
	templateIDStr := c.Param("id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid template ID format").JSON(c)
	}

	// Get existing template
	existingTemplate, err := h.server.DB.Queries().GetEsignTemplateByID(c.Request().Context(), templateID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching existing template", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Template not found").JSON(c)
	}

	// Parse form values with fallbacks to existing values
	name := c.FormValue("name")
	if name == "" {
		name = existingTemplate.Name
	}

	templateType := c.FormValue("type")
	if templateType == "" {
		templateType = existingTemplate.Type
	}

	status := c.FormValue("status")
	if status == "" {
		status = existingTemplate.Status
	}

	description := c.FormValue("description")
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	} else {
		descriptionPtr = existingTemplate.Description
	}

	// Handle file upload (optional for update)
	blobUrl := existingTemplate.BlobUrl // Keep existing URL by default

	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err == nil {
		defer file.Close()
		// Upload the file to S3 using document storage service
		filePath, err := h.esignService.UploadFileToS3(c.Request().Context(), file, header, "esign_template")
		if err != nil {
			h.server.Logger.Zap.Error("Error uploading file to S3", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
		}
		blobUrl = filePath
	}

	// Parse optional jsonData field
	var jsonDataBytes []byte
	jsonDataStr := c.FormValue("jsonData")
	if jsonDataStr != "" {
		if err := json.Unmarshal([]byte(jsonDataStr), new(map[string]interface{})); err != nil {
			h.server.Logger.Zap.Error("Invalid jsonData", err)
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid jsonData: "+err.Error()).JSON(c)
		}
		jsonDataBytes = []byte(jsonDataStr)
	} else {
		jsonDataBytes = existingTemplate.JsonData
	}

	// Update template
	template, err := h.server.DB.Queries().UpdateEsignTemplate(c.Request().Context(), db.UpdateEsignTemplateParams{
		ID:           templateID,
		Name:         name,
		Description:  descriptionPtr,
		Type:         templateType,
		Status:       status,
		BlobUrl:      blobUrl,
		BlobMetadata: nil, // Ignoring blob metadata for now as requested
		JsonData:     jsonDataBytes,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error updating e-signature template", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating template").JSON(c)
	}

	return responses.NewEsignTemplateResponseSuccess(template).JSON(c)
}

// DeleteEsignTemplate soft deletes an e-signature template
//
//	@Summary		Delete e-signature template
//	@Description	Soft deletes an e-signature template
//	@Tags			E-signature Templates
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Template ID"	Format(uuid)
//	@Success		204	{object}	responses.BaseResponse
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		401	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/templates/{id} [delete]
func (h *EsignHandler) DeleteEsignTemplate(c echo.Context) error {
	_, _, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	// Parse template ID
	templateIDStr := c.Param("id")
	templateID, err := uuid.Parse(templateIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid template ID format").JSON(c)
	}

	// Soft delete template
	err = h.server.DB.Queries().SoftDeleteEsignTemplate(c.Request().Context(), templateID)
	if err != nil {
		h.server.Logger.Zap.Error("Error deleting e-signature template", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error deleting template").JSON(c)
	}

	return responses.NewSuccessResponse(nil).JSON(c)
}

// ====================
// E-SIGNATURE DOCUMENTS
// ====================

// ListEsignDocuments retrieves all e-signature documents for the user's marina
//
//	@Summary		List e-signature documents
//	@Description	Retrieves all e-signature documents for the authenticated user's marina
//	@Tags			E-signature Documents
//	@Accept			json
//	@Produce		json
//	@Param			status	query		string	false	"Document status (optional)" Enums(draft, signed, questions, sent)
//	@Param			page	query		int	false	"Page number"	default(1)	minimum(1)
//	@Param			pageSize	query		int	false	"Page size"	default(10)	minimum(1)	maximum(100)
//	@Success		200		{object}	responses.EsignDocumentListResponse
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		401		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/documents [get]
func (h *EsignHandler) ListEsignDocuments(c echo.Context) error {
	userID, organizationID, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}
	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina by user ID", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}
	marinaID := user.MarinaID

	status := c.QueryParam("status")

	var req requests.PaginationQuery
	if err := c.Bind(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	ctx := c.Request().Context()
	var documents []db.EsignDocument
	var total int64
	if status != "" {
		documents, err = h.server.DB.Queries().ListEsignDocumentsByMarinaStatus(ctx, db.ListEsignDocumentsByMarinaStatusParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Column3:        status,
			Limit:          req.PageSize,
			Offset:         (req.Page - 1) * req.PageSize,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error fetching filtered documents", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching documents").JSON(c)
		}
		total, err = h.server.DB.Queries().CountEsignDocumentsByMarinaStatus(ctx, db.CountEsignDocumentsByMarinaStatusParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Column3:        status,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error counting filtered documents", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error counting documents").JSON(c)
		}
	} else {
		total64, err := h.server.DB.Queries().CountEsignDocumentsByMarina(ctx, db.CountEsignDocumentsByMarinaParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error counting documents", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error counting documents").JSON(c)
		}
		total = int64(total64)
		documents, err = h.server.DB.Queries().ListEsignDocumentsByMarina(ctx, db.ListEsignDocumentsByMarinaParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Limit:          req.PageSize,
			Offset:         (req.Page - 1) * req.PageSize,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error fetching documents", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching documents").JSON(c)
		}
	}
	return responses.NewEsignDocumentsPaginatedResponse(documents, total, req.PageSize, req.Page).JSON(c)
}

// CreateEsignDocument creates a new e-signature document
//
//	@Summary		Create e-signature document
//	@Description	Creates a new e-signature document for the authenticated user's marina
//	@Tags			E-signature Documents
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			templateId		formData	string	false	"Template ID (optional)"	Format(uuid)
//	@Param			type			formData	string	true	"Document type"
//	@Param			status			formData	string	false	"Document status (default: draft)"
//	@Param			file			formData	file	true	"Document file"
//	@Success		201				{object}	responses.BaseResponse{data=responses.EsignDocumentResponse}
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		401				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/documents [post]
func (h *EsignHandler) CreateEsignDocument(c echo.Context) error {
	userID, organizationID, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina by user ID", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}
	marinaID := user.MarinaID

	// Parse form values
	documentType := c.FormValue("type")
	if documentType == "" {
		documentType = "document"
	}

	status := c.FormValue("status")
	if status == "" {
		status = "draft" // Default status
	}
	// Parse optional template ID
	templateIDStr := c.FormValue("templateId")
	var document db.EsignDocument

	if templateIDStr != "" {
		// Parse template ID
		templateID, parseErr := uuid.Parse(templateIDStr)
		if parseErr != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid template ID format").JSON(c)
		}

		template, err := h.server.DB.Queries().GetEsignTemplateByID(c.Request().Context(), templateID)
		if err != nil {
			h.server.Logger.Zap.Error("Error fetching template", err)
			return responses.NewErrorResponse(http.StatusNotFound, "Template not found").JSON(c)
		}

		// Duplicate the document file in S3
		duplicatedFilePath, err := h.esignService.DuplicateFile(c.Request().Context(), template.BlobUrl)
		if err != nil {
			h.server.Logger.Zap.Error("Error duplicating document file", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error duplicating document file: "+err.Error()).JSON(c)
		}

		// Create document with template
		document, err = h.server.DB.Queries().CreateEsignDocumentWithTemplate(c.Request().Context(), db.CreateEsignDocumentWithTemplateParams{
			TemplateID:     templateID,
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Type:           documentType,
			Status:         status,
			BlobUrl:        duplicatedFilePath,
			BlobMetadata:   nil, // Ignoring blob metadata for now as requested
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error creating e-signature document with template", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating document").JSON(c)
		}
	} else {

		// Get file from form
		file, header, err := c.Request().FormFile("file")
		if err != nil {
			h.server.Logger.Zap.Error("Error getting file", err)
			return responses.NewErrorResponse(http.StatusBadRequest, "Document file is required").JSON(c)
		}
		defer file.Close()

		// Upload the file to S3 using document storage service
		filePath, err := h.esignService.UploadFileToS3(c.Request().Context(), file, header, "esign_document")
		if err != nil {
			h.server.Logger.Zap.Error("Error uploading file to S3", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
		}

		// Create document without template
		document, err = h.server.DB.Queries().CreateEsignDocument(c.Request().Context(), db.CreateEsignDocumentParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Type:           documentType,
			Status:         status,
			BlobUrl:        filePath,
			BlobMetadata:   nil, // Ignoring blob metadata for now as requested
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error creating e-signature document", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating document").JSON(c)
		}
	}

	response := responses.NewEsignDocumentResponseSuccess(document)
	response.Code = http.StatusCreated
	return response.JSON(c)
}

// GetEsignDocument retrieves an e-signature document by ID
//
//	@Summary		Get e-signature document
//	@Description	Retrieves an e-signature document by ID
//	@Tags			E-signature Documents
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Document ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse{data=responses.EsignDocumentResponse}
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		401	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/documents/{id} [get]
func (h *EsignHandler) GetEsignDocument(c echo.Context) error {
	_, _, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	// Parse document ID
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid document ID format").JSON(c)
	}

	// Get document
	document, err := h.server.DB.Queries().GetEsignDocumentByID(c.Request().Context(), documentID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching e-signature document", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Document not found").JSON(c)
	}

	return responses.NewEsignDocumentResponseSuccess(document).JSON(c)
}

// UpdateEsignDocument updates an existing e-signature document
//
//	@Summary		Update e-signature document
//	@Description	Updates an existing e-signature document
//	@Tags			E-signature Documents
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id				path		string	true	"Document ID"	Format(uuid)
//	@Param			type			formData	string	false	"Document type"
//	@Param			status			formData	string	false	"Document status"
//	@Param			file			formData	file	false	"Document file (optional for update)"
//	@Success		200				{object}	responses.BaseResponse{data=responses.EsignDocumentResponse}
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		401				{object}	responses.BaseResponse
//	@Failure		404				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/documents/{id} [put]
func (h *EsignHandler) UpdateEsignDocument(c echo.Context) error {
	_, _, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	// Parse document ID
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid document ID format").JSON(c)
	}

	// Get existing document
	existingDocument, err := h.server.DB.Queries().GetEsignDocumentByID(c.Request().Context(), documentID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching existing document", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Document not found").JSON(c)
	}

	// Parse form values with fallbacks to existing values
	documentType := c.FormValue("type")
	if documentType == "" {
		documentType = existingDocument.Type
	}

	status := c.FormValue("status")
	if status == "" {
		status = existingDocument.Status
	}

	// Handle file upload (optional for update)
	blobUrl := existingDocument.BlobUrl // Keep existing URL by default

	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err == nil {
		defer file.Close()
		// Upload the file to S3 using document storage service
		filePath, err := h.esignService.UploadFileToS3(c.Request().Context(), file, header, "esign_document")
		if err != nil {
			h.server.Logger.Zap.Error("Error uploading file to S3", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
		}
		blobUrl = filePath
	}

	// Update document
	document, err := h.server.DB.Queries().UpdateEsignDocument(c.Request().Context(), db.UpdateEsignDocumentParams{
		ID:           documentID,
		Type:         documentType,
		Status:       status,
		BlobUrl:      blobUrl,
		BlobMetadata: nil, // Ignoring blob metadata for now as requested
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error updating e-signature document", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating document").JSON(c)
	}

	return responses.NewEsignDocumentResponseSuccess(document).JSON(c)
}

// DeleteEsignDocument soft deletes an e-signature document
//
//	@Summary		Delete e-signature document
//	@Description	Soft deletes an e-signature document
//	@Tags			E-signature Documents
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Document ID"	Format(uuid)
//	@Success		204	{object}	responses.BaseResponse
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		401	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/documents/{id} [delete]
func (h *EsignHandler) DeleteEsignDocument(c echo.Context) error {
	_, _, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	// Parse document ID
	documentIDStr := c.Param("id")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid document ID format").JSON(c)
	}

	// Soft delete document
	err = h.server.DB.Queries().SoftDeleteEsignDocument(c.Request().Context(), documentID)
	if err != nil {
		h.server.Logger.Zap.Error("Error deleting e-signature document", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error deleting document").JSON(c)
	}

	return responses.NewSuccessResponse(nil).JSON(c)
}

// ====================
// E-SIGNATURE SUBMISSIONS
// ====================

// ListEsignSubmissions retrieves all e-signature submissions for the user's marina, with optional filtering by customerId and status
//
//	@Summary		List e-signature submissions
//	@Description	Retrieves all e-signature submissions for the authenticated user's marina, with optional filtering by customerId and status
//	@Tags			E-signature Submissions
//	@Accept			json
//	@Produce		json
//	@Param			customerId	query		string	false	"Customer ID (optional)"
//	@Param			status		query		string	false	"Submission status (optional)" Enums(pending, signed, questions, sent)
//	@Param			page		query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			pageSize	query		int		false	"Page size"	default(10)	minimum(1)	maximum(100)
//	@Success		200		{object}	responses.EsignSubmissionListResponse
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		401		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/submissions [get]
func (h *EsignHandler) ListEsignSubmissions(c echo.Context) error {
	userID, organizationID, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina by user ID", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}
	marinaID := user.MarinaID

	// Parse query parameters
	customerID := c.QueryParam("customerId")
	status := c.QueryParam("status")

	// Parse pagination parameters using standard pattern
	var req requests.PaginationQuery
	if err := c.Bind(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	// Set defaults if not provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// Use new filtered queries if either filter is provided
	useFilter := customerID != "" || status != ""
	var (
		submissions []db.EsignSubmission
		total       int64
	)
	ctx := c.Request().Context()
	if useFilter {
		var customerIDVal, statusVal string
		if customerID != "" {
			customerIDVal = customerID
		}
		if status != "" {
			statusVal = status
		}
		submissions, err = h.server.DB.Queries().ListEsignSubmissionsByMarinaFiltered(ctx, db.ListEsignSubmissionsByMarinaFilteredParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Column3:        customerIDVal,
			Column4:        statusVal,
			Limit:          req.PageSize,
			Offset:         (req.Page - 1) * req.PageSize,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error fetching filtered e-signature submissions", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching submissions").JSON(c)
		}
		total, err = h.server.DB.Queries().CountEsignSubmissionsByMarinaFiltered(ctx, db.CountEsignSubmissionsByMarinaFilteredParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Column3:        customerIDVal,
			Column4:        statusVal,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error counting filtered e-signature submissions", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error counting submissions").JSON(c)
		}
	} else {
		submissions, err = h.server.DB.Queries().ListEsignSubmissionsByMarina(ctx, db.ListEsignSubmissionsByMarinaParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Limit:          req.PageSize,
			Offset:         (req.Page - 1) * req.PageSize,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error fetching e-signature submissions", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching submissions").JSON(c)
		}
		total64, err := h.server.DB.Queries().CountEsignSubmissionsByMarina(ctx, db.CountEsignSubmissionsByMarinaParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
		})
		if err != nil {
			h.server.Logger.Zap.Error("Error counting e-signature submissions", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error counting submissions").JSON(c)
		}
		total = int64(total64)
	}

	return responses.NewEsignSubmissionsPaginatedResponse(submissions, total, req.PageSize, req.Page).JSON(c)
}

// CreateEsignSubmission creates a new e-signature submission
//
//	@Summary		Create e-signature submission
//	@Description	Creates a new e-signature submission by duplicating the document file
//	@Tags			E-signature Submissions
//	@Accept			json
//	@Produce		json
//	@Param			request			body		requests.CreateEsignSubmissionRequest	true	"Create e-signature submission request"
//	@Success		201				{object}	responses.BaseResponse{data=responses.EsignSubmissionResponse}
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		401				{object}	responses.BaseResponse
//	@Failure		404				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/submissions [post]
func (h *EsignHandler) CreateEsignSubmission(c echo.Context) error {
	userID, organizationID, _, err := h.getUserInfoFromContext(c)
	logger := h.server.Logger
	if err != nil {
		return err
	}

	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina by user ID", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}
	marinaID := user.MarinaID

	// Check document usage limit
	// if err := h.checkDocumentLimit(c.Request().Context(), marinaID); err != nil {
	// 	return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	// }

	// Parse request body
	var req requests.CreateEsignSubmissionRequest
	if err := c.Bind(&req); err != nil {
		h.server.Logger.Zap.Error("Error parsing request body", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request body").JSON(c)
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		h.server.Logger.Zap.Error("Error validating request", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Validation failed: "+err.Error()).JSON(c)
	}

	// Get the document to duplicate
	document, err := h.server.DB.Queries().GetEsignDocumentByID(c.Request().Context(), req.DocumentID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching document by ID", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Document not found").JSON(c)
	}

	// Duplicate the document file in S3
	duplicatedFilePath, err := h.esignService.DuplicateFile(c.Request().Context(), document.BlobUrl)
	if err != nil {
		h.server.Logger.Zap.Error("Error duplicating document file", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error duplicating document file: "+err.Error()).JSON(c)
	}

	// Create submission with duplicated file
	submission, err := h.server.DB.Queries().CreateEsignSubmission(c.Request().Context(), db.CreateEsignSubmissionParams{
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		DocumentID:     req.DocumentID,
		Status:         "pending", // Default status
		BlobUrl:        duplicatedFilePath,
		BlobMetadata:   nil, // Ignoring blob metadata for now as requested
		CustomerID:     req.CustomerID,
		Email:          req.Email,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error creating e-signature submission", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating submission").JSON(c)
	}

	// Increment document usage immediately after successful submission
	_, err = h.server.DB.Queries().IncrementMarinaDocumentUsage(c.Request().Context(), db.IncrementMarinaDocumentUsageParams{
		ID:      marinaID,
		Column2: 1,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error incrementing document usage", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error incrementing document usage").JSON(c)
	}

	// Fetch marina again for email sender info
	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina by ID", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}
	// Create email data
	email := sendgrid.ESignSubmissionTemplateData{
		DocumentURL:     h.server.Config.App.EsignDocumentURL(submission.ID.String()),
		Recipient:       "",
		Sender:          marina.Name,
		ReplyTo:         marina.Email,
		TermsConditions: h.server.Config.App.TermsConditionsURL(),
	}
	to := []string{req.Email}
	subject := "New e-signature submission"

	// Send email asynchronously
	taskID, resultChan, err := h.server.SendGrid.SendESignSubmissionEmail(to, subject, email)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Log the task
	logger.Zap.Infow("Email queued", "task_id", taskID.String(), "to", req.Email)

	// Process the result asynchronously to log success/failure (no longer increments usage here)
	go func() {
		result := <-resultChan
		if result.Status == sendgrid.StatusSent {
			logger.Zap.Infow("Email sent successfully",
				"to", req.Email,
				"task_id", result.ID.String(),
				"message_id", result.ID,
				"status", result.Status)

		} else {
			logger.Zap.Errorw("Failed to send email",
				"to", req.Email,
				"task_id", result.ID.String(),
				"error", result.Error)
		}
	}()

	response := responses.NewEsignSubmissionResponseSuccess(submission)
	response.Code = http.StatusCreated
	return response.JSON(c)
}

// GetEsignSubmission retrieves an e-signature submission by ID
//
//	@Summary		Get e-signature submission
//	@Description	Retrieves an e-signature submission by ID
//	@Tags			E-signature Submissions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Submission ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse{data=responses.EsignSubmissionResponse}
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		401	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/submissions/{id} [get]
func (h *EsignHandler) GetEsignSubmission(c echo.Context) error {
	_, _, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	// Parse submission ID
	submissionIDStr := c.Param("id")
	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid submission ID format").JSON(c)
	}

	// Get submission
	submission, err := h.server.DB.Queries().GetEsignSubmissionByID(c.Request().Context(), submissionID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching e-signature submission", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Submission not found").JSON(c)
	}

	return responses.NewEsignSubmissionResponseSuccess(submission).JSON(c)
}

// UpdateEsignSubmission updates an existing e-signature submission
//
//	@Summary		Update e-signature submission
//	@Description	Updates an existing e-signature submission
//	@Tags			E-signature Submissions
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id				path		string	true	"Submission ID"	Format(uuid)
//	@Param			status			formData	string	false	"Submission status"
//	@Param			customerId		formData	string	false	"Customer ID"
//	@Param			email			formData	string	false	"Customer email"
//	@Param			file			formData	file	false	"Submission file (optional for update)"
//	@Success		200				{object}	responses.BaseResponse{data=responses.EsignSubmissionResponse}
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		401				{object}	responses.BaseResponse
//	@Failure		404				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/submissions/{id} [put]
func (h *EsignHandler) UpdateEsignSubmission(c echo.Context) error {
	_, _, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	// Parse submission ID
	submissionIDStr := c.Param("id")
	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid submission ID format").JSON(c)
	}

	// Get existing submission
	existingSubmission, err := h.server.DB.Queries().GetEsignSubmissionByID(c.Request().Context(), submissionID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching existing submission", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Submission not found").JSON(c)
	}

	// Parse form values with fallbacks to existing values
	status := c.FormValue("status")
	if status == "" {
		status = existingSubmission.Status
	}

	// email := c.FormValue("email")
	// if email == "" {
	// 	email = existingSubmission.Email
	// }

	customerID := c.FormValue("customerId")
	var customerIDPtr *string
	if customerID != "" {
		customerIDPtr = &customerID
	} else {
		customerIDPtr = existingSubmission.CustomerID
	}

	// Handle file upload (optional for update)
	blobUrl := existingSubmission.BlobUrl // Keep existing URL by default

	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err == nil {
		defer file.Close()
		// Upload the file to S3 using document storage service
		filePath, err := h.esignService.UploadFileToS3(c.Request().Context(), file, header, "esign_submission")
		if err != nil {
			h.server.Logger.Zap.Error("Error uploading file to S3", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
		}
		blobUrl = filePath
	}

	// Update submission
	submission, err := h.server.DB.Queries().UpdateEsignSubmission(c.Request().Context(), db.UpdateEsignSubmissionParams{
		ID:           submissionID,
		Status:       status,
		BlobUrl:      blobUrl,
		BlobMetadata: nil, // Ignoring blob metadata for now as requested
		CustomerID:   customerIDPtr,
		Email:        existingSubmission.Email,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error updating e-signature submission", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating submission").JSON(c)
	}

	return responses.NewEsignSubmissionResponseSuccess(submission).JSON(c)
}

// DeleteEsignSubmission soft deletes an e-signature submission
//
//	@Summary		Delete e-signature submission
//	@Description	Soft deletes an e-signature submission
//	@Tags			E-signature Submissions
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Submission ID"	Format(uuid)
//	@Success		204	{object}	responses.BaseResponse
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		401	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/submissions/{id} [delete]
func (h *EsignHandler) DeleteEsignSubmission(c echo.Context) error {
	_, _, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	// Parse submission ID
	submissionIDStr := c.Param("id")
	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid submission ID format").JSON(c)
	}

	// Soft delete submission
	err = h.server.DB.Queries().SoftDeleteEsignSubmission(c.Request().Context(), submissionID)
	if err != nil {
		h.server.Logger.Zap.Error("Error deleting e-signature submission", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error deleting submission").JSON(c)
	}

	return responses.NewSuccessResponse(nil).JSON(c)
}

// ListEsignSubmissionsByDocument retrieves submissions for a specific document
//
//	@Summary		List submissions by document
//	@Description	Retrieves all submissions for a specific document
//	@Tags			E-signature Submissions
//	@Accept			json
//	@Produce		json
//	@Param			documentId	path		string	true	"Document ID"	Format(uuid)
//	@Param			page		query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			pageSize	query		int		false	"Page size"	default(10)	minimum(1)	maximum(100)
//	@Success		200			{object}	responses.EsignSubmissionListResponse
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		401			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/documents/{documentId}/submissions [get]
func (h *EsignHandler) ListEsignSubmissionsByDocument(c echo.Context) error {
	_, _, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	// Parse document ID
	documentIDStr := c.Param("documentId")
	documentID, err := uuid.Parse(documentIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid document ID format").JSON(c)
	}

	// Parse pagination parameters
	var req requests.PaginationQuery
	if err := c.Bind(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	// Set defaults if not provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// Get total count for pagination
	total, err := h.server.DB.Queries().CountEsignSubmissionsByDocument(c.Request().Context(), documentID)
	if err != nil {
		h.server.Logger.Zap.Error("Error counting e-signature submissions", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error counting submissions").JSON(c)
	}
	totalInt := int64(total)

	// Get submissions for the document
	submissions, err := h.server.DB.Queries().ListEsignSubmissionsByDocument(c.Request().Context(), db.ListEsignSubmissionsByDocumentParams{
		DocumentID: documentID,
		Limit:      req.PageSize,
		Offset:     (req.Page - 1) * req.PageSize,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching e-signature submissions", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching submissions").JSON(c)
	}

	return responses.NewEsignSubmissionsPaginatedResponse(submissions, totalInt, req.PageSize, req.Page).JSON(c)
}

// ListEsignSubmissionsByStatus retrieves submissions filtered by status
//
//	@Summary		List submissions by status
//	@Description	Retrieves submissions filtered by status for the authenticated user's marina
//	@Tags			E-signature Submissions
//	@Accept			json
//	@Produce		json
//	@Param			status		query		string	true	"Submission status"	Enums(pending, signed, questions, sent)
//	@Param			page		query		int		false	"Page number"	default(1)	minimum(1)
//	@Param			pageSize	query		int		false	"Page size"	default(10)	minimum(1)	maximum(100)
//	@Success		200			{object}	responses.EsignSubmissionListResponse
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		401			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/esign/submissions/status [get]
func (h *EsignHandler) ListEsignSubmissionsByStatus(c echo.Context) error {
	userID, organizationID, _, err := h.getUserInfoFromContext(c)
	if err != nil {
		return err
	}

	user, err := h.server.DB.Queries().GetUserByID(c.Request().Context(), userID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina by user ID", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}
	marinaID := user.MarinaID

	// Parse status parameter
	status := c.QueryParam("status")
	if status == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Status parameter is required").JSON(c)
	}

	// Parse pagination parameters
	var req requests.PaginationQuery
	if err := c.Bind(&req); err != nil {
		req.Page = 1
		req.PageSize = 10
	}

	// Set defaults if not provided
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.PageSize <= 0 {
		req.PageSize = 10
	}

	// Get total count for pagination
	total, err := h.server.DB.Queries().CountEsignSubmissionsByStatus(c.Request().Context(), db.CountEsignSubmissionsByStatusParams{
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Status:         status,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error counting e-signature submissions", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error counting submissions").JSON(c)
	}
	totalInt := int64(total)

	// Get submissions for the marina with status filter
	submissions, err := h.server.DB.Queries().ListEsignSubmissionsByStatus(c.Request().Context(), db.ListEsignSubmissionsByStatusParams{
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Status:         status,
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching e-signature submissions", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching submissions").JSON(c)
	}

	return responses.NewEsignSubmissionsPaginatedResponse(submissions, totalInt, req.PageSize, req.Page).JSON(c)
}

// GetEsignSubmissionPublic retrieves an e-signature submission by ID (public endpoint)
//
//	@Summary		Get e-signature submission (public)
//	@Description	Retrieves an e-signature submission by ID without authentication
//	@Tags			E-signature Submissions
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string	true	"Submission ID"	Format(uuid)
//	@Success		200				{object}	responses.BaseResponse{data=responses.EsignSubmissionResponse}
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		404				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Router			/public/esign/submissions/{id} [get]
func (h *EsignHandler) GetEsignSubmissionPublic(c echo.Context) error {
	submissionIDStr := c.Param("id")
	if submissionIDStr == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Submission ID is required").JSON(c)
	}

	submissionID, err := uuid.Parse(submissionIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid submission ID format").JSON(c)
	}

	// Get the submission
	submission, err := h.server.DB.Queries().GetEsignSubmissionByID(c.Request().Context(), submissionID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching e-signature submission", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Submission not found").JSON(c)
	}

	response := responses.NewEsignSubmissionResponseSuccess(submission)
	return response.JSON(c)
}

// UpdateEsignSubmissionPublic updates an e-signature submission (public endpoint)
//
//	@Summary		Update e-signature submission (public)
//	@Description	Updates an e-signature submission status and file without authentication
//	@Tags			E-signature Submissions
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id				path		string	true	"Submission ID"	Format(uuid)
//	@Param			status			formData	string	true	"Submission status"
//	@Param			file			formData	file	false	"New submission file"
//	@Success		200				{object}	responses.BaseResponse{data=responses.EsignSubmissionResponse}
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		404				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Router			/public/esign/submissions/{id} [put]
func (h *EsignHandler) UpdateEsignSubmissionPublic(c echo.Context) error {
	// Parse submission ID
	submissionIDStr := c.Param("id")
	submissionID, err := uuid.Parse(submissionIDStr)
	logger := h.server.Logger
	queries := h.server.DB.Queries()
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid submission ID format").JSON(c)
	}

	// Get existing submission
	existingSubmission, err := h.server.DB.Queries().GetEsignSubmissionByID(c.Request().Context(), submissionID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching existing submission", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Submission not found").JSON(c)
	}

	// Parse form values with fallbacks to existing values
	status := c.FormValue("status")
	if status == "" {
		status = existingSubmission.Status
	}

	// Handle file upload (optional for update)
	blobUrl := existingSubmission.BlobUrl // Keep existing URL by default
	var uploadedFileName string

	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err == nil {
		defer file.Close()
		// Upload the file to S3 using document storage service
		filePath, err := h.esignService.UploadFileToS3(c.Request().Context(), file, header, "esign_submission")
		if err != nil {
			h.server.Logger.Zap.Error("Error uploading file to S3", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
		}
		blobUrl = filePath
		uploadedFileName = header.Filename
	}

	// Update submission
	submission, err := h.server.DB.Queries().UpdateEsignSubmission(c.Request().Context(), db.UpdateEsignSubmissionParams{
		ID:           submissionID,
		Status:       status,
		BlobUrl:      blobUrl,
		BlobMetadata: nil, // Ignoring blob metadata for now as requested
		CustomerID:   existingSubmission.CustomerID,
		Email:        existingSubmission.Email,
	})

	if err != nil {
		h.server.Logger.Zap.Error("Error updating e-signature submission", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating submission").JSON(c)
	}

	// If a file was uploaded AND the submission has a customer ID, update the DME customer with the attachment asynchronously
	if uploadedFileName != "" && submission.CustomerID != nil && *submission.CustomerID != "" {
		// Update DME customer asynchronously
		go func() {
			// Create a background context for the async operation
			ctx := context.Background()

			// Get marina information for DME API calls
			marina, err := h.server.DB.Queries().GetMarinaByID(ctx, submission.MarinaID)
			if err != nil {
				h.server.Logger.Zap.Error("[DME API] Error fetching marina for DME update", err)
				return
			}

			if marina.SystemID == nil {
				h.server.Logger.Zap.Warn("[DME API] Marina has no system ID, skipping DME customer update")
				return
			}

			orgID := marina.OrganizationID
			systemID := *marina.SystemID

			// Retrieve existing customer from DME
			existingCustomer, err := h.server.DME.CustomerRetrieve(ctx, *submission.CustomerID, orgID, systemID)
			if err != nil {
				h.server.Logger.Zap.Error("[DME API] Error retrieving customer from DME for attachment update", err)
				return
			}

			// Create new attachment
			datetime := time.Now().Format("2006-01-02 15:04:05")
			newAttachment := dme.Attachment{
				FileName:    uploadedFileName,
				Description: fmt.Sprintf("E-signature submission attachment %s", datetime),
				S3Path:      blobUrl,
				FileType:    utils.Pointer("application/pdf"),
				FromDMWeb:   utils.Pointer(true),
			}

			// Append to existing attachments
			updatedAttachments := existingCustomer.Attachments
			if updatedAttachments == nil {
				updatedAttachments = []dme.Attachment{}
			}
			updatedAttachments = append(updatedAttachments, newAttachment)

			// Create CustomerUpdate with all existing data plus new attachment
			customerUpdate := &dme.CustomerUpdate{
				ID:                        existingCustomer.ID,
				Name:                      existingCustomer.Name,
				FirstName:                 existingCustomer.FirstName,
				LastName:                  existingCustomer.LastName,
				Email:                     existingCustomer.Email,
				Address1:                  existingCustomer.Address1,
				Address2:                  existingCustomer.Address2,
				Address3:                  existingCustomer.Address3,
				City:                      existingCustomer.City,
				State:                     existingCustomer.State,
				Zip:                       existingCustomer.Zip,
				Country:                   existingCustomer.Country,
				Phone:                     existingCustomer.Phone,
				AltFirstName:              existingCustomer.AltFirstName,
				AltLastName:               existingCustomer.AltLastName,
				AltAddress1:               existingCustomer.AltAddress1,
				AltAddress2:               existingCustomer.AltAddress2,
				AltAddress3:               existingCustomer.AltAddress3,
				AltCity:                   existingCustomer.AltCity,
				AltState:                  existingCustomer.AltState,
				AltZip:                    existingCustomer.AltZip,
				AltCountry:                existingCustomer.AltCountry,
				AltPhone:                  existingCustomer.AltPhone,
				UseAltAddress:             existingCustomer.UseAltAddress,
				WorkPhone:                 existingCustomer.WorkPhone,
				CellPhone:                 existingCustomer.CellPhone,
				EmergencyContact:          existingCustomer.EmergencyContact,
				EmergencyPhone:            existingCustomer.EmergencyPhone,
				CompanyName:               existingCustomer.CompanyName,
				ShipmentMethod:            existingCustomer.ShipmentMethod,
				ShipmentMethodDescription: existingCustomer.ShipmentMethodDescription,
				CustomInformation:         existingCustomer.CustomInformation,
				Attachments:               updatedAttachments,
			}

			// Update customer in DME
			_, err = h.server.DME.CustomerUpdate(ctx, customerUpdate, orgID, systemID)
			if err != nil {
				h.server.Logger.Zap.Error("[DME API] Error updating customer attachments in DME", err)
			} else {
				h.server.Logger.Zap.Info("[DME API] Successfully updated DME customer with e-signature attachment",
					"customerID", *submission.CustomerID,
					"fileName", uploadedFileName,
					"submissionID", submission.ID.String())
			}
		}()
	}

	// Create notification for marina staff about new customer message
	// Find marina users to notify
	marinaUsers, err := queries.GetUsersByMarinaAdmin(c.Request().Context(), db.GetUsersByMarinaAdminParams{
		IsCustomer: utils.Pointer(false),
		MarinaID:   submission.MarinaID,
	})
	logger.Zap.Infow("Marina users", "marina_id", submission.MarinaID, "count", len(marinaUsers))
	if err != nil {
		logger.Zap.Warnw("Failed to get marina users for notification", "marina_id", submission.MarinaID, "error", err)
	} else {
		// Create notifications for marina staff
		for _, userRow := range marinaUsers {
			// Only notify active users
			if userRow.IsActive != nil && *userRow.IsActive {
				notificationErr := h.notificationService.CreateESignNotification(
					c.Request().Context(),
					userRow.ID,
					userRow.OrganizationID,
					userRow.MarinaID,
					submission.BlobUrl,
					submission.ID.String(),
					submission.Email,
				)
				if notificationErr != nil {
					logger.Zap.Warnw("Failed to create e-sign notification for marina user",
						"user_id", userRow.ID,
						"error", notificationErr)
				}
				logger.Zap.Infow("E-sign notification created for marina user",
					"user_id", userRow.ID,
					"user_name", userRow.FirstName+" "+userRow.LastName,
					"submission_id", submission.ID,
					"email", submission.Email)

			}
		}
	}

	return responses.NewEsignSubmissionResponseSuccess(submission).JSON(c)
}

// CreateEsignDocument creates a new e-signature document
//
//	@Summary		Create e-signature document for DME
//	@Description	Creates a new e-signature document for the DME system
//	@Tags			E-signature Documents
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			systemId		formData	string	true	"System ID"
//	@Param			status			formData	string	false	"Document status (default: dme_draft)"
//	@Param			type			formData	string	false	"Document type (default: document)"
//	@Param			file			formData	file	true	"Document file"
//	@Success		201				{object}	responses.BaseResponse{data=responses.EsignDocumentResponse}
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Param			X-API-Key		header		string	true	"DME API Key"
//	@Router			/external/dme/esign/documents [post]
func (h *EsignHandler) CreateEsignDocumentDME(c echo.Context) error {
	systemID := c.FormValue("systemId")
	if systemID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required").JSON(c)
	}

	sysid, err := h.server.DB.Queries().GetDMESysIdBySystemID(c.Request().Context(), systemID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching DME sysid by system ID - Not found", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "System Id is not linked to any marina").JSON(c)
	}

	if sysid.MarinaID == uuid.Nil {
		h.server.Logger.Zap.Error("No marina linked to DME sysid", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "System Id is not linked to any marina").JSON(c)
	}

	organizationID := sysid.OrganizationID
	marinaID := sysid.MarinaID

	// Parse form values
	documentType := c.FormValue("type")
	if documentType == "" {
		documentType = "document"
	}

	status := c.FormValue("status")
	if status == "" {
		status = "dme_draft" // Default status
	}
	customerID := c.FormValue("customerId")
	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		h.server.Logger.Zap.Error("Error getting file", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Document file is required").JSON(c)
	}
	defer file.Close()

	// Upload the file to S3 using document storage service
	filePath, err := h.esignService.UploadFileToS3(c.Request().Context(), file, header, "esign_document")
	if err != nil {
		h.server.Logger.Zap.Error("Error uploading file to S3", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
	}

	// Create document without template
	document, err := h.server.DB.Queries().CreateEsignDocument(c.Request().Context(), db.CreateEsignDocumentParams{
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Type:           documentType,
		Status:         status,
		BlobUrl:        filePath,
		BlobMetadata:   fmt.Appendf(nil, `{"customerId": "%s"}`, customerID),
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error creating e-signature document", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating document").JSON(c)
	}

	response := responses.NewEsignDocumentResponseSuccess(document)
	response.Code = http.StatusCreated
	return response.JSON(c)
}
