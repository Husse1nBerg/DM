package handlers

import (
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

// VendorHandler handles operations related to DME Vendor API endpoints
type VendorHandler struct {
	server *s.Server
}

// NewVendorHandler creates a new vendor handler instance
func NewVendorHandler(server *s.Server) *VendorHandler {
	return &VendorHandler{server: server}
}

// ListVendorsHandler godoc
// @Summary List all vendors
// @Description Retrieves a list of all vendors from the DME API
// @Tags Vendors
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Success 200 {object} responses.VendorListResponse
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /api/v1/vendors/list [get]
func (h *VendorHandler) ListVendorsHandler(c echo.Context) error {
	// Get user from token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	req := new(requests.VendorListRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Call DME API to list vendors
	vendors, err := h.server.DME.ListVendors(c.Request().Context(), claims.OrgId, req.SystemID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to list vendors from DME API", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve vendors").JSON(c)
	}

	// Convert to response format
	response := responses.NewVendorListResponse(vendors)

	return responses.NewSuccessResponse(response).JSON(c)
}

// SearchVendorsHandler godoc
// @Summary Search vendors
// @Description Search for vendors based on search criteria
// @Tags Vendors
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Param searchString query string true "Search string (vendor name, phone, email, etc.)"
// @Param directHit query bool false "Only return result if exactly one match is found"
// @Success 200 {object} responses.VendorSearchResponse
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /api/v1/vendors/search [get]
func (h *VendorHandler) SearchVendorsHandler(c echo.Context) error {
	// Get user from token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	req := new(requests.VendorSearchRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Default directHit to false if not provided
	directHit := false
	if req.DirectHit != nil {
		directHit = *req.DirectHit
	}

	// Call DME API to search vendors
	vendors, err := h.server.DME.SearchVendors(c.Request().Context(), req.SearchString, directHit, claims.OrgId, req.SystemID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to search vendors from DME API", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to search vendors").JSON(c)
	}

	// Convert to response format
	response := responses.NewVendorSearchResponse(vendors)

	return responses.NewSuccessResponse(response).JSON(c)
}

// RetrieveVendorHandler godoc
// @Summary Retrieve a vendor
// @Description Retrieves a specific vendor by ID from the DME API
// @Tags Vendors
// @Accept json
// @Produce json
// @Param systemId query string true "System ID"
// @Param vendorId query string true "Vendor ID"
// @Success 200 {object} responses.VendorResponse
// @Failure 400 {object} responses.Error
// @Failure 401 {object} responses.Error
// @Failure 404 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /api/v1/vendors [get]
func (h *VendorHandler) RetrieveVendorHandler(c echo.Context) error {
	// Get user from token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*token.JwtCustomClaims)

	req := new(requests.VendorRetrieveRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Call DME API to retrieve vendor
	vendor, err := h.server.DME.RetrieveVendor(c.Request().Context(), req.VendorID, claims.OrgId, req.SystemID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to retrieve vendor from DME API", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve vendor").JSON(c)
	}

	// Convert to response format
	response := responses.NewVendorResponse(*vendor)

	return responses.NewSuccessResponse(response).JSON(c)
}

