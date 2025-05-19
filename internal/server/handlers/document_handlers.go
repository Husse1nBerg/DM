package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DocumentHandler handles operations related to documents
type DocumentHandler struct {
	server *s.Server
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(server *s.Server) *DocumentHandler {
	return &DocumentHandler{server: server}
}

// UploadDocument creates a new document for a customer entity
//
//	@Summary		Upload document
//	@Description	Creates a new document for a customer
//	@Tags			Documents
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			marinaId	formData	string	true	"Marina ID"	Format(uuid)
//	@Param			entityId	formData	string	true	"Entity ID"
//	@Param			file		formData	file	true	"Document file"
//	@Success		201			{object}	responses.BaseResponse{data=responses.DocumentResponse}
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/documents/customer [post]
func (h *DocumentHandler) CustomerUploadDocument(c echo.Context) error {
	// Parse marina ID from form
	marinaIDStr := c.FormValue("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing marina ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID format").JSON(c)
	}

	// Check if marina exists
	_, err = h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(c)
	}

	entityType := "customer"
	entityID := c.FormValue("entityId")

	if entityID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Entity ID is required").JSON(c)
	}

	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		h.server.Logger.Zap.Error("Error getting file", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "File is required").JSON(c)
	}
	defer file.Close()

	// Upload the file to S3 using document storage service
	filePath, err := h.server.DocumentService.UploadFileToS3(c.Request().Context(), file, header, entityType)
	if err != nil {
		h.server.Logger.Zap.Error("Error uploading file to S3", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
	}

	// Log the file path and other details for debugging
	h.server.Logger.Zap.Info("File uploaded successfully",
		"filePath", filePath,
		"fileName", header.Filename,
		"fileType", header.Header.Get("Content-Type"),
		"fileSize", header.Size,
		"marinaID", marinaID,
		"entityType", entityType,
		"entityID", entityID,
	)

	// Create document record
	doc, err := h.server.DB.Queries().CreateDocument(c.Request().Context(), db.CreateDocumentParams{
		MarinaID:   marinaID,
		EntityType: entityType,
		EntityID:   entityID,
		FileName:   header.Filename,
		FileType:   header.Header.Get("Content-Type"),
		FilePath:   filePath,
		FileSize:   header.Size,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error creating document in database",
			"error", err,
			"marinaID", marinaID,
			"entityType", entityType,
			"entityID", entityID,
			"filePath", filePath,
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating document: "+err.Error()).JSON(c)
	}

	// Return created document
	response := responses.NewDocumentResponseSuccess(doc)
	response.Code = http.StatusCreated
	return response.JSON(c)
}

// GetDocumentsByEntity retrieves all documents for a customer entity
//
//	@Summary		Get customer documents
//	@Description	Retrieves all documents for a customer entity
//	@Tags			Documents
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	query		string	true	"Marina ID"	Format(uuid)
//	@Param			entityId	query		string	true	"Entity ID"
//	@Success		200			{array}		responses.BaseResponse{data=[]responses.DocumentResponse}
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/documents/customer [get]
func (h *DocumentHandler) CustomerGetDocumentsByEntity(c echo.Context) error {
	// Parse marina ID from query
	marinaIDStr := c.QueryParam("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing marina ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID format").JSON(c)
	}

	entityType := "customer"
	entityID := c.QueryParam("entityId")

	if entityID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Entity ID is required").JSON(c)
	}

	// Get documents from database
	documents, err := h.server.DB.Queries().ListDocumentsByEntity(c.Request().Context(), db.ListDocumentsByEntityParams{
		MarinaID:   marinaID,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching documents", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching documents").JSON(c)
	}

	// Return documents
	return responses.NewDocumentsResponseSuccess(documents).JSON(c)
}

// UploadDocument creates a new document for a boat entity
//
//	@Summary		Upload document
//	@Description	Creates a new document for a boat
//	@Tags			Documents
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			marinaId	formData	string	true	"Marina ID"	Format(uuid)
//	@Param			entityId	formData	string	true	"Entity ID"
//	@Param			file		formData	file	true	"Document file"
//	@Success		201			{object}	responses.BaseResponse{data=responses.DocumentResponse}
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/documents/boat [post]
func (h *DocumentHandler) BoatUploadDocument(c echo.Context) error {
	// Parse marina ID from form
	marinaIDStr := c.FormValue("marinaId")
	h.server.Logger.Zap.Info("Received marinaId from form",
		"raw_marinaId", marinaIDStr,
		"content_type", c.Request().Header.Get("Content-Type"))

	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing marina ID",
			"error", err,
			"raw_marinaId", marinaIDStr)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID format").JSON(c)
	}

	// Check if marina exists
	_, err = h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(c)
	}

	entityType := "boat"
	entityID := c.FormValue("entityId")

	if entityID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Entity ID is required").JSON(c)
	}

	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		h.server.Logger.Zap.Error("Error getting file", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "File is required").JSON(c)
	}
	defer file.Close()

	// Upload the file to S3 using document storage service
	filePath, err := h.server.DocumentService.UploadFileToS3(c.Request().Context(), file, header, entityType)
	if err != nil {
		h.server.Logger.Zap.Error("Error uploading file to S3", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
	}

	// Log the file path and other details for debugging
	h.server.Logger.Zap.Info("File uploaded successfully",
		"filePath", filePath,
		"fileName", header.Filename,
		"fileType", header.Header.Get("Content-Type"),
		"fileSize", header.Size,
		"marinaID", marinaID,
		"entityType", entityType,
		"entityID", entityID,
	)

	// Create document record
	doc, err := h.server.DB.Queries().CreateDocument(c.Request().Context(), db.CreateDocumentParams{
		MarinaID:   marinaID,
		EntityType: entityType,
		EntityID:   entityID,
		FileName:   header.Filename,
		FileType:   header.Header.Get("Content-Type"),
		FilePath:   filePath,
		FileSize:   header.Size,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error creating document in database",
			"error", err,
			"marinaID", marinaID,
			"entityType", entityType,
			"entityID", entityID,
			"filePath", filePath,
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating document: "+err.Error()).JSON(c)
	}

	// Return created document
	response := responses.NewDocumentResponseSuccess(doc)
	response.Code = http.StatusCreated
	return response.JSON(c)
}

// GetDocumentsByEntity retrieves all documents for a boat entity
//
//	@Summary		Get boat documents
//	@Description	Retrieves all documents for a boat entity
//	@Tags			Documents
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	query		string	true	"Marina ID"	Format(uuid)
//	@Param			entityId	query		string	true	"Entity ID"
//	@Success		200			{array}		responses.BaseResponse{data=[]responses.DocumentResponse}
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/documents/boat [get]
func (h *DocumentHandler) BoatGetDocumentsByEntity(c echo.Context) error {
	// Parse marina ID from query
	marinaIDStr := c.QueryParam("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing marina ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID format").JSON(c)
	}

	entityType := "boat"
	entityID := c.QueryParam("entityId")

	if entityID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Entity ID is required").JSON(c)
	}

	// Get documents from database
	documents, err := h.server.DB.Queries().ListDocumentsByEntity(c.Request().Context(), db.ListDocumentsByEntityParams{
		MarinaID:   marinaID,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching documents", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching documents").JSON(c)
	}

	// Return documents
	return responses.NewDocumentsResponseSuccess(documents).JSON(c)
}

// UploadDocument creates a new document for a user entity
//
//	@Summary		Upload document
//	@Description	Creates a new document for a user
//	@Tags			Documents
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			marinaId	formData	string	true	"Marina ID"	Format(uuid)
//	@Param			entityId	formData	string	true	"Entity ID"
//	@Param			file		formData	file	true	"Document file"
//	@Success		201			{object}	responses.BaseResponse{data=responses.DocumentResponse}
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/documents/user [post]
func (h *DocumentHandler) UserUploadDocument(c echo.Context) error {
	// Parse marina ID from form
	marinaIDStr := c.FormValue("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing marina ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID format").JSON(c)
	}

	// Check if marina exists
	_, err = h.server.DB.Queries().GetMarinaByID(c.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(c)
	}

	entityType := "user"
	entityID := c.FormValue("entityId")

	if entityID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Entity ID is required").JSON(c)
	}

	// Get file from form
	file, header, err := c.Request().FormFile("file")
	if err != nil {
		h.server.Logger.Zap.Error("Error getting file", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "File is required").JSON(c)
	}
	defer file.Close()

	// Upload the file to S3 using document storage service
	filePath, err := h.server.DocumentService.UploadFileToS3(c.Request().Context(), file, header, entityType)
	if err != nil {
		h.server.Logger.Zap.Error("Error uploading file to S3", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
	}

	// Log the file path and other details for debugging
	h.server.Logger.Zap.Info("File uploaded successfully",
		"filePath", filePath,
		"fileName", header.Filename,
		"fileType", header.Header.Get("Content-Type"),
		"fileSize", header.Size,
		"marinaID", marinaID,
		"entityType", entityType,
		"entityID", entityID,
	)

	// Create document record
	doc, err := h.server.DB.Queries().CreateDocument(c.Request().Context(), db.CreateDocumentParams{
		MarinaID:   marinaID,
		EntityType: entityType,
		EntityID:   entityID,
		FileName:   header.Filename,
		FileType:   header.Header.Get("Content-Type"),
		FilePath:   filePath,
		FileSize:   header.Size,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error creating document in database",
			"error", err,
			"marinaID", marinaID,
			"entityType", entityType,
			"entityID", entityID,
			"filePath", filePath,
		)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating document: "+err.Error()).JSON(c)
	}

	// Return created document
	response := responses.NewDocumentResponseSuccess(doc)
	response.Code = http.StatusCreated
	return response.JSON(c)
}

// GetDocumentsByEntity retrieves all documents for a user entity
//
//	@Summary		Get user documents
//	@Description	Retrieves all documents for a user entity
//	@Tags			Documents
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	query		string	true	"Marina ID"	Format(uuid)
//	@Param			entityId	query		string	true	"Entity ID"
//	@Success		200			{array}		responses.BaseResponse{data=[]responses.DocumentResponse}
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/documents/user [get]
func (h *DocumentHandler) UserGetDocumentsByEntity(c echo.Context) error {
	// Parse marina ID from query
	marinaIDStr := c.QueryParam("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing marina ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID format").JSON(c)
	}

	entityType := "user"
	entityID := c.QueryParam("entityId")

	if entityID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Entity ID is required").JSON(c)
	}

	// Get documents from database
	documents, err := h.server.DB.Queries().ListDocumentsByEntity(c.Request().Context(), db.ListDocumentsByEntityParams{
		MarinaID:   marinaID,
		EntityType: entityType,
		EntityID:   entityID,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching documents", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching documents").JSON(c)
	}

	// Return documents
	return responses.NewDocumentsResponseSuccess(documents).JSON(c)
}

// GetDocument retrieves a specific document
//
//	@Summary		Get document
//	@Description	Retrieves a specific document by ID
//	@Tags			Documents
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Document ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse{data=responses.DocumentResponse}
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/documents/{id} [get]
func (h *DocumentHandler) GetDocument(c echo.Context) error {
	// Parse document ID from path
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing document ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid document ID format").JSON(c)
	}

	// Get document from database
	document, err := h.server.DB.Queries().GetDocumentByID(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching document", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Document not found").JSON(c)
	}

	// Return document
	return responses.NewDocumentResponseSuccess(document).JSON(c)
}

// DeleteDocument deletes a document
//
//	@Summary		Delete document
//	@Description	Deletes a document
//	@Tags			Documents
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Document ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/documents/{id} [delete]
func (h *DocumentHandler) DeleteDocument(c echo.Context) error {
	// Parse document ID from path
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing document ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid document ID format").JSON(c)
	}

	// Delete document from database
	err = h.server.DB.Queries().DeleteDocument(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Error("Error deleting document", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error deleting document").JSON(c)
	}

	// Return success message
	return responses.NewMessageResponse(http.StatusOK, "Document successfully deleted").JSON(c)
}
