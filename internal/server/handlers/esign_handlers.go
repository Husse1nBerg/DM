package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// EsignHandler handles operations related to e-signature templates and documents
type EsignHandler struct {
	server       *s.Server
	esignService *s3.ESignService
}

// NewEsignHandler creates a new e-signature handler
func NewEsignHandler(server *s.Server) *EsignHandler {
	return &EsignHandler{server: server, esignService: server.ESignService}
}

// getUserInfoFromContext extracts user information from JWT token
func (h *EsignHandler) getUserInfoFromContext(c echo.Context) (userID uuid.UUID, organizationID uuid.UUID, marinaID uuid.UUID, err error) {
	userToken := c.Get("user").(*jwt.Token)
	if userToken == nil {
		err = responses.NewErrorResponse(http.StatusUnauthorized, "Authentication required").JSON(c)
		return
	}

	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID = claims.ID
	organizationID = claims.OrgId
	marinaID = claims.MarinaId
	return
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

	// Get total count for pagination
	total, err := h.server.DB.Queries().CountEsignTemplatesByMarina(c.Request().Context(), db.CountEsignTemplatesByMarinaParams{
		OrganizationID: organizationID,
		MarinaID:       marinaID,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error counting e-signature templates", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error counting templates").JSON(c)
	}
	totalInt := int64(total)

	// Get templates for the marina
	templates, err := h.server.DB.Queries().ListEsignTemplatesByMarina(c.Request().Context(), db.ListEsignTemplatesByMarinaParams{
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching e-signature templates", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching templates").JSON(c)
	}

	return responses.NewEsignTemplatesPaginatedResponse(templates, totalInt, req.PageSize, req.Page).JSON(c)
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
	file, header, err := c.Request().FormFile("file")
	if err == nil {
		// New file provided, upload it
		defer file.Close()

		err := h.esignService.UpdateFile(c.Request().Context(), file, header, blobUrl)
		if err != nil {
			h.server.Logger.Zap.Error("Error uploading file to S3", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
		}
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

	// Get total count for pagination
	total, err := h.server.DB.Queries().CountEsignDocumentsByMarina(c.Request().Context(), db.CountEsignDocumentsByMarinaParams{
		OrganizationID: organizationID,
		MarinaID:       marinaID,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error counting e-signature documents", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error counting documents").JSON(c)
	}
	totalInt := int64(total)

	// Get documents for the marina
	documents, err := h.server.DB.Queries().ListEsignDocumentsByMarina(c.Request().Context(), db.ListEsignDocumentsByMarinaParams{
		OrganizationID: organizationID,
		MarinaID:       marinaID,
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching e-signature documents", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching documents").JSON(c)
	}

	return responses.NewEsignDocumentsPaginatedResponse(documents, totalInt, req.PageSize, req.Page).JSON(c)
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

	// Parse optional template ID
	templateIDStr := c.FormValue("templateId")
	var document db.EsignDocument

	if templateIDStr != "" {
		// Parse template ID
		templateID, parseErr := uuid.Parse(templateIDStr)
		if parseErr != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid template ID format").JSON(c)
		}

		// Create document with template
		document, err = h.server.DB.Queries().CreateEsignDocumentWithTemplate(c.Request().Context(), db.CreateEsignDocumentWithTemplateParams{
			TemplateID:     templateID,
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Type:           documentType,
			Status:         status,
			BlobUrl:        filePath,
			BlobMetadata:   nil, // Ignoring blob metadata for now as requested
		})
	} else {
		// Create document without template
		document, err = h.server.DB.Queries().CreateEsignDocument(c.Request().Context(), db.CreateEsignDocumentParams{
			OrganizationID: organizationID,
			MarinaID:       marinaID,
			Type:           documentType,
			Status:         status,
			BlobUrl:        filePath,
			BlobMetadata:   nil, // Ignoring blob metadata for now as requested
		})
	}

	if err != nil {
		h.server.Logger.Zap.Error("Error creating e-signature document", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating document").JSON(c)
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
	file, header, err := c.Request().FormFile("file")
	if err == nil {
		// New file provided, upload it
		defer file.Close()

		err := h.esignService.UpdateFile(c.Request().Context(), file, header, blobUrl)
		if err != nil {
			h.server.Logger.Zap.Error("Error uploading file to S3", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
		}
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
