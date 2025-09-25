package handlers

import (
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
)

// ServiceHandler handles general service-related requests
type ServiceHandler struct {
	server *s.Server
}

// NewServiceHandler creates a new service handler
func NewServiceHandler(server *s.Server) *ServiceHandler {
	return &ServiceHandler{
		server: server,
	}
}

// @Summary List new or changed operation codes
// @Description Searches for operation codes created or changed as of a date
// @Tags Service
// @Accept json
// @Produce json
// @Param AsOfDate query string true "As of date"
// @Param page query int true "Page number" minimum(1)
// @Param pageSize query int true "Page size" minimum(1) maximum(100)
// @Success 200 {object} responses.ServiceOpCodesResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/opcodes/new-or-changed [get]
func (h *ServiceHandler) ListNewOrChangedOpCodes(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ServiceListNewOrChangedOpCodesRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	dmeResponse, err := h.server.DME.ListNewOrChangedOpCodes(ctx, req.AsOfDate, req.Page, req.PageSize, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list new or changed operation codes",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertServiceOpCodes(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary List work order category codes
// @Description Returns a list of work order category codes for the page and page size specified
// @Tags Service
// @Accept json
// @Produce json
// @Param page query int true "Page number" minimum(1)
// @Param pageSize query int true "Page size" minimum(1) maximum(100)
// @Success 200 {object} responses.ServiceWOCategoryCodesResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/work-order-category-codes [get]
func (h *ServiceHandler) ListWOCategoryCodes(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ServiceListWOCategoryCodesRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	dmeResponse, err := h.server.DME.ListWOCategoryCodes(ctx, req.Page, req.PageSize, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list work order category codes",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertServiceWOCategoryCodes(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary List operation category codes
// @Description Returns a list of operation category codes for the page and page size specified
// @Tags Service
// @Accept json
// @Produce json
// @Param page query int true "Page number" minimum(1)
// @Param pageSize query int true "Page size" minimum(1) maximum(100)
// @Success 200 {object} responses.ServiceOPCategoryCodesResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/operation-category-codes [get]
func (h *ServiceHandler) ListOPCategoryCodes(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ServiceListOPCategoryCodesRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	dmeResponse, err := h.server.DME.ListOPCategoryCodes(ctx, req.Page, req.PageSize, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list operation category codes",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertServiceOPCategoryCodes(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve operation descriptions
// @Description Retrieve the operation descriptions from the opcode template
// @Tags Service
// @Accept json
// @Produce json
// @Success 200 {object} responses.ServiceOperationDescriptionsResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/operation-descriptions [get]
func (h *ServiceHandler) RetrieveOperationDescriptions(c echo.Context) error {
	ctx := c.Request().Context()

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	dmeResponse, err := h.server.DME.RetrieveOperationDescriptions(ctx, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve operation descriptions",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertServiceOperationDescriptions(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary List technicians
// @Description Retrieves a list of technician records
// @Tags Service
// @Accept json
// @Produce json
// @Success 200 {object} responses.ServiceTechniciansResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/technicians [get]
func (h *ServiceHandler) ListTechnicians(c echo.Context) error {
	ctx := c.Request().Context()

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	userID := claims.ID
	user, err := h.server.DB.Queries().GetUserByID(ctx, userID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get user: "+err.Error()).JSON(c)
	}

	marina, err := h.server.DB.Queries().GetMarinaByID(ctx, user.MarinaID)
	if err != nil {
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to get marina: "+err.Error()).JSON(c)
	}

	orgID := marina.OrganizationID
	systemID := marina.SystemID

	if systemID == nil {
		return responses.NewErrorResponse(http.StatusBadRequest, "System ID is required for DME operations").JSON(c)
	}

	dmeResponse, err := h.server.DME.ListTechnicians(ctx, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to list technicians",
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, err).JSON(c)
	}

	response := responses.ConvertServiceTechnicians(dmeResponse)
	return c.JSON(http.StatusOK, response)
}
