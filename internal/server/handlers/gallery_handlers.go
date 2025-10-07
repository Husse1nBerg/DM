package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// GalleryHandler handles operations related to marina and vessel galleries
type GalleryHandler struct {
	server *s.Server
}

// NewGalleryHandler creates a new gallery handler
func NewGalleryHandler(server *s.Server) *GalleryHandler {
	return &GalleryHandler{server: server}
}

// checkStorageLimit checks if the marina has enough storage space for the new file
func (h *GalleryHandler) checkStorageLimit(ctx echo.Context, marinaID uuid.UUID, fileSize int64) error {
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
func (h *GalleryHandler) updateStorageUsage(ctx echo.Context, marinaID uuid.UUID, fileSize int64, isIncrement bool) error {
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

// CreateMarinaGalleryItem creates a new marina gallery item
//
//	@Summary		Create marina gallery item
//	@Description	Creates a new gallery item for a marina
//	@Tags			Marina Gallery
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			marinaId		formData	string	true	"Marina ID"	Format(uuid)
//	@Param			description		formData	string	false	"Image description"
//	@Param			image			formData	file	true	"Image file"
//	@Success		201				{object}	responses.BaseResponse{data=responses.MarinaGalleryItemResponse}
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		404				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/gallery/marina [post]
func (h *GalleryHandler) CreateMarinaGalleryItem(c echo.Context) error {
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

	// Get description from form
	description := c.FormValue("description")
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}

	// Get image file from form
	file, header, err := c.Request().FormFile("image")
	if err != nil {
		h.server.Logger.Zap.Error("Error getting image file", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Image file is required").JSON(c)
	}
	defer file.Close()

	// Check storage limit before upload
	// if err := h.checkStorageLimit(c, marinaID, header.Size); err != nil {
	// 	h.server.Logger.Zap.Error("Storage limit check failed", err)
	// 	return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	// }

	// Upload the image to S3
	imagePath, err := h.server.ImageService.UploadImage(c.Request().Context(), file, header, s3.MarinaGalleryImageType)
	if err != nil {
		h.server.Logger.Zap.Error("Error uploading image to S3", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading image: "+err.Error()).JSON(c)
	}

	// Update storage usage
	err = h.updateStorageUsage(c, marinaID, header.Size, true)
	if err != nil {
		h.server.Logger.Zap.Error("Error updating marina storage usage", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating storage usage").JSON(c)
	}

	// Create gallery item in database
	createParams := db.CreateMarinaGalleryItemParams{
		MarinaID:    marinaID,
		ImageUrl:    imagePath,
		Description: descriptionPtr,
	}

	galleryItem, err := h.server.DB.Queries().CreateMarinaGalleryItem(c.Request().Context(), createParams)
	if err != nil {
		h.server.Logger.Zap.Error("Error creating marina gallery item", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating gallery item").JSON(c)
	}

	// Return created gallery item
	response := responses.NewMarinaGalleryItemResponseSuccess(galleryItem)
	response.Code = http.StatusCreated
	return response.JSON(c)
}

// GetMarinaGallery retrieves all gallery items for a marina
//
//	@Summary		Get marina gallery
//	@Description	Retrieves all gallery items for a marina
//	@Tags			Marina Gallery
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	path		string	true	"Marina ID"	Format(uuid)
//	@Success		200			{array}		responses.BaseResponse{data=[]responses.MarinaGalleryItemResponse}
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/gallery/marina/{marinaId} [get]
func (h *GalleryHandler) GetMarinaGallery(c echo.Context) error {
	// Parse marina ID from path
	marinaIDStr := c.Param("marinaId")
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

	// Get gallery items from database
	galleryItems, err := h.server.DB.Queries().GetMarinaGallery(c.Request().Context(), marinaID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina gallery items", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching gallery items").JSON(c)
	}

	// Return gallery items
	return responses.NewMarinaGalleryResponseSuccess(galleryItems).JSON(c)
}

// GetMarinaGalleryItem retrieves a specific gallery item
//
//	@Summary		Get marina gallery item
//	@Description	Retrieves a specific gallery item by ID
//	@Tags			Marina Gallery
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Gallery Item ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse{data=responses.MarinaGalleryItemResponse}
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/gallery/marina/item/{id} [get]
func (h *GalleryHandler) GetMarinaGalleryItem(c echo.Context) error {
	// Parse gallery item ID from path
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing gallery item ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid gallery item ID format").JSON(c)
	}

	// Get gallery item from database
	galleryItem, err := h.server.DB.Queries().GetMarinaGalleryItemByID(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina gallery item", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Gallery item not found").JSON(c)
	}

	// Return gallery item
	return responses.NewMarinaGalleryItemResponseSuccess(galleryItem).JSON(c)
}

// UpdateMarinaGalleryItem updates a marina gallery item
//
//	@Summary		Update marina gallery item
//	@Description	Updates a marina gallery item
//	@Tags			Marina Gallery
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id			path		string	true	"Gallery Item ID"	Format(uuid)
//	@Param			description	formData	string	false	"Image description"
//	@Param			image		formData	file	false	"Image file"
//	@Success		200			{object}	responses.BaseResponse{data=responses.MarinaGalleryItemResponse}
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/gallery/marina/item/{id} [put]
func (h *GalleryHandler) UpdateMarinaGalleryItem(c echo.Context) error {
	// Parse gallery item ID from path
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing gallery item ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid gallery item ID format").JSON(c)
	}

	// Get existing gallery item to update
	existingItem, err := h.server.DB.Queries().GetMarinaGalleryItemByID(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching marina gallery item", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Gallery item not found").JSON(c)
	}

	// Initialize update parameters with current values
	imageUrl := existingItem.ImageUrl
	description := existingItem.Description
	public := existingItem.Public
	var fileSize int64

	// Check if there's an image file in the form
	file, header, err := c.Request().FormFile("image")
	if err == nil {
		defer file.Close()

		// Check storage limit before upload
		// if err := h.checkStorageLimit(c, existingItem.MarinaID, header.Size); err != nil {
		// 	h.server.Logger.Zap.Error("Storage limit check failed", err)
		// 	return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
		// }

		// Upload the new image to S3
		imagePath, err := h.server.ImageService.UploadImage(c.Request().Context(), file, header, s3.MarinaGalleryImageType)
		if err != nil {
			h.server.Logger.Zap.Error("Error uploading image to S3", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading image: "+err.Error()).JSON(c)
		}

		// Update the image path
		imageUrl = imagePath
	}

	// Parse description from form if provided
	if desc := c.FormValue("description"); desc != "" {
		description = &desc
	}

	// Parse public flag from form if provided
	if publicStr := c.FormValue("public"); publicStr != "" {
		public = publicStr == "false"
	}

	// Update gallery item in database
	updateParams := db.UpdateMarinaGalleryItemParams{
		ID:          id,
		ImageUrl:    imageUrl,
		Description: description,
		Public:      public,
	}

	updatedItem, err := h.server.DB.Queries().UpdateMarinaGalleryItem(c.Request().Context(), updateParams)
	if err != nil {
		h.server.Logger.Zap.Error("Error updating marina gallery item", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating gallery item").JSON(c)
	}

	// Update storage usage only if a new file was uploaded
	if fileSize > 0 {
		err = h.updateStorageUsage(c, existingItem.MarinaID, fileSize, true)
		if err != nil {
			h.server.Logger.Zap.Error("Error updating marina storage usage", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating storage usage").JSON(c)
		}
	}

	// Return updated gallery item
	return responses.NewMarinaGalleryItemResponseSuccess(updatedItem).JSON(c)
}

// DeleteMarinaGalleryItem deletes a marina gallery item
//
//	@Summary		Delete marina gallery item
//	@Description	Deletes a marina gallery item
//	@Tags			Marina Gallery
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Gallery Item ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/gallery/marina/item/{id} [delete]
func (h *GalleryHandler) DeleteMarinaGalleryItem(c echo.Context) error {
	// Parse gallery item ID from path
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing gallery item ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid gallery item ID format").JSON(c)
	}
	// Soft delete the gallery item
	err = h.server.DB.Queries().SoftDeleteMarinaGalleryItem(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Error("Error deleting marina gallery item", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error deleting gallery item").JSON(c)
	}

	// Return success message
	return responses.NewMessageResponse(http.StatusOK, "Gallery item successfully deleted").JSON(c)
}

// CreateVesselGalleryItem creates a new vessel gallery item
//
//	@Summary		Create vessel gallery item
//	@Description	Creates a new gallery item for a vessel
//	@Tags			Vessel Gallery
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			marinaId		formData	string	true	"Marina ID"	Format(uuid)
//	@Param			customerId		formData	string	true	"Customer ID"
//	@Param			boatId			formData	string	true	"Boat ID"
//	@Param			description		formData	string	false	"Image description"
//	@Param			main			formData	boolean	false	"Whether this is the main image"
//	@Param			image			formData	file	true	"Image file"
//	@Success		201				{object}	responses.BaseResponse{data=responses.VesselGalleryItemResponse}
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		404				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/gallery/boat [post]
func (h *GalleryHandler) CreateVesselGalleryItem(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)

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

	// Get customer ID from form
	customerID := c.FormValue("customerId")
	if customerID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Customer ID is required").JSON(c)
	}

	// Get vessel ID from form
	boatID := c.FormValue("boatId")
	if boatID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Boat ID is required").JSON(c)
	}

	// Get description from form
	description := c.FormValue("description")
	var descriptionPtr *string
	if description != "" {
		descriptionPtr = &description
	}

	// Get main flag from form
	mainStr := c.FormValue("main")
	var mainPtr *bool
	if mainStr != "" {
		isMain := mainStr == "true"
		mainPtr = &isMain

		// If this is the main image, update existing main images to not be main
		if isMain {
			removeMainParams := db.RemoveMainVesselImageParams{
				VesselID:   boatID,
				CustomerID: customerID,
				MarinaID:   marinaID,
			}
			err = h.server.DB.Queries().RemoveMainVesselImage(c.Request().Context(), removeMainParams)
			if err != nil {
				h.server.Logger.Zap.Error("Error removing main flag from existing images", err)
				// Continue even if this fails
			}
		}
	}

	// Get image file from form
	file, header, err := c.Request().FormFile("image")
	if err != nil {
		h.server.Logger.Zap.Error("Error getting image file", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Image file is required").JSON(c)
	}
	defer file.Close()

	// Check storage limit before upload
	// if err := h.checkStorageLimit(c, marinaID, header.Size); err != nil {
	// 	h.server.Logger.Zap.Error("Storage limit check failed", err)
	// 	return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
	// }

	// Upload the image to S3
	imagePath, err := h.server.ImageService.UploadImage(c.Request().Context(), file, header, s3.VesselGalleryImageType)
	if err != nil {
		h.server.Logger.Zap.Error("Error uploading image to S3", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading image: "+err.Error()).JSON(c)
	}

	// Determine public flag: External users (IsCustomer=true) should have Public=true, Internal users (IsCustomer=false or nil) should have Public=false
	var public bool
	if claims.IsCustomer != nil {
		public = *claims.IsCustomer
	} else {
		public = false // nil IsCustomer is treated as internal user
	}

	// Create gallery item in database
	createParams := db.CreateVesselGalleryItemParams{
		MarinaID:    marinaID,
		CustomerID:  customerID,
		VesselID:    boatID,
		ImageUrl:    imagePath,
		Description: descriptionPtr,
		Main:        mainPtr,
		Public:      public,
	}

	galleryItem, err := h.server.DB.Queries().CreateVesselGalleryItem(c.Request().Context(), createParams)
	if err != nil {
		h.server.Logger.Zap.Error("Error creating vessel gallery item", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating gallery item").JSON(c)
	}

	// Update storage usage
	err = h.updateStorageUsage(c, marinaID, header.Size, true)
	if err != nil {
		h.server.Logger.Zap.Error("Error updating marina storage usage", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating storage usage").JSON(c)
	}

	go func() {
		ctx := context.Background()

		marina, err := h.server.DB.Queries().GetMarinaByID(ctx, marinaID)
		if err != nil {
			h.server.Logger.Zap.Error("[DME API] Error fetching marina for DME update", err)
			return
		}
		if marina.SystemID == nil {
			h.server.Logger.Zap.Warn("[DME API] Marina has no system ID, skipping DME boat update")
			return
		}
		orgID := marina.OrganizationID
		systemID := *marina.SystemID

		dmeBoat, err := h.server.DME.RetrieveBoatByID(ctx, boatID, orgID, systemID)
		if err != nil {
			h.server.Logger.Zap.Error("[DME API] Error retrieving boat from DME for attachment update", err)
			return
		}

		desc := ""
		if descriptionPtr != nil {
			desc = *descriptionPtr
		}
		fileType := header.Header.Get("Content-Type")
		newAttachment := dme.Attachment{
			FileName:    header.Filename,
			Description: desc,
			S3Path:      imagePath,
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
			h.server.Logger.Zap.Error("[DME API] Error updating boat attachments in DME", err)
		} else {
			h.server.Logger.Zap.Info("[DME API] Successfully updated DME boat with gallery attachment",
				"boatID", boatID,
				"fileName", header.Filename,
				"galleryItemID", galleryItem.ID.String())
		}
	}()

	// Return created gallery item
	response := responses.NewVesselGalleryItemResponseSuccess(galleryItem)
	response.Code = http.StatusCreated

	return response.JSON(c)
}

// GetVesselGallery retrieves all gallery items for a vessel
//
//	@Summary		Get vessel gallery
//	@Description	Retrieves all gallery items for a vessel. External users (customers) only see public images, internal users (marina staff) see all images.
//	@Tags			Vessel Gallery
//	@Accept			json
//	@Produce		json
//	@Param			marinaId	query		string	true	"Marina ID"	Format(uuid)
//	@Param			customerId	query		string	true	"Customer ID"
//	@Param			boatId		path		string	true	"Boat ID"
//	@Success		200			{array}		responses.BaseResponse{data=[]responses.VesselGalleryItemResponse}
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/gallery/boat/{boatId} [get]
func (h *GalleryHandler) GetVesselGallery(c echo.Context) error {
	// Parse marina ID from query
	marinaIDStr := c.QueryParam("marinaId")
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

	// Get customer ID from query
	customerID := c.QueryParam("customerId")
	if customerID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Customer ID is required").JSON(c)
	}

	// Get vessel ID from path
	boatID := c.Param("boatId")
	if boatID == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Boat ID is required").JSON(c)
	}

	// Get current user to determine if they are external (customer) or internal (marina staff)
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	isInternalUser := claims.IsCustomer == nil || !*claims.IsCustomer

	// Use single query with visibility filtering based on user type
	galleryItems, err := h.server.DB.Queries().GetVesselGalleryWithVisibility(c.Request().Context(), db.GetVesselGalleryWithVisibilityParams{
		VesselID:   boatID,
		CustomerID: customerID,
		MarinaID:   marinaID,
		Column4:    isInternalUser, // true (internal) = show all, false (external) = show only public
	})
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching vessel gallery items", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error fetching gallery items").JSON(c)
	}

	// Return gallery items
	return responses.NewVesselGalleryResponseSuccess(galleryItems).JSON(c)
}

// GetVesselGalleryItem retrieves a specific vessel gallery item
//
//	@Summary		Get vessel gallery item
//	@Description	Retrieves a specific vessel gallery item by ID
//	@Tags			Vessel Gallery
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Gallery Item ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse{data=responses.VesselGalleryItemResponse}
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/gallery/boat/item/{id} [get]
func (h *GalleryHandler) GetVesselGalleryItem(c echo.Context) error {
	// Parse gallery item ID from path
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing gallery item ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid gallery item ID format").JSON(c)
	}

	// Get gallery item from database
	galleryItem, err := h.server.DB.Queries().GetVesselGalleryItemByID(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching vessel gallery item", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Gallery item not found").JSON(c)
	}

	// Return gallery item
	return responses.NewVesselGalleryItemResponseSuccess(galleryItem).JSON(c)
}

// UpdateVesselGalleryItem updates a vessel gallery item
//
//	@Summary		Update vessel gallery item
//	@Description	Updates a vessel gallery item
//	@Tags			Vessel Gallery
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			id			path		string	true	"Gallery Item ID"	Format(uuid)
//	@Param			description	formData	string	false	"Image description"
//	@Param			main		formData	boolean	false	"Whether this is the main image"
//	@Param			image		formData	file	false	"Image file"
//	@Success		200			{object}	responses.BaseResponse{data=responses.VesselGalleryItemResponse}
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/gallery/boat/item/{id} [put]
func (h *GalleryHandler) UpdateVesselGalleryItem(c echo.Context) error {
	// Parse gallery item ID from path
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing gallery item ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid gallery item ID format").JSON(c)
	}

	// Get existing gallery item to update
	existingItem, err := h.server.DB.Queries().GetVesselGalleryItemByID(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching vessel gallery item", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Gallery item not found").JSON(c)
	}

	// Initialize update parameters with current values
	imageUrl := existingItem.ImageUrl
	description := existingItem.Description
	var fileSize int64

	// Check if there's an image file in the form
	file, header, err := c.Request().FormFile("image")
	if err == nil {
		defer file.Close()

		// Check storage limit before upload
		// if err := h.checkStorageLimit(c, existingItem.MarinaID, header.Size); err != nil {
		// 	h.server.Logger.Zap.Error("Storage limit check failed", err)
		// 	return responses.NewErrorResponse(http.StatusBadRequest, err.Error()).JSON(c)
		// }

		// Upload the new image to S3
		imagePath, err := h.server.ImageService.UploadImage(c.Request().Context(), file, header, s3.VesselGalleryImageType)
		if err != nil {
			h.server.Logger.Zap.Error("Error uploading image to S3", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading image: "+err.Error()).JSON(c)
		}

		// Update the image path
		imageUrl = imagePath
	}

	// Parse description from form if provided
	if desc := c.FormValue("description"); desc != "" {
		description = &desc
	}

	// Update gallery item in database
	updateParams := db.UpdateVesselGalleryItemParams{
		ID:          id,
		ImageUrl:    imageUrl,
		Description: description,
	}

	updatedItem, err := h.server.DB.Queries().UpdateVesselGalleryItem(c.Request().Context(), updateParams)
	if err != nil {
		h.server.Logger.Zap.Error("Error updating vessel gallery item", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating gallery item").JSON(c)
	}

	// Update storage usage only if a new file was uploaded
	if fileSize > 0 {
		err = h.updateStorageUsage(c, existingItem.MarinaID, fileSize, true)
		if err != nil {
			h.server.Logger.Zap.Error("Error updating marina storage usage", err)
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating storage usage").JSON(c)
		}
	}

	// Return updated gallery item
	return responses.NewVesselGalleryItemResponseSuccess(updatedItem).JSON(c)
}

// DeleteVesselGalleryItem deletes a vessel gallery item
//
//	@Summary		Delete vessel gallery item
//	@Description	Deletes a vessel gallery item
//	@Tags			Vessel Gallery
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Gallery Item ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/gallery/boat/item/{id} [delete]
func (h *GalleryHandler) DeleteVesselGalleryItem(c echo.Context) error {
	// Parse gallery item ID from path
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing gallery item ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid gallery item ID format").JSON(c)
	}

	// Soft delete the gallery item
	err = h.server.DB.Queries().SoftDeleteVesselGalleryItem(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Error("Error deleting vessel gallery item", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error deleting gallery item").JSON(c)
	}

	// Return success message
	return responses.NewMessageResponse(http.StatusOK, "Gallery item successfully deleted").JSON(c)
}
