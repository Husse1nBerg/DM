package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/models"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// MarinaHandler handles marina-related operations
type MarinaHandler struct {
	server *s.Server
}

// NewMarinaHandler creates a new marina handler
func NewMarinaHandler(server *s.Server) *MarinaHandler {
	return &MarinaHandler{server: server}
}

// CreateMarina creates a new marina
//
//	@Summary		Create marina
//	@Description	Creates a new marina in the system
//	@Tags			Marinas
//	@Accept			json
//	@Produce		json
//	@Param			marina	body		requests.CreateMarinaRequest	true	"Marina details"
//	@Success		201		{object}	responses.MarinaResponseWrapper
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/marinas [post]
func (h *MarinaHandler) CreateMarina(c echo.Context) error {
	var req requests.CreateMarinaRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Set default values
	isActive := true
	isTest := false
	maxUsers := int32(100)

	if req.IsActive != nil {
		isActive = *req.IsActive
	}
	if req.IsTest != nil {
		isTest = *req.IsTest
	}
	if req.MaxUsers != nil {
		maxUsers = *req.MaxUsers
	}

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

	// Convert WorkingHours to []byte for database storage
	var workingHoursBytes []byte
	var err error
	if req.WorkingHours != nil {
		workingHoursBytes, err = req.WorkingHours.ToBytes()
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid working hours format").JSON(c)
		}
	} else {
		// Use default working hours
		defaultHours := models.DefaultWorkingHours()
		workingHoursBytes, err = defaultHours.ToBytes()
		if err != nil {
			return responses.NewErrorResponse(http.StatusInternalServerError, "Error creating default working hours").JSON(c)
		}
	}

	params := db.CreateMarinaParams{
		OrganizationID: req.OrganizationID,
		Name:           req.Name,
		Email:          req.Email,
		Location:       req.Location,
		Phone:          req.Phone,
		Country:        req.Country,
		Currency:       req.Currency,
		WorkingHours:   workingHoursBytes,
		Website:        req.Website,
		Image:          req.Image,
		MaxUsers:       &maxUsers,
		IsActive:       &isActive,
		IsTest:         &isTest,
		AddressID:      addressID,
	}

	marina, err := h.server.DB.Queries().CreateMarina(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.NewSuccessResponse(responses.ConvertMarinaToResponse(marina))
	response.Code = http.StatusCreated
	return response.JSON(c)
}

// GetMarinaByID retrieves a marina by its ID
//
//	@Summary		Get marina by ID
//	@Description	Retrieves a marina by its ID
//	@Tags			Marinas
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Marina ID"	Format(uuid)
//	@Success		200	{object}	responses.MarinaResponseWrapper
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/marinas/{id} [get]
func (h *MarinaHandler) GetMarinaByID(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID").JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(c)
	}

	return responses.NewMarinaResponseSuccess(marina).JSON(c)
}

// GetMarinaWithAddress retrieves a marina and its address
//
//	@Summary		Get marina with address
//	@Description	Retrieves a marina and its address details
//	@Tags			Marinas
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Marina ID"	Format(uuid)
//	@Success		200	{object}	responses.MarinaWithAddressResponseWrapper
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/marinas/{id}/with-address [get]
func (h *MarinaHandler) GetMarinaWithAddress(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID").JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(c)
	}

	// Get the address
	address, err := h.server.DB.Queries().GetAddressByID(c.Request().Context(), marina.AddressID)
	if err != nil {
		// Return marina without address if address not found
		return responses.NewMarinaWithAddressResponse(marina, nil).JSON(c)
	}

	return responses.NewMarinaWithAddressResponse(marina, &address).JSON(c)
}

// GetMarinaByEmail retrieves a marina by its email
//
//	@Summary		Get marina by email
//	@Description	Retrieves a marina by its email address
//	@Tags			Marinas
//	@Accept			json
//	@Produce		json
//	@Param			email	query		string	true	"Marina email"
//	@Success		200		{object}	responses.MarinaResponseWrapper
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		404		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/marinas/by-email [get]
func (h *MarinaHandler) GetMarinaByEmail(c echo.Context) error {
	email := c.QueryParam("email")
	if email == "" {
		return responses.NewErrorResponse(http.StatusBadRequest, "Email parameter is required").JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByEmail(c.Request().Context(), email)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(c)
	}

	return responses.NewMarinaResponseSuccess(marina).JSON(c)
}

// GetMarinasPaginated retrieves marinas with pagination
//
//	@Summary		Get paginated marinas
//	@Description	Retrieves marinas with pagination support
//	@Tags			Marinas
//	@Accept			json
//	@Produce		json
//	@Param			page		query		int		false	"Page number"	default(1)
//	@Param			pageSize	query		int		false	"Page size"		default(10)
//	@Success		200		{array}		responses.MarinaListResponse
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/marinas [get]
func (h *MarinaHandler) GetMarinasPaginated(c echo.Context) error {
	var req requests.PaginationQuery

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

	// Calculate total count
	allMarinas, err := h.server.DB.Queries().GetAllMarinas(c.Request().Context())
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allMarinas))

	// Fetch paginated data
	params := db.GetMarinasPaginatedParams{
		Limit:  req.PageSize,
		Offset: (req.Page - 1) * req.PageSize,
	}

	marinas, err := h.server.DB.Queries().GetMarinasPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Calculate current page
	currentPage := req.Page

	return responses.NewMarinasPaginatedResponse(marinas, total, req.PageSize, currentPage).JSON(c)
}

// GetMarinasByOrganization retrieves marinas for a specific organization with pagination
//
//	@Summary		Get marinas by organization
//	@Description	Retrieves marinas belonging to a specific organization with pagination support
//	@Tags			Marinas
//	@Accept			json
//	@Produce		json
//	@Param			organizationId	path		string	true	"Organization ID"	Format(uuid)
//	@Param			page				query		int		false	"Page number"	default(1)
//	@Param			pageSize			query		int		false	"Page size"		default(10)
//	@Success		200				{array}		responses.MarinaListResponse
//	@Failure		400				{object}	responses.BaseResponse
//	@Failure		404				{object}	responses.BaseResponse
//	@Failure		500				{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/marinas/organization/{organizationId} [get]
func (h *MarinaHandler) GetMarinasByOrganization(c echo.Context) error {
	orgIDStr := c.Param("organizationId")
	orgID, err := uuid.Parse(orgIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid organization ID").JSON(c)
	}

	var req requests.PaginationQuery
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

	// Get all marinas for this organization to calculate total
	allOrgMarinas, err := h.server.DB.Queries().GetMarinasByOrganization(c.Request().Context(), orgID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allOrgMarinas))

	if total == 0 {
		// Return empty response if no marinas found
		return responses.NewMarinasPaginatedResponse([]db.Marina{}, 0, req.PageSize, 1).JSON(c)
	}

	// Fetch paginated data for this organization
	params := db.GetMarinasByOrganizationPaginatedParams{
		OrganizationID: orgID,
		Limit:          req.PageSize,
		Offset:         (req.Page - 1) * req.PageSize,
	}

	marinas, err := h.server.DB.Queries().GetMarinasByOrganizationPaginated(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	// Calculate current page
	currentPage := req.Page

	return responses.NewMarinasPaginatedResponse(marinas, total, req.PageSize, currentPage).JSON(c)
}

// UpdateMarina updates an existing marina
//
//	@Summary		Update marina
//	@Description	Updates an existing marina
//	@Tags			Marinas
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Marina ID"	Format(uuid)
//	@Param			marina	body		requests.UpdateMarinaRequest	true	"Marina details"
//	@Success		200		{object}	responses.MarinaResponseWrapper
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		404		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/marinas/{id} [put]
func (h *MarinaHandler) UpdateMarina(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID").JSON(c)
	}

	// Check if marina exists and get current values
	currentMarina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(c)
	}

	var req requests.UpdateMarinaRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Build update params with current values that will be overridden
	params := db.UpdateMarinaParams{
		ID:           id,
		Name:         currentMarina.Name,
		Email:        currentMarina.Email,
		Location:     currentMarina.Location,
		Phone:        currentMarina.Phone,
		Country:      currentMarina.Country,
		Currency:     currentMarina.Currency,
		WorkingHours: currentMarina.WorkingHours,
		Website:      currentMarina.Website,
		Image:        currentMarina.Image,
		MaxUsers:     currentMarina.MaxUsers,
		IsActive:     currentMarina.IsActive,
		IsTest:       currentMarina.IsTest,
		AddressID:    currentMarina.AddressID,
		SystemID:     currentMarina.SystemID,
	}

	// Update only fields that are provided
	if req.Name != nil {
		params.Name = *req.Name
	}
	if req.Email != nil {
		params.Email = *req.Email
	}
	if req.Location != nil {
		params.Location = req.Location
	}
	if req.Phone != nil {
		params.Phone = req.Phone
	}
	if req.Country != nil {
		params.Country = req.Country
	}
	if req.Currency != nil {
		params.Currency = req.Currency
	}
	if req.WorkingHours != nil {
		// Convert WorkingHours struct to []byte for database storage
		workingHoursBytes, err := req.WorkingHours.ToBytes()
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid working hours format").JSON(c)
		}
		params.WorkingHours = workingHoursBytes
	}
	if req.Website != nil {
		params.Website = req.Website
	}
	if req.Image != nil {
		params.Image = req.Image
	}
	if req.MaxUsers != nil {
		params.MaxUsers = req.MaxUsers
	}
	if req.IsActive != nil {
		params.IsActive = req.IsActive
	}
	if req.IsTest != nil {
		params.IsTest = req.IsTest
	}
	if req.SystemID != nil {
		params.SystemID = req.SystemID
	}

	updatedMarina, err := h.server.DB.Queries().UpdateMarina(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMarinaResponseSuccess(updatedMarina).JSON(c)
}

// UpdateMarinaWithAddress updates a marina and its address
//
//	@Summary		Update marina with address
//	@Description	Updates a marina and its address details
//	@Tags			Marinas
//	@Accept			json
//	@Produce		json
//	@Param			id		path		string						true	"Marina ID"	Format(uuid)
//	@Param			marina	body		requests.UpdateMarinaRequest	true	"Marina details with address"
//	@Success		200		{object}	responses.MarinaWithAddressResponseWrapper
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		404		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/marinas/{id}/with-address [put]
func (h *MarinaHandler) UpdateMarinaWithAddress(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID").JSON(c)
	}

	// Check if marina exists and get current values
	currentMarina, err := h.server.DB.Queries().GetMarinaByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(c)
	}

	var req requests.UpdateMarinaRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Build update params with current values that will be overridden
	params := db.UpdateMarinaParams{
		ID:           id,
		Name:         currentMarina.Name,
		Email:        currentMarina.Email,
		Location:     currentMarina.Location,
		Phone:        currentMarina.Phone,
		Country:      currentMarina.Country,
		Currency:     currentMarina.Currency,
		WorkingHours: currentMarina.WorkingHours,
		Website:      currentMarina.Website,
		Image:        currentMarina.Image,
		MaxUsers:     currentMarina.MaxUsers,
		IsActive:     currentMarina.IsActive,
		IsTest:       currentMarina.IsTest,
		AddressID:    currentMarina.AddressID,
		SystemID:     currentMarina.SystemID,
	}

	// Update only fields that are provided
	if req.Name != nil {
		params.Name = *req.Name
	}
	if req.Email != nil {
		params.Email = *req.Email
	}
	if req.Location != nil {
		params.Location = req.Location
	}
	if req.Phone != nil {
		params.Phone = req.Phone
	}
	if req.Country != nil {
		params.Country = req.Country
	}
	if req.Currency != nil {
		params.Currency = req.Currency
	}
	if req.WorkingHours != nil {
		// Convert WorkingHours struct to []byte for database storage
		workingHoursBytes, err := req.WorkingHours.ToBytes()
		if err != nil {
			return responses.NewErrorResponse(http.StatusBadRequest, "Invalid working hours format").JSON(c)
		}
		params.WorkingHours = workingHoursBytes
	}
	if req.Website != nil {
		params.Website = req.Website
	}
	if req.Image != nil {
		params.Image = req.Image
	}
	if req.MaxUsers != nil {
		params.MaxUsers = req.MaxUsers
	}
	if req.IsActive != nil {
		params.IsActive = req.IsActive
	}
	if req.IsTest != nil {
		params.IsTest = req.IsTest
	}
	if req.SystemID != nil {
		params.SystemID = req.SystemID
	}

	updatedMarina, err := h.server.DB.Queries().UpdateMarina(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Error updating marina").JSON(c)
	}

	// Update address if provided
	var updatedAddress *db.Address
	if req.Address != nil {
		// Get current address
		currentAddress, err := h.server.DB.Queries().GetAddressByID(c.Request().Context(), currentMarina.AddressID)
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
				ID:         currentMarina.AddressID,
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

	return responses.NewMarinaWithAddressResponse(updatedMarina, updatedAddress).JSON(c)
}

// DeleteMarina deletes a marina (soft delete)
//
//	@Summary		Delete marina
//	@Description	Soft deletes a marina
//	@Tags			Marinas
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Marina ID"	Format(uuid)
//	@Success		200	{object}	responses.BaseResponse
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/marinas/{id} [delete]
func (h *MarinaHandler) DeleteMarina(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid marina ID").JSON(c)
	}

	// Check if marina exists
	_, err = h.server.DB.Queries().GetMarinaByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Marina not found").JSON(c)
	}

	err = h.server.DB.Queries().SoftDeleteMarina(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewMessageResponse(http.StatusOK, "Marina successfully deleted").JSON(c)
}

// GetUserMarinas retrieves marinas associated with a user with pagination
//
//	@Summary		Get user marinas
//	@Description	Retrieves marinas associated with a specific user with pagination support
//	@Tags			Marinas
//	@Accept			json
//	@Produce		json
//	@Param			userId		path		string	true	"User ID"	Format(uuid)
//	@Success		200			{array}		responses.MarinaListResponse
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/marinas/user/{userId} [get]
func (h *MarinaHandler) GetUserMarinas(c echo.Context) error {
	userIDStr := c.Param("userId")
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid user ID").JSON(c)
	}

	// Get all marinas for this user to calculate total
	allUserMarinas, err := h.server.DB.Queries().GetUserMarinasList(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allUserMarinas))

	if total == 0 {
		// Return empty response if no marinas found
		return responses.NewMarinasPaginatedResponse([]db.Marina{}, 0, int32(total), 1).JSON(c)
	}

	return responses.NewMarinasPaginatedResponse(allUserMarinas, total, int32(total), 1).JSON(c)
}

// GetMyUserMarinas retrieves marinas associated with the current user
//
//	@Summary		Get my user marinas
//	@Description	Retrieves marinas associated with the current authenticated user
//	@Tags			Marinas
//	@Accept			json
//	@Produce		json
//	@Success		200			{array}		responses.MarinaListResponse
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/marinas/user [get]
func (h *MarinaHandler) GetMyUserMarinas(c echo.Context) error {
	// Get user ID from the token
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID

	// Get all marinas for this user to calculate total
	allUserMarinas, err := h.server.DB.Queries().GetUserMarinasList(c.Request().Context(), userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}
	total := int64(len(allUserMarinas))

	if total == 0 {
		// Return empty response if no marinas found
		return responses.NewMarinasPaginatedResponse([]db.Marina{}, 0, int32(total), 1).JSON(c)
	}

	return responses.NewMarinasPaginatedResponse(allUserMarinas, total, int32(total), 1).JSON(c)
}
