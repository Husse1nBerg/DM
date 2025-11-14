package handlers

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/notifications"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// DocumentHandler handles operations related to documents
type DocumentHandler struct {
	server              *s.Server
	notificationService *notifications.NotificationService
}

// NewDocumentHandler creates a new document handler
func NewDocumentHandler(server *s.Server) *DocumentHandler {
	// Initialize notification service
	notificationService := notifications.NewNotificationService(
		server.DB.Queries(),
		server.Redis,
		server.Logger,
		server.SendGrid,
		server.Config,
	)

	return &DocumentHandler{
		server:              server,
		notificationService: notificationService,
	}
}

// checkStorageLimit checks if the marina has enough storage space for the new file
func (h *DocumentHandler) checkStorageLimit(ctx echo.Context, marinaID uuid.UUID, fileSize int64) error {
	// Get marina to check current storage usage
	marina, err := h.server.DB.Queries().GetMarinaByID(ctx.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(ctx)
	}

	// Get marina's storage plan
	storagePlan, err := h.server.DB.Queries().GetMarinaStoragePlan(ctx.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching storage plan", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching storage plan").JSON(ctx)
	}

	// If storage limit is nil, it means unlimited storage
	if storagePlan.StorageLimitGb == nil {
		return nil
	}

	const bytesInGB = 1024 * 1024 * 1024 // 1 GB in bytes
	maxStorageLimit := int64(*storagePlan.StorageLimitGb) * bytesInGB

	// Check if adding the new file would exceed the limit
	currentUsage := int64(0)
	if marina.StorageUsage != nil {
		currentUsage = *marina.StorageUsage
	}

	if currentUsage+fileSize > maxStorageLimit {
		h.server.Logger.Zap.Info("Storage limit would be exceeded",
			"currentUsage", currentUsage,
			"fileSize", fileSize,
			"maxLimit", maxStorageLimit)

		// Convert values to GB for the error message
		currentUsageGB := float64(currentUsage) / float64(bytesInGB)
		maxLimitGB := float64(maxStorageLimit) / float64(bytesInGB)
		fileSizeKB := float64(fileSize) / 1024.0 // Convert to KB

		return fmt.Errorf("storage limit exceeded: current usage %.2f GB + file size %.2f KB would exceed limit of %.2f GB",
			currentUsageGB, fileSizeKB, maxLimitGB)
	}

	return nil
}

// updateStorageUsage updates the marina's storage usage
func (h *DocumentHandler) updateStorageUsage(ctx echo.Context, marinaID uuid.UUID, fileSize int64, isIncrement bool) error {
	var err error
	if isIncrement {
		_, err = h.server.DB.Queries().IncrementMarinaStorageUsage(ctx.Request().Context(), db.IncrementMarinaStorageUsageParams{
			ID:           marinaID,
			StorageUsage: &fileSize,
		})
	} else {
		_, err = h.server.DB.Queries().DecrementMarinaStorageUsage(ctx.Request().Context(), db.DecrementMarinaStorageUsageParams{
			ID:           marinaID,
			StorageUsage: &fileSize,
		})
	}

	if err != nil {
		h.server.Logger.Zap.Error("Error updating marina storage usage", err)
		return err
	}
	return nil
}

// CustomerUploadDocument creates a new document for a customer entity
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

	// Check storage limit before upload
	// if err := h.checkStorageLimit(c, marinaID, header.Size); err != nil {
	// 	h.server.Logger.Zap.Error("Storage limit check failed", err)
	// 	return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	// }

	// Upload the file to S3 using document storage service
	filePath, err := h.server.DocumentService.UploadFileToS3(c.Request().Context(), file, header, entityType)
	if err != nil {
		h.server.Logger.Zap.Error("Error uploading file to S3", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
	}

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
		h.server.Logger.Zap.Error("Error creating document in database", err)
		// TODO: Delete file from S3 since document creation failed
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating document: "+err.Error()).JSON(c)
	}

	// Update marina storage usage
	if err := h.updateStorageUsage(c, marinaID, header.Size, true); err != nil {
		h.server.Logger.Zap.Error("Error updating storage usage", err)
		// TODO: Delete file from S3 and document record since storage update failed
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating storage usage").JSON(c)
	}

	// Return created document
	response := responses.NewDocumentResponseSuccess(doc)
	response.Code = http.StatusCreated

	// Get current user to check if they are a customer
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID

	// Get marina information for both notification and DME operations
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("[DME API] Error fetching marina for DME update (document)", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}

	// Only send notifications if the current user is a customer
	if claims.IsCustomer == nil || *claims.IsCustomer {
		// Create notification for marina staff about new customer document
		marinaUsers, err := h.server.DB.Queries().GetUsersByMarina(ctx, db.GetUsersByMarinaParams{
			MarinaID:   marinaID,
			IsCustomer: utils.Pointer(false), // Get marina staff, not customers
		})
		if err != nil {
			h.server.Logger.Zap.Warnw("Failed to get marina users for notification", "marina_id", marinaID, "error", err)
		} else {
			// Create notifications for marina staff using smart notification system
			// Use background context with timeout for notification operations
			// Timeout calculation: 37 users × 5.5 seconds = ~3.4 minutes, so we use 5 minutes for safety
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()

				emailData := &notifications.EmailNotificationData{
					To:      []string{}, // No specific email recipients for document notifications
					Subject: "New Customer Document Uploaded",
				}

				results, err := h.notificationService.CreateBulkDocumentNotifications(
					ctx,
					marinaUsers,
					marina.OrganizationID,
					marinaID,
					header.Filename, // Document filename
					entityID,        // Customer ID
					emailData,
				)
				if err != nil {
					h.server.Logger.Zap.Warnw("Failed to create bulk document notifications", "error", err)
				} else {
					// Log notification results
					for _, result := range results {
						if len(result.Errors) > 0 {
							h.server.Logger.Zap.Warnw("Document notification delivery had errors",
								"user_id", result.UserID,
								"errors", result.Errors)
						} else {
							h.server.Logger.Zap.Infow("Document notification delivered successfully",
								"user_id", result.UserID,
								"system", result.SystemDelivered,
								"email", result.EmailDelivered)
						}
					}
				}
			}()
		}
	} else {
		h.server.Logger.Zap.Infow("Skipping notifications - document uploaded by customer user",
			"user_id", userID,
			"is_customer", *claims.IsCustomer,
			"document_id", doc.ID)
	}

	// --- DME Attachment Insert (Async, e-sign/gallery style) ---
	go func() {
		ctx := context.Background()

		if marina.SystemID == nil {
			h.server.Logger.Zap.Warn("[DME API] Marina has no system ID, skipping DME customer update (document)")
			return
		}
		orgID := marina.OrganizationID
		systemID := *marina.SystemID

		dmeCustomer, err := h.server.DME.CustomerRetrieve(ctx, entityID, orgID, systemID)
		if err != nil {
			h.server.Logger.Zap.Error("[DME API] Error retrieving customer from DME for attachment update (document)", err)
			return
		}

		fileType := header.Header.Get("Content-Type")
		datetime := time.Now().Format("2006-01-02 15:04:05")
		newAttachment := dme.Attachment{
			FileName:    header.Filename,
			Description: fmt.Sprintf("Document attachment %s", datetime),
			S3Path:      filePath,
			FileType:    utils.Pointer(fileType),
			FromDMWeb:   utils.Pointer(true),
		}

		updatedAttachments := dmeCustomer.Attachments
		if updatedAttachments == nil {
			updatedAttachments = []dme.Attachment{}
		}
		updatedAttachments = append(updatedAttachments, newAttachment)

		customerUpdate := &dme.CustomerUpdate{
			ID:                        dmeCustomer.ID,
			Name:                      dmeCustomer.Name,
			FirstName:                 dmeCustomer.FirstName,
			LastName:                  dmeCustomer.LastName,
			Email:                     dmeCustomer.Email,
			Address1:                  dmeCustomer.Address1,
			Address2:                  dmeCustomer.Address2,
			Address3:                  dmeCustomer.Address3,
			City:                      dmeCustomer.City,
			State:                     dmeCustomer.State,
			Zip:                       dmeCustomer.Zip,
			Country:                   dmeCustomer.Country,
			Phone:                     dmeCustomer.Phone,
			AltFirstName:              dmeCustomer.AltFirstName,
			AltLastName:               dmeCustomer.AltLastName,
			AltAddress1:               dmeCustomer.AltAddress1,
			AltAddress2:               dmeCustomer.AltAddress2,
			AltAddress3:               dmeCustomer.AltAddress3,
			AltCity:                   dmeCustomer.AltCity,
			AltState:                  dmeCustomer.AltState,
			AltZip:                    dmeCustomer.AltZip,
			AltCountry:                dmeCustomer.AltCountry,
			AltPhone:                  dmeCustomer.AltPhone,
			UseAltAddress:             dmeCustomer.UseAltAddress,
			WorkPhone:                 dmeCustomer.WorkPhone,
			CellPhone:                 dmeCustomer.CellPhone,
			EmergencyContact:          dmeCustomer.EmergencyContact,
			EmergencyPhone:            dmeCustomer.EmergencyPhone,
			CompanyName:               dmeCustomer.CompanyName,
			ShipmentMethod:            dmeCustomer.ShipmentMethod,
			ShipmentMethodDescription: dmeCustomer.ShipmentMethodDescription,
			CustomInformation:         dmeCustomer.CustomInformation,
			Attachments:               updatedAttachments,
		}

		_, err = h.server.DME.CustomerUpdate(ctx, customerUpdate, orgID, systemID)
		if err != nil {
			h.server.Logger.Zap.Error("[DME API] Error updating customer attachments in DME (document)", err)
		} else {
			h.server.Logger.Zap.Info("[DME API] Successfully updated DME customer with document attachment",
				"customerID", entityID,
				"fileName", header.Filename,
				"documentID", doc.ID.String())
		}
	}()

	return response.JSON(c)
}

// CustomerUploadDocument uploads a document for a customer entity (public endpoint)
//
//	@Summary		Upload document (public)
//	@Description	Creates a new document for a customer (public, no authentication required)
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
//	@Router			/public/documents/customer [post]
func (h *DocumentHandler) CustomerUploadDocumentPublic(c echo.Context) error {
	// Parse marina ID from form
	marinaIDStr := c.FormValue("marinaId")
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing marina ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID format").JSON(c)
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

	// Check storage limit before upload
	// if err := h.checkStorageLimit(c, marinaID, header.Size); err != nil {
	// 	h.server.Logger.Zap.Error("Storage limit check failed", err)
	// 	return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	// }

	// Upload the file to S3 using document storage service
	filePath, err := h.server.DocumentService.UploadFileToS3(c.Request().Context(), file, header, entityType)
	if err != nil {
		h.server.Logger.Zap.Error("Error uploading file to S3", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
	}

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
		h.server.Logger.Zap.Error("Error creating document in database", err)
		// TODO: Delete file from S3 since document creation failed
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating document: "+err.Error()).JSON(c)
	}

	// Update marina storage usage
	if err := h.updateStorageUsage(c, marinaID, header.Size, true); err != nil {
		h.server.Logger.Zap.Error("Error updating storage usage", err)
		// TODO: Delete file from S3 and document record since storage update failed
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating storage usage").JSON(c)
	}

	// Return created document
	response := responses.NewDocumentResponseSuccess(doc)
	response.Code = http.StatusCreated

	// --- DME Attachment Insert (Async, e-sign/gallery style) ---
	go func() {
		ctx := context.Background()

		marina, err := h.server.DB.Queries().GetMarinaByID(ctx, marinaID)
		if err != nil {
			h.server.Logger.Zap.Error("[DME API] Error fetching marina for DME update (customer document)", err)
			return
		}
		if marina.SystemID == nil {
			h.server.Logger.Zap.Warn("[DME API] Marina has no system ID, skipping DME customer update (document)")
			return
		}
		orgID := marina.OrganizationID
		systemID := *marina.SystemID

		dmeCustomer, err := h.server.DME.CustomerRetrieve(ctx, entityID, orgID, systemID)
		if err != nil {
			h.server.Logger.Zap.Error("[DME API] Error retrieving customer from DME for attachment update (document)", err)
			return
		}

		fileType := header.Header.Get("Content-Type")
		datetime := time.Now().Format("2006-01-02 15:04:05")
		newAttachment := dme.Attachment{
			FileName:    header.Filename,
			Description: fmt.Sprintf("Document attachment %s", datetime),
			S3Path:      filePath,
			FileType:    utils.Pointer(fileType),
			FromDMWeb:   utils.Pointer(true),
		}

		updatedAttachments := dmeCustomer.Attachments
		if updatedAttachments == nil {
			updatedAttachments = []dme.Attachment{}
		}
		updatedAttachments = append(updatedAttachments, newAttachment)

		customerUpdate := &dme.CustomerUpdate{
			ID:                        dmeCustomer.ID,
			Name:                      dmeCustomer.Name,
			FirstName:                 dmeCustomer.FirstName,
			LastName:                  dmeCustomer.LastName,
			Email:                     dmeCustomer.Email,
			Address1:                  dmeCustomer.Address1,
			Address2:                  dmeCustomer.Address2,
			Address3:                  dmeCustomer.Address3,
			City:                      dmeCustomer.City,
			State:                     dmeCustomer.State,
			Zip:                       dmeCustomer.Zip,
			Country:                   dmeCustomer.Country,
			Phone:                     dmeCustomer.Phone,
			AltFirstName:              dmeCustomer.AltFirstName,
			AltLastName:               dmeCustomer.AltLastName,
			AltAddress1:               dmeCustomer.AltAddress1,
			AltAddress2:               dmeCustomer.AltAddress2,
			AltAddress3:               dmeCustomer.AltAddress3,
			AltCity:                   dmeCustomer.AltCity,
			AltState:                  dmeCustomer.AltState,
			AltZip:                    dmeCustomer.AltZip,
			AltCountry:                dmeCustomer.AltCountry,
			AltPhone:                  dmeCustomer.AltPhone,
			UseAltAddress:             dmeCustomer.UseAltAddress,
			WorkPhone:                 dmeCustomer.WorkPhone,
			CellPhone:                 dmeCustomer.CellPhone,
			EmergencyContact:          dmeCustomer.EmergencyContact,
			EmergencyPhone:            dmeCustomer.EmergencyPhone,
			CompanyName:               dmeCustomer.CompanyName,
			ShipmentMethod:            dmeCustomer.ShipmentMethod,
			ShipmentMethodDescription: dmeCustomer.ShipmentMethodDescription,
			CustomInformation:         dmeCustomer.CustomInformation,
			Attachments:               updatedAttachments,
		}

		_, err = h.server.DME.CustomerUpdate(ctx, customerUpdate, orgID, systemID)
		if err != nil {
			h.server.Logger.Zap.Error("[DME API] Error updating customer attachments in DME (document)", err)
		} else {
			h.server.Logger.Zap.Info("[DME API] Successfully updated DME customer with document attachment",
				"customerID", entityID,
				"fileName", header.Filename,
				"documentID", doc.ID.String())
		}
	}()

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
//	@Router			/public/documents/customer [get]
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

// CustomerGetDocumentsByEntityPublic retrieves all documents for a customer entity (public endpoint)
//
//	@Summary		Get customer documents (public)
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
//	@Router			/public/documents/customer [get]
func (h *DocumentHandler) CustomerGetDocumentsByEntityPublic(c echo.Context) error {
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
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	documents, err := h.server.DB.Queries().ListDocumentsByEntity(ctx, db.ListDocumentsByEntityParams{
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
	marinaID, err := uuid.Parse(marinaIDStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing marina ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID format").JSON(c)
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

	// Check storage limit before upload
	// if err := h.checkStorageLimit(c, marinaID, header.Size); err != nil {
	// 	h.server.Logger.Zap.Error("Storage limit check failed", err)
	// 	return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	// }

	// Upload the file to S3 using document storage service
	filePath, err := h.server.DocumentService.UploadFileToS3(c.Request().Context(), file, header, entityType)
	if err != nil {
		h.server.Logger.Zap.Error("Error uploading file to S3", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	isInternalUser := !*claims.IsCustomer || false

	var isPublic bool
	if claims.IsCustomer != nil {
		isPublic = *claims.IsCustomer
	} else {
		isPublic = false // nil IsCustomer is treated as internal user
	}

	// Create document record
	doc, err := h.server.DB.Queries().CreateDocument(c.Request().Context(), db.CreateDocumentParams{
		MarinaID:   marinaID,
		EntityType: entityType,
		EntityID:   entityID,
		FileName:   header.Filename,
		FileType:   header.Header.Get("Content-Type"),
		FilePath:   filePath,
		FileSize:   header.Size,
		Public:     isPublic,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error creating document in database", err)
		// TODO: Delete file from S3 since document creation failed
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating document: "+err.Error()).JSON(c)
	}

	// Update marina storage usage
	if err := h.updateStorageUsage(c, marinaID, header.Size, true); err != nil {
		h.server.Logger.Zap.Error("Error updating storage usage", err)
		// TODO: Delete file from S3 and document record since storage update failed
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating storage usage").JSON(c)
	}

	// Return created document
	response := responses.NewDocumentResponseSuccess(doc)
	response.Code = http.StatusCreated

	// Get marina information for both notification and DME operations
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("[DME API] Error fetching marina for DME update (document)", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}

	// Only send notifications if the current user is a customer
	if isInternalUser {
		// Create notification for marina staff about new customer document
		marinaUsers, err := h.server.DB.Queries().GetUsersByMarina(ctx, db.GetUsersByMarinaParams{
			MarinaID:   marinaID,
			IsCustomer: utils.Pointer(false), // Get marina staff, not customers
		})
		if err != nil {
			h.server.Logger.Zap.Warnw("Failed to get marina users for notification", "marina_id", marinaID, "error", err)
		} else {
			// Create notifications for marina staff using smart notification system
			// Use background context with timeout for notification operations
			// Timeout calculation: 37 users × 5.5 seconds = ~3.4 minutes, so we use 5 minutes for safety
			go func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				defer cancel()

				emailData := &notifications.EmailNotificationData{
					To:      []string{}, // No specific email recipients for document notifications
					Subject: "New Customer Document Uploaded",
				}

				results, err := h.notificationService.CreateBulkDocumentNotifications(
					ctx,
					marinaUsers,
					marina.OrganizationID,
					marinaID,
					header.Filename, // Document filename
					entityID,        // Customer ID
					emailData,
				)
				if err != nil {
					h.server.Logger.Zap.Warnw("Failed to create bulk document notifications", "error", err)
				} else {
					// Log notification results
					for _, result := range results {
						if len(result.Errors) > 0 {
							h.server.Logger.Zap.Warnw("Document notification delivery had errors",
								"user_id", result.UserID,
								"errors", result.Errors)
						} else {
							h.server.Logger.Zap.Infow("Document notification delivered successfully",
								"user_id", result.UserID,
								"system", result.SystemDelivered,
								"email", result.EmailDelivered)
						}
					}
				}
			}()
		}
	} else {
		h.server.Logger.Zap.Infow("Skipping notifications - document uploaded by customer user",
			"user_id", claims.ID,
			"is_customer", *claims.IsCustomer,
			"document_id", doc.ID)
	}

	go func() {
		ctx := context.Background()

		if marina.SystemID == nil {
			h.server.Logger.Zap.Warn("[DME API] Marina has no system ID, skipping DME boat update (document)")
			return
		}
		orgID := marina.OrganizationID
		systemID := *marina.SystemID

		dmeBoat, err := h.server.DME.RetrieveBoatByID(ctx, entityID, orgID, systemID)
		if err != nil {
			h.server.Logger.Zap.Error("[DME API] Error retrieving boat from DME for attachment update (document)", err)
			return
		}

		fileType := header.Header.Get("Content-Type")
		datetime := time.Now().Format("2006-01-02 15:04:05")
		newAttachment := dme.Attachment{
			FileName:    header.Filename,
			Description: fmt.Sprintf("Document attachment %s", datetime),
			S3Path:      filePath,
			FileType:    utils.Pointer(fileType),
			FromDMWeb:   utils.Pointer(true),
		}

		updatedAttachments := dmeBoat.Attachments
		if updatedAttachments == nil {
			updatedAttachments = []dme.Attachment{}
		}
		updatedAttachments = append(updatedAttachments, newAttachment)

		boatUpdate := &dme.BoatUpdate{
			ID:                   dmeBoat.ID,
			Name:                 dmeBoat.Name,
			Registration:         dmeBoat.Registration,
			Year:                 dmeBoat.Year,
			Make:                 dmeBoat.Make,
			Model:                dmeBoat.Model,
			HIN:                  dmeBoat.HIN,
			LOA:                  dmeBoat.LOA,
			LWL:                  dmeBoat.LWL,
			Draft:                dmeBoat.Draft,
			Beam:                 dmeBoat.Beam,
			Height:               dmeBoat.Height,
			Color:                dmeBoat.Color,
			TrailerMake:          dmeBoat.TrailerMake,
			TrailerModel:         dmeBoat.TrailerModel,
			TrailerSerial:        dmeBoat.TrailerSerial,
			TrailerRegistration:  dmeBoat.TrailerRegistration,
			TrailerLocation:      dmeBoat.TrailerLocation,
			SummerSlip:           dmeBoat.SummerSlip,
			WinterSlip:           dmeBoat.WinterSlip,
			InsuranceCompany:     dmeBoat.InsuranceCompany,
			InsuranceExpDate:     dmeBoat.InsuranceExpDate,
			SlipID:               dmeBoat.SlipID,
			Slip:                 dmeBoat.Slip,
			Motors:               dmeBoat.Motors,
			Drives:               dmeBoat.Drives,
			Generators:           dmeBoat.Generators,
			DoNotLaunch:          dmeBoat.DoNotLaunch,
			BillingCodes:         dmeBoat.BillingCodes,
			BoatDescriptionCodes: dmeBoat.BoatDescriptionCodes,
			CustomInformation:    dmeBoat.CustomInformation,
			OperationsHistory:    dmeBoat.OperationsHistory,
			IntegrationID:        dmeBoat.IntegrationID,
			OwnerIntegrationID:   dmeBoat.OwnerIntegrationID,
			LastModified:         dmeBoat.LastModified,
			Comments:             dmeBoat.Comments,
			Attachments:          updatedAttachments,
		}

		_, err = h.server.DME.UpdateBoat(ctx, boatUpdate, orgID, systemID)
		if err != nil {
			h.server.Logger.Zap.Error("[DME API] Error updating boat attachments in DME (document)", err)
		} else {
			h.server.Logger.Zap.Info("[DME API] Successfully updated DME boat with document attachment",
				"boatID", entityID,
				"fileName", header.Filename,
				"documentID", doc.ID.String())
		}
	}()

	return response.JSON(c)
}

// GetDocumentsByEntity retrieves all documents for a boat entity
//
//	@Summary		Get boat documents
//	@Description	Retrieves all documents for a boat entity. External users (customers) only see public documents, internal users (marina staff) see all documents.
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

	// Get current user to determine if they are external (customer) or internal (marina staff)
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	isInternalUser := claims.IsCustomer == nil || !*claims.IsCustomer

	// Use single query with visibility filtering based on user type
	documents, err := h.server.DB.Queries().ListDocumentsByEntityWithVisibility(c.Request().Context(), db.ListDocumentsByEntityWithVisibilityParams{
		MarinaID:   marinaID,
		EntityType: entityType,
		EntityID:   entityID,
		Column4:    isInternalUser, // true (internal) = show all, false (external) = show only public
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching documents", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching documents").JSON(c)
	}

	// Return documents
	return responses.NewDocumentsResponseSuccess(documents).JSON(c)
}

// UpdateBoatDocumentPublic updates the public field of a boat document
//
//	@Summary		Update boat document public field
//	@Description	Updates the public field of a boat document and syncs with DME
//	@Tags			Documents
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string	true	"Document ID"	Format(uuid)
//	@Param			request	body		requests.UpdateBoatDocumentPublicRequest	true	"Update request"
//	@Success		200		{object}	responses.BaseResponse{data=responses.DocumentResponse}
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		404		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/documents/boat/{id}/public [put]
func (h *DocumentHandler) UpdateBoatDocumentPublic(c echo.Context) error {
	// Parse document ID from path
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing document ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid document ID format").JSON(c)
	}

	// Parse request body
	var req requests.UpdateBoatDocumentPublicRequest
	if err := c.Bind(&req); err != nil {
		h.server.Logger.Zap.Error("Error binding request", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request format").JSON(c)
	}

	// Validate request
	if err := c.Validate(&req); err != nil {
		h.server.Logger.Zap.Error("Error validating request", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Validation failed: "+err.Error()).JSON(c)
	}

	// Get the existing document to verify it's a boat document and get its details
	existingDoc, err := h.server.DB.Queries().GetDocumentByID(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching document", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Document not found").JSON(c)
	}

	// Verify this is a boat document
	if existingDoc.EntityType != "boat" {
		return responses.NewErrorResponse(http.StatusBadRequest, "This endpoint is only for boat documents").JSON(c)
	}

	// Update the document's public field
	updatedDoc, err := h.server.DB.Queries().UpdateDocument(c.Request().Context(), db.UpdateDocumentParams{
		ID:       id,
		FileName: existingDoc.FileName,
		FileType: existingDoc.FileType,
		FilePath: existingDoc.FilePath,
		FileSize: existingDoc.FileSize,
		Public:   req.Public,
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error updating document", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating document").JSON(c)
	}

	// Get marina information for DME operations
	ctx, cancel := context.WithTimeout(c.Request().Context(), 10*time.Second)
	defer cancel()

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, existingDoc.MarinaID)
	if err != nil {
		h.server.Logger.Zap.Error("[DME API] Error fetching marina for DME update (document public)", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching marina").JSON(c)
	}

	// Update DME attachment metadata asynchronously
	go func() {
		ctx := context.Background()

		if marina.SystemID == nil {
			h.server.Logger.Zap.Warn("[DME API] Marina has no system ID, skipping DME boat update (document public)")
			return
		}
		orgID := marina.OrganizationID
		systemID := *marina.SystemID

		dmeBoat, err := h.server.DME.RetrieveBoatByID(ctx, existingDoc.EntityID, orgID, systemID)
		if err != nil {
			h.server.Logger.Zap.Error("[DME API] Error retrieving boat from DME for attachment update (document public)", err)
			return
		}

		// Find and update the specific attachment in DME
		updatedAttachments := dmeBoat.Attachments
		if updatedAttachments == nil {
			updatedAttachments = []dme.Attachment{}
		}

		// Find the attachment by S3 path and update its public status
		attachmentFound := false
		for i, attachment := range updatedAttachments {
			if attachment.S3Path == existingDoc.FilePath {
				// Update the attachment's public status
				// Note: DME Attachment struct might not have a public field directly
				// We'll update the description to reflect the public status
				updatedAttachments[i].Description = fmt.Sprintf("Document attachment (Public: %t)", req.Public)
				attachmentFound = true
				break
			}
		}

		if !attachmentFound {
			h.server.Logger.Zap.Warn("[DME API] Attachment not found in DME boat for public status update",
				"boatID", existingDoc.EntityID,
				"s3Path", existingDoc.FilePath,
				"documentID", id.String())
			return
		}

		boatUpdate := &dme.BoatUpdate{
			ID:                   dmeBoat.ID,
			Name:                 dmeBoat.Name,
			Registration:         dmeBoat.Registration,
			Year:                 dmeBoat.Year,
			Make:                 dmeBoat.Make,
			Model:                dmeBoat.Model,
			HIN:                  dmeBoat.HIN,
			LOA:                  dmeBoat.LOA,
			LWL:                  dmeBoat.LWL,
			Draft:                dmeBoat.Draft,
			Beam:                 dmeBoat.Beam,
			Height:               dmeBoat.Height,
			Color:                dmeBoat.Color,
			TrailerMake:          dmeBoat.TrailerMake,
			TrailerModel:         dmeBoat.TrailerModel,
			TrailerSerial:        dmeBoat.TrailerSerial,
			TrailerRegistration:  dmeBoat.TrailerRegistration,
			TrailerLocation:      dmeBoat.TrailerLocation,
			SummerSlip:           dmeBoat.SummerSlip,
			WinterSlip:           dmeBoat.WinterSlip,
			InsuranceCompany:     dmeBoat.InsuranceCompany,
			InsuranceExpDate:     dmeBoat.InsuranceExpDate,
			SlipID:               dmeBoat.SlipID,
			Slip:                 dmeBoat.Slip,
			Motors:               dmeBoat.Motors,
			DoNotLaunch:          dmeBoat.DoNotLaunch,
			BillingCodes:         dmeBoat.BillingCodes,
			BoatDescriptionCodes: dmeBoat.BoatDescriptionCodes,
			CustomInformation:    dmeBoat.CustomInformation,
			OperationsHistory:    dmeBoat.OperationsHistory,
			IntegrationID:        dmeBoat.IntegrationID,
			OwnerIntegrationID:   dmeBoat.OwnerIntegrationID,
			LastModified:         dmeBoat.LastModified,
			Comments:             dmeBoat.Comments,
			Attachments:          updatedAttachments,
		}

		_, err = h.server.DME.UpdateBoat(ctx, boatUpdate, orgID, systemID)
		if err != nil {
			h.server.Logger.Zap.Error("[DME API] Error updating boat attachment public status in DME", err)
		} else {
			h.server.Logger.Zap.Info("[DME API] Successfully updated DME boat attachment public status",
				"boatID", existingDoc.EntityID,
				"fileName", existingDoc.FileName,
				"documentID", id.String(),
				"public", req.Public)
		}
	}()

	// Return updated document
	return responses.NewDocumentResponseSuccess(updatedDoc).JSON(c)
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

	// Check storage limit before upload
	// if err := h.checkStorageLimit(c, marinaID, header.Size); err != nil {
	// 	h.server.Logger.Zap.Error("Storage limit check failed", err)
	// 	return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	// }

	// Upload the file to S3 using document storage service
	filePath, err := h.server.DocumentService.UploadFileToS3(c.Request().Context(), file, header, entityType)
	if err != nil {
		h.server.Logger.Zap.Error("Error uploading file to S3", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading file: "+err.Error()).JSON(c)
	}

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
		h.server.Logger.Zap.Error("Error creating document in database", err)
		// TODO: Delete file from S3 since document creation failed
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating document: "+err.Error()).JSON(c)
	}

	// Update marina storage usage
	if err := h.updateStorageUsage(c, marinaID, header.Size, true); err != nil {
		h.server.Logger.Zap.Error("Error updating storage usage", err)
		// TODO: Delete file from S3 and document record since storage update failed
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating storage usage").JSON(c)
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

	// Get document to get its size before deletion
	doc, err := h.server.DB.Queries().GetDocumentByID(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching document", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Document not found").JSON(c)
	}

	// Delete document from database
	err = h.server.DB.Queries().DeleteDocument(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Error("Error deleting document", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error deleting document").JSON(c)
	}

	// Update marina storage usage
	if err := h.updateStorageUsage(c, doc.MarinaID, doc.FileSize, false); err != nil {
		// Log error but don't fail the request
		h.server.Logger.Zap.Error("Error updating storage usage", err)
	}

	// Return success message
	return responses.NewMessageResponse(http.StatusOK, "Document successfully deleted").JSON(c)
}
