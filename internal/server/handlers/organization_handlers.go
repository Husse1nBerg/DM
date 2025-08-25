package handlers

import (
	"math"
	"net/http"
	"strings"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// OrganizationHandler handles organization-related operations
type OrganizationHandler struct {
	server *s.Server
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(server *s.Server) *OrganizationHandler {
	return &OrganizationHandler{server: server}
}

// CreateOrganization creates a new organization
//
//	@Summary		Create organization
//	@Description	Creates a new organization in the system
//	@Tags			Organizations
//	@Accept			json
//	@Produce		json
//	@Param			organization	body		requests.CreateOrganizationRequest	true	"Organization details"
//	@Success		201				{object}	responses.OrganizationResponse
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/organizations [post]
func (h *OrganizationHandler) CreateOrganization(c echo.Context) error {
	var req requests.CreateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Set default values
	isActive := req.IsActive
	isTest := req.IsTest

	// Create address if provided
	addressID := uuid.New()

	if req.Address != nil {
		// Convert numeric values
		var lat, lng float64
		if req.Address.Latitude != nil {
			lat = *req.Address.Latitude
		}
		if req.Address.Longitude != nil {
			lng = *req.Address.Longitude
		}

		// Create the address
		addrParams := db.CreateAddressParams{
			Street:     req.Address.Street,
			City:       req.Address.City,
			State:      req.Address.State,
			PostalCode: req.Address.PostalCode,
			Country:    req.Address.Country,
			Latitude:   lat,
			Longitude:  lng,
		}

		address, err := h.server.DB.Queries().CreateAddress(c.Request().Context(), addrParams)
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating address").JSON(c)
		}

		addressID = address.ID
	}

	params := db.CreateOrganizationParams{
		Email:     utils.LowerCase(req.Email),
		Name:      req.Name,
		Image:     req.Image,
		Website:   req.Website,
		Country:   req.Country,
		Phone:     req.Phone,
		IsActive:  &isActive,
		IsTest:    &isTest,
		AddressID: addressID,
	}

	org, err := h.server.DB.Queries().CreateOrganization(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.NewSuccessResponse(responses.ConvertOrganizationToResponse(org))
	response.Code = http.StatusCreated
	return response.JSON(c)
}

// GetOrganizationByID retrieves an organization by its ID
//
//	@Summary		Get organization by ID
//	@Description	Retrieves an organization by its ID
//	@Tags			Organizations
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Organization ID"	Format(uuid)
//	@Success		200	{object}	responses.OrganizationResponse
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/organizations/{id} [get]
func (h *OrganizationHandler) GetOrganizationByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID").JSON(c)
	}

	org, err := h.server.DB.Queries().GetOrganizationByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Organization not found").JSON(c)
	}

	return responses.NewOrganizationResponseSuccess(org).JSON(c)
}

// GetOrganizationWithAddress retrieves an organization and its address
//
//	@Summary		Get organization with address
//	@Description	Retrieves an organization and its address details
//	@Tags			Organizations
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Organization ID"	Format(uuid)
//	@Success		200	{object}	responses.OrganizationWithAddressResponse
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/organizations/{id}/with-address [get]
func (h *OrganizationHandler) GetOrganizationWithAddress(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID").JSON(c)
	}

	org, err := h.server.DB.Queries().GetOrganizationByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Organization not found").JSON(c)
	}

	// Get the address
	address, err := h.server.DB.Queries().GetAddressByID(c.Request().Context(), org.AddressID)
	if err != nil {
		// Return organization without address if address not found
		return responses.NewOrganizationWithAddressResponse(org, nil).JSON(c)
	}

	return responses.NewOrganizationWithAddressResponse(org, &address).JSON(c)
}

// GetOrganizationByEmail retrieves an organization by its email
//
//	@Summary		Get organization by email
//	@Description	Retrieves an organization by its email address
//	@Tags			Organizations
//	@Accept			json
//	@Produce		json
//	@Param			email	query		string	true	"Organization email"
//	@Success		200		{object}	responses.OrganizationResponse
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		404		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/organizations/by-email [get]
func (h *OrganizationHandler) GetOrganizationByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if email == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Email parameter is required").JSON(c)
	}

	org, err := h.server.DB.Queries().GetOrganizationByEmail(c.Request().Context(), email)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Organization not found").JSON(c)
	}

	return responses.NewOrganizationResponseSuccess(org).JSON(c)
}

// GetOrganizationsPaginated retrieves organizations with filtering, sorting, and pagination
//
//	@Summary		Get paginated organizations
//	@Description	Retrieves organizations with filtering, sorting, and pagination support
//	@Tags			Organizations
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number" default(1)
//	@Param			pageSize	query		int		false	"Page size" default(10)
//	@Param			sortBy		query		string	false	"Sort by field (name, email, website, country, phone, is_active, is_test, created_at, updated_at)"
//	@Param			sortOrder	query		string	false	"Sort order (asc, desc)" default(asc)
//	@Param			search		query		string	false	"Global search across multiple fields"
//	@Param			isActive	query		bool	false	"Filter by active status"
//	@Param			isTest		query		bool	false	"Filter by test status"
//	@Success		200		{array}		responses.OrganizationsPaginatedResponse
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/organizations [get]
func (h *OrganizationHandler) GetOrganizationsPaginated(c echo.Context) error {
	var req requests.ListOrganizationsRequest

	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Set defaults
	if req.PageSize <= 0 {
		req.PageSize = 10
	}
	if req.Page <= 0 {
		req.Page = 1
	}
	if req.SortOrder == "" {
		req.SortOrder = "asc"
	}
	if req.SortBy == "" {
		req.SortBy = "created_at"
	}

	// Prepare filter params
	search := req.Search
	isActive := ""
	if v, ok := req.Filters["isActive"]; ok {
		isActive = v
	}
	isTest := ""
	if v, ok := req.Filters["isTest"]; ok {
		isTest = v
	}

	// Count total (for pagination)
	allOrgsAsc, err := h.server.DB.Queries().GetOrganizationsWithFiltersAsc(c.Request().Context(), db.GetOrganizationsWithFiltersAscParams{
		Column1: search,
		Column2: isActive,
		Column3: isTest,
		Column4: req.SortBy,
		Limit:   int32(math.MaxInt32),
		Offset:  0,
	})
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allOrgsAsc))

	// Choose ASC or DESC query
	var orgs []db.Organization
	if req.SortOrder == "desc" {
		orgs, err = h.server.DB.Queries().GetOrganizationsWithFiltersDesc(c.Request().Context(), db.GetOrganizationsWithFiltersDescParams{
			Column1: search,
			Column2: isActive,
			Column3: isTest,
			Column4: req.SortBy,
			Limit:   req.PageSize,
			Offset:  (req.Page - 1) * req.PageSize,
		})
	} else {
		orgs, err = h.server.DB.Queries().GetOrganizationsWithFiltersAsc(c.Request().Context(), db.GetOrganizationsWithFiltersAscParams{
			Column1: search,
			Column2: isActive,
			Column3: isTest,
			Column4: req.SortBy,
			Limit:   req.PageSize,
			Offset:  (req.Page - 1) * req.PageSize,
		})
	}
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	currentPage := req.Page
	return responses.NewOrganizationsPaginatedResponse(orgs, total, req.PageSize, currentPage).JSON(c)
}

// UpdateOrganization updates an existing organization
//
//	@Summary		Update organization
//	@Description	Updates an existing organization
//	@Tags			Organizations
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string								true	"Organization ID"	Format(uuid)
//	@Param			organization	body		requests.UpdateOrganizationRequest	true	"Organization details"
//	@Success		200				{object}	responses.OrganizationResponse
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		404				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/organizations/{id} [put]
func (h *OrganizationHandler) UpdateOrganization(c echo.Context) error {
	// Parse organization ID from path parameter
	orgIDStr := c.Param("id")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		h.server.Logger.Zap.Error("Error parsing organization ID", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID format").JSON(c)
	}

	// Get existing organization to update
	queries := h.server.DB.Queries()
	org, err := queries.GetOrganizationByID(c.Request().Context(), orgID)
	if err != nil {
		h.server.Logger.Zap.Error("Error fetching organization", err)
		return responses.NewErrorResponse(http.StatusNotFound, "Organization not found").JSON(c)
	}

	// Check if this is a multipart form (which would include a file upload)
	contentType := c.Request().Header.Get("Content-Type")
	isMultipart := strings.HasPrefix(contentType, "multipart/form-data")

	// Initialize update parameters with current values
	updateParams := db.UpdateOrganizationParams{
		ID:        orgID,
		Email:     utils.LowerCase(org.Email),
		Name:      org.Name,
		Image:     org.Image,
		Website:   org.Website,
		Country:   org.Country,
		Phone:     org.Phone,
		IsActive:  org.IsActive,
		IsTest:    org.IsTest,
		AddressID: org.AddressID,
	}

	if isMultipart {
		// Check if there's an image file in the form
		file, header, err := c.Request().FormFile("image")
		if err == nil {
			defer file.Close()

			// Upload the image to S3
			imagePath, err := h.server.ImageService.UploadImage(c.Request().Context(), file, header, s3.OrganizationImageType)
			if err != nil {
				h.server.Logger.Zap.Error("Error uploading organization image to S3", err)
				return responses.NewErrorResponse(http.StatusInternalServerError, "Error uploading image: "+err.Error()).JSON(c)
			}

			// Update the image path
			updateParams.Image = &imagePath
		}

		// Parse other form fields
		if name := c.FormValue("name"); name != "" {
			updateParams.Name = name
		}
		if email := c.FormValue("email"); email != "" {
			updateParams.Email = utils.LowerCase(email)
		}
		if website := c.FormValue("website"); website != "" {
			updateParams.Website = &website
		}
		if country := c.FormValue("country"); country != "" {
			updateParams.Country = &country
		}
		if phone := c.FormValue("phone"); phone != "" {
			updateParams.Phone = &phone
		}
		// Parse boolean fields
		if isActiveStr := c.FormValue("isActive"); isActiveStr != "" {
			isActive := isActiveStr == "true"
			updateParams.IsActive = &isActive
		}
		if isTestStr := c.FormValue("isTest"); isTestStr != "" {
			isTest := isTestStr == "true"
			updateParams.IsTest = &isTest
		}
	} else {
		// Parse and validate the JSON request body
		req := new(requests.UpdateOrganizationRequest)
		if err := c.Bind(req); err != nil {
			h.server.Logger.Zap.Error("Error binding request", err)
			return responses.NewErrorResponse(http.StatusBadRequest, "Error parsing request").JSON(c)
		}
		if err := c.Validate(req); err != nil {
			h.server.Logger.Zap.Error("Error validating request", err)
			return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
		}

		// Update fields if provided in request
		if req.Email != nil {
			updateParams.Email = utils.LowerCase(*req.Email)
		}
		if req.Name != nil {
			updateParams.Name = *req.Name
		}
		if req.Image != nil {
			updateParams.Image = req.Image
		}
		if req.Website != nil {
			updateParams.Website = req.Website
		}
		if req.Country != nil {
			updateParams.Country = req.Country
		}
		if req.Phone != nil {
			updateParams.Phone = req.Phone
		}
		if req.IsActive != nil {
			updateParams.IsActive = req.IsActive
		}
		if req.IsTest != nil {
			updateParams.IsTest = req.IsTest
		}
	}

	// Update organization in database
	updatedOrg, err := queries.UpdateOrganization(c.Request().Context(), updateParams)
	if err != nil {
		h.server.Logger.Zap.Error("Error updating organization", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating organization").JSON(c)
	}

	// Return updated organization
	return responses.NewOrganizationResponseSuccess(updatedOrg).JSON(c)
}

// UpdateOrgAddress updates an organization and its address
//
//	@Summary		Update organization with address
//	@Description	Updates an organization and its address details
//	@Tags			Organizations
//	@Accept			json
//	@Produce		json
//	@Param			id				path		string								true	"Organization ID"	Format(uuid)
//	@Param			orgAddress	body		requests.UpdateOrgAddressRequest	true	"Organization and address details"
//	@Success		200				{object}	responses.OrganizationWithAddressResponse
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		404				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/organizations/{id}/with-address [put]
func (h *OrganizationHandler) UpdateOrgAddress(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID").JSON(c)
	}

	// Check if organization exists and get current values
	currentOrg, err := h.server.DB.Queries().GetOrganizationByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Organization not found").JSON(c)
	}

	var req requests.UpdateOrgAddressRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Update organization
	orgParams := db.UpdateOrganizationParams{
		ID:        id,
		Email:     currentOrg.Email,
		Name:      currentOrg.Name,
		Image:     currentOrg.Image,
		Website:   currentOrg.Website,
		Country:   currentOrg.Country,
		Phone:     currentOrg.Phone,
		IsActive:  currentOrg.IsActive,
		IsTest:    currentOrg.IsTest,
		AddressID: currentOrg.AddressID,
	}

	// Update only organization fields that are provided
	if req.Email != nil {
		orgParams.Email = utils.LowerCase(*req.Email)
	}
	if req.Name != nil {
		orgParams.Name = *req.Name
	}
	if req.Image != nil {
		orgParams.Image = req.Image
	}
	if req.Website != nil {
		orgParams.Website = req.Website
	}
	if req.Country != nil {
		orgParams.Country = req.Country
	}
	if req.Phone != nil {
		orgParams.Phone = req.Phone
	}
	if req.IsActive != nil {
		orgParams.IsActive = req.IsActive
	}
	if req.IsTest != nil {
		orgParams.IsTest = req.IsTest
	}

	updatedOrg, err := h.server.DB.Queries().UpdateOrganization(c.Request().Context(), orgParams)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating organization").JSON(c)
	}

	// Update address if provided
	var updatedAddress *db.Address
	if req.Address != nil {
		// Get current address
		currentAddress, err := h.server.DB.Queries().GetAddressByID(c.Request().Context(), currentOrg.AddressID)
		if err == nil { // Update only if address exists
			// Build address update params
			street := currentAddress.Street
			city := currentAddress.City
			state := currentAddress.State
			postalCode := currentAddress.PostalCode
			country := currentAddress.Country
			lat := currentAddress.Latitude
			lng := currentAddress.Longitude

			// Update fields that are provided
			if req.Address.Street != nil {
				street = req.Address.Street
			}
			if req.Address.City != nil {
				city = req.Address.City
			}
			if req.Address.State != nil {
				state = req.Address.State
			}
			if req.Address.PostalCode != nil {
				postalCode = req.Address.PostalCode
			}
			if req.Address.Country != nil {
				country = req.Address.Country
			}
			if req.Address.Latitude != nil {
				lat = *req.Address.Latitude
			}
			if req.Address.Longitude != nil {
				lng = *req.Address.Longitude
			}

			addrParams := db.UpdateAddressParams{
				ID:         currentOrg.AddressID,
				Street:     street,
				City:       city,
				State:      state,
				PostalCode: postalCode,
				Country:    country,
				Latitude:   lat,
				Longitude:  lng,
			}

			address, err := h.server.DB.Queries().UpdateAddress(c.Request().Context(), addrParams)
			if err == nil {
				updatedAddress = &address
			}
		}
	}

	return responses.NewOrganizationWithAddressResponse(updatedOrg, updatedAddress).JSON(c)
}

// DeleteOrganization deletes an organization (soft delete)
//
//	@Summary		Delete organization
//	@Description	Soft deletes an organization
//	@Tags			Organizations
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Organization ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/organizations/{id} [delete]
func (h *OrganizationHandler) DeleteOrganization(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID").JSON(c)
	}

	// Check if organization exists
	_, err = h.server.DB.Queries().GetOrganizationByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Organization not found").JSON(c)
	}

	err = h.server.DB.Queries().SoftDeleteOrganization(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMessageResponse(http.StatusOK, "Organization successfully deleted").JSON(c)
}
