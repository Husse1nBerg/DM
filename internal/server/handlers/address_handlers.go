package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
)

// AddressHandler handles address-related operations
type AddressHandler struct {
	server *s.Server
}

// NewAddressHandler creates a new address handler
func NewAddressHandler(server *s.Server) *AddressHandler {
	return &AddressHandler{server: server}
}

// GetAddressById retrieves an address by its ID
//
//	@Summary		Get address by ID
//	@Description	Retrieves an address by its ID
//	@Tags			Addresses
//	@Accept			json
//	@Produce		json
//	@Param			id	path		string	true	"Address ID"	Format(uuid)
//	@Success		200	{object}	responses.AddressResponse
//	@Failure		400	{object}	responses.BaseResponse
//	@Failure		404	{object}	responses.BaseResponse
//	@Failure		500	{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/addresses/{id} [get]
func (h *AddressHandler) GetAddressById(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid address ID").JSON(c)
	}

	address, err := h.server.DB.Queries().GetAddressByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Address not found").JSON(c)
	}

	return responses.NewAddressResponseSuccess(address).JSON(c)
}

// CreateAddress creates a new address
//
//	@Summary		Create address
//	@Description	Creates a new address in the system
//	@Tags			Addresses
//	@Accept			json
//	@Produce		json
//	@Param			address	body		requests.CreateAddressRequest	true	"Address details"
//	@Success		201		{object}	responses.AddressResponse
//	@Failure		400		{object}	responses.BaseResponse
//	@Failure		500		{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/addresses [post]
func (h *AddressHandler) CreateAddress(c echo.Context) error {
	var req requests.CreateAddressRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Convert numeric values
	var lat, lng float64
	if req.Latitude != nil {
		lat = *req.Latitude
	}
	if req.Longitude != nil {
		lng = *req.Longitude
	}

	// Create the address
	params := db.CreateAddressParams{
		Street:     req.Street,
		City:       req.City,
		State:      req.State,
		PostalCode: req.PostalCode,
		Country:    req.Country,
		Latitude:   lat,
		Longitude:  lng,
	}

	address, err := h.server.DB.Queries().CreateAddress(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.NewAddressResponseSuccess(address)
	response.Code = http.StatusCreated
	return response.JSON(c)
}

// UpdateAddress updates an existing address
//
//	@Summary		Update address
//	@Description	Updates an existing address
//	@Tags			Addresses
//	@Accept			json
//	@Produce		json
//	@Param			id			path		string							true	"Address ID"	Format(uuid)
//	@Param			address		body		requests.UpdateAddressRequest	true	"Address details"
//	@Success		200			{object}	responses.AddressResponse
//	@Failure		400			{object}	responses.BaseResponse
//	@Failure		404			{object}	responses.BaseResponse
//	@Failure		500			{object}	responses.BaseResponse
//	@Security		ApiKeyAuth
//	@Router			/addresses/{id} [put]
func (h *AddressHandler) UpdateAddress(c echo.Context) error {
	idStr := c.Param("id")
	id, err := uuid.Parse(idStr)
	if err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid address ID").JSON(c)
	}

	// Get current address
	currentAddress, err := h.server.DB.Queries().GetAddressByID(c.Request().Context(), id)
	if err != nil {
		return responses.NewErrorResponse(http.StatusNotFound, "Address not found").JSON(c)
	}

	var req requests.UpdateAddressRequest
	if err := c.Bind(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(&req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Build update params with current values
	street := currentAddress.Street
	city := currentAddress.City
	state := currentAddress.State
	postalCode := currentAddress.PostalCode
	country := currentAddress.Country
	lat := currentAddress.Latitude
	lng := currentAddress.Longitude

	// Update fields that are provided
	if req.Street != nil {
		street = req.Street
	}
	if req.City != nil {
		city = req.City
	}
	if req.State != nil {
		state = req.State
	}
	if req.PostalCode != nil {
		postalCode = req.PostalCode
	}
	if req.Country != nil {
		country = req.Country
	}
	if req.Latitude != nil {
		lat = *req.Latitude
	}
	if req.Longitude != nil {
		lng = *req.Longitude
	}

	params := db.UpdateAddressParams{
		ID:         id,
		Street:     street,
		City:       city,
		State:      state,
		PostalCode: postalCode,
		Country:    country,
		Latitude:   lat,
		Longitude:  lng,
	}

	updatedAddress, err := h.server.DB.Queries().UpdateAddress(c.Request().Context(), params)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	return responses.NewAddressResponseSuccess(updatedAddress).JSON(c)
}
