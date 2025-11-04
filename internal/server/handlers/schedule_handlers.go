package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"

	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
)

// ScheduleHandler handles schedule-related requests
type ScheduleHandler struct {
	server *s.Server
}

// NewScheduleHandler creates a new schedule handler
func NewScheduleHandler(server *s.Server) *ScheduleHandler {
	return &ScheduleHandler{
		server: server,
	}
}

// @Summary Retrieve schedule
// @Description Retrieves schedule appointments with optional filters
// @Tags Schedule
// @Accept json
// @Produce json
// @Param locationCode query string true "Location code"
// @Param startDate query string true "Start date (YYYY-MM-DD)"
// @Param sessionId query string true "Session ID (GUID)"
// @Success 200 {object} responses.ScheduleResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/schedule/retrieve [get]
func (h *ScheduleHandler) RetrieveSchedule(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ScheduleRetrieveRequest)
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

	dmeResponse, err := h.server.DME.RetrieveSchedule(ctx, req.LocationCode, req.StartDate, req.SessionID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve schedule",
			zap.String("locationCode", req.LocationCode),
			zap.String("startDate", req.StartDate),
			zap.String("sessionId", req.SessionID),
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve schedule: "+err.Error()).JSON(c)
	}

	response := responses.ConvertScheduleResponse(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve schedule for manager
// @Description Retrieves schedule appointments for a specific manager
// @Tags Schedule
// @Accept json
// @Produce json
// @Param locationCode query string true "Location code"
// @Param startDate query string true "Start date (YYYY-MM-DD)"
// @Param managerId query string true "Manager ID"
// @Param sessionId query string true "Session ID (GUID)"
// @Success 200 {object} responses.ScheduleResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/schedule/retrieve-for-manager [get]
func (h *ScheduleHandler) RetrieveScheduleForManager(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ScheduleRetrieveForManagerRequest)
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

	dmeResponse, err := h.server.DME.RetrieveScheduleForManager(ctx, req.LocationCode, req.StartDate, req.ManagerID, req.SessionID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve schedule for manager",
			zap.String("locationCode", req.LocationCode),
			zap.String("startDate", req.StartDate),
			zap.String("managerId", req.ManagerID),
			zap.String("sessionId", req.SessionID),
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve schedule for manager: "+err.Error()).JSON(c)
	}

	response := responses.ConvertScheduleResponse(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve schedule for technician
// @Description Retrieves schedule appointments for a specific technician
// @Tags Schedule
// @Accept json
// @Produce json
// @Param locationCode query string true "Location code"
// @Param startDate query string true "Start date (YYYY-MM-DD)"
// @Param techId query string true "Technician ID"
// @Param sessionId query string true "Session ID (GUID)"
// @Success 200 {object} responses.ScheduleResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/schedule/retrieve-for-tech [get]
func (h *ScheduleHandler) RetrieveScheduleForTech(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ScheduleRetrieveForTechRequest)
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

	dmeResponse, err := h.server.DME.RetrieveScheduleForTech(ctx, req.LocationCode, req.StartDate, req.TechID, req.SessionID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve schedule for tech",
			zap.String("locationCode", req.LocationCode),
			zap.String("startDate", req.StartDate),
			zap.String("techId", req.TechID),
			zap.String("sessionId", req.SessionID),
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve schedule for tech: "+err.Error()).JSON(c)
	}

	response := responses.ConvertScheduleResponse(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve schedule for work order
// @Description Retrieves schedule appointments for a specific work order
// @Tags Schedule
// @Accept json
// @Produce json
// @Param locationCode query string true "Location code"
// @Param startDate query string true "Start date (YYYY-MM-DD)"
// @Param workOrderId query string true "Work order ID"
// @Param sessionId query string true "Session ID (GUID)"
// @Success 200 {object} responses.ScheduleResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/schedule/retrieve-for-work-order [get]
func (h *ScheduleHandler) RetrieveScheduleForWorkOrder(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ScheduleRetrieveForWorkOrderRequest)
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

	dmeResponse, err := h.server.DME.RetrieveScheduleForWorkOrder(ctx, req.LocationCode, req.StartDate, req.WorkOrderID, req.SessionID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve schedule for work order",
			zap.String("locationCode", req.LocationCode),
			zap.String("startDate", req.StartDate),
			zap.String("workOrderId", req.WorkOrderID),
			zap.String("sessionId", req.SessionID),
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve schedule for work order: "+err.Error()).JSON(c)
	}

	response := responses.ConvertScheduleResponse(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve work order schedule
// @Description Retrieves work order schedule
// @Tags Schedule
// @Accept json
// @Produce json
// @Param locationCode query string true "Location code"
// @Param workOrderId query string true "Work order ID"
// @Param sessionId query string true "Session ID (GUID)"
// @Success 200 {object} responses.ScheduleResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/schedule/work-order-schedule [get]
func (h *ScheduleHandler) RetrieveWorkOrderSchedule(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ScheduleWorkOrderScheduleRequest)
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

	dmeResponse, err := h.server.DME.RetrieveWorkOrderSchedule(ctx, req.LocationCode, req.WorkOrderID, req.SessionID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve work order schedule",
			zap.String("locationCode", req.LocationCode),
			zap.String("workOrderId", req.WorkOrderID),
			zap.String("sessionId", req.SessionID),
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve work order schedule: "+err.Error()).JSON(c)
	}

	response := responses.ConvertScheduleResponse(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Retrieve operation schedule
// @Description Retrieves operation schedule
// @Tags Schedule
// @Accept json
// @Produce json
// @Param locationCode query string true "Location code"
// @Param workOrderId query string true "Work order ID"
// @Param opcode query string true "Operation code"
// @Param sessionId query string true "Session ID (GUID)"
// @Success 200 {object} responses.ScheduleResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/schedule/operation-schedule [get]
func (h *ScheduleHandler) RetrieveOperationSchedule(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ScheduleOperationScheduleRequest)
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

	dmeResponse, err := h.server.DME.RetrieveOperationSchedule(ctx, req.LocationCode, req.WorkOrderID, req.Opcode, req.SessionID, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to retrieve operation schedule",
			zap.String("locationCode", req.LocationCode),
			zap.String("workOrderId", req.WorkOrderID),
			zap.String("opcode", req.Opcode),
			zap.String("sessionId", req.SessionID),
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve operation schedule: "+err.Error()).JSON(c)
	}

	response := responses.ConvertScheduleResponse(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Update schedule
// @Description Updates schedule appointments
// @Tags Schedule
// @Accept json
// @Produce json
// @Param body body requests.ScheduleUpdateRequest true "Schedule update request"
// @Success 200 {object} responses.ScheduleUpdateResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/schedule/update [post]
func (h *ScheduleHandler) UpdateSchedule(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ScheduleUpdateRequest)
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

	// Convert appointments to []interface{} to ensure proper JSON marshaling
	appointmentsInterface := make([]interface{}, len(req.Appointments))
	for i, appt := range req.Appointments {
		appointmentsInterface[i] = appt
	}

	// Convert request to map for DME API
	// DME API expects the payload wrapped in a "scheduleUpdate" field
	payload := map[string]interface{}{
		"scheduleUpdate": map[string]interface{}{
			"locationCode": req.LocationCode,
			"clerkId":      req.ClerkID,
			"sessionId":    req.SessionID,
			"appointments": appointmentsInterface,
		},
	}

	// Debug: Log the payload being sent
	payloadJSON, err := json.Marshal(payload)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to marshal payload for logging", zap.Error(err))
	}
	h.server.Logger.DesugarZap.Info("==================== SCHEDULE UPDATE PAYLOAD ====================")
	h.server.Logger.DesugarZap.Info("Sending schedule update to DME",
		zap.String("payload", string(payloadJSON)),
		zap.String("clerkId", req.ClerkID),
		zap.String("locationCode", req.LocationCode),
		zap.Int("appointments_count", len(req.Appointments)))
	h.server.Logger.DesugarZap.Info("===============================================================")

	dmeResponse, err := h.server.DME.UpdateSchedule(ctx, payload, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to update schedule",
			zap.String("locationCode", req.LocationCode),
			zap.String("clerkId", req.ClerkID),
			zap.String("sessionId", req.SessionID),
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to update schedule: "+err.Error()).JSON(c)
	}

	response := responses.ConvertScheduleUpdateResponse(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

// @Summary Resolve merge conflict
// @Description Resolves merge conflicts in schedule updates
// @Tags Schedule
// @Accept json
// @Produce json
// @Param body body requests.ScheduleResolveMergeConflictRequest true "Merge conflict resolution request"
// @Success 200 {object} responses.ScheduleUpdateResponse
// @Failure 400 {object} responses.Error
// @Failure 500 {object} responses.Error
// @Router /service/schedule/resolve-merge-conflict [post]
func (h *ScheduleHandler) ResolveMergeConflict(c echo.Context) error {
	ctx := c.Request().Context()
	req := new(requests.ScheduleResolveMergeConflictRequest)
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

	// Convert request to map for DME API
	// DME API expects the payload wrapped in a field
	payload := map[string]interface{}{
		"scheduleUpdate": map[string]interface{}{
			"locationCode": req.LocationCode,
			"clerkId":      req.ClerkID,
			"sessionId":    req.SessionID,
			"appointments": req.Appointments,
		},
	}

	dmeResponse, err := h.server.DME.ResolveMergeConflict(ctx, payload, orgID, *systemID)
	if err != nil {
		h.server.Logger.DesugarZap.Error("Failed to resolve merge conflict",
			zap.String("locationCode", req.LocationCode),
			zap.String("clerkId", req.ClerkID),
			zap.String("sessionId", req.SessionID),
			zap.Error(err))
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to resolve merge conflict: "+err.Error()).JSON(c)
	}

	response := responses.ConvertScheduleUpdateResponse(dmeResponse)
	return c.JSON(http.StatusOK, response)
}

