package handlers

import (
	"bytes"
	"fmt"
	"io"
	"net/http"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
)

type NotificationPreferenceHandler struct {
	server *s.Server
}

func NewNotificationPreferenceHandler(server *s.Server) *NotificationPreferenceHandler {
	return &NotificationPreferenceHandler{server: server}
}

// ListNotificationPreferencesHandler lists all existing preferences
//
//	@Summary		List notification preferences
//	@Description	Get all notification preferences with pagination
//	@Tags			Notification Preference
//	@Accept			json
//	@Produce		json
//	@Success		200	{array}	responses.NotificationPreferenceResponse "List of notification preferences"
//	@Failure		500	{object}	responses.Error "Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/notification-preference/list [get]
func (h *NotificationPreferenceHandler) ListNotificationPreferencesHandler(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	id := claims.ID

	queries := h.server.DB.Queries()

	preferences, err := queries.GetNotificationPreferences(c.Request().Context(), id)
	if err != nil {
		h.server.Logger.Zap.Errorw("Failed to retrieve notification preferences", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to retrieve notification preferences").JSON(c)
	}

	response := responses.NotificationPreferenceDBToResponseList(preferences)
	return response.JSON(c)
}

// UpdateNotificationPreferenceHandler creates or updates a user's notification preference
//
//	@Summary		Upsert notification preference
//	@Description	Create or update a notification preference for the authenticated user
//	@Tags			Notification Preference
//	@Accept			json
//	@Produce		json
//	@Param			body	body		requests.NotificationPreferenceRequest	true	"Notification preference payload"
//	@Success		200		{object}	responses.NotificationPreferenceResponse	"Updated notification preference"
//	@Failure		400		{object}	responses.Error	"Bad request"
//	@Failure		500		{object}	responses.Error	"Server error"
//	@Security		ApiKeyAuth
//
//	@Router			/notification-preference [put]
func (h *NotificationPreferenceHandler) UpdateNotificationPreferenceHandler(c echo.Context) error {
	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	id := claims.ID

	// Bind and validate request body
	req := new(requests.NotificationPreferenceRequest)
	if err := c.Bind(req); err != nil {
		h.server.Logger.Zap.Errorw("Failed to bind request", "error", err.Error())
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}
	if err := req.Validate(); err != nil {
		h.server.Logger.Zap.Errorw("Validation failed", "error", err.Error())
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	queries := h.server.DB.Queries()

	// Log the request details for debugging
	h.server.Logger.Zap.Infow("Updating notification preference",
		"userID", id,
		"notificationType", req.NotificationType,
		"enabled", req.Enabled,
		"deliveryMethod", req.DeliveryMethod,
	)

	// Check if preference already exists
	existingPref, err := queries.GetNotificationPreference(c.Request().Context(), db.GetNotificationPreferenceParams{
		UserID:           id,
		NotificationType: req.NotificationType,
	})
	if err != nil {
		// Preference doesn't exist, will be created by upsert
		h.server.Logger.Zap.Infow("Preference does not exist, will create new one")
	} else {
		h.server.Logger.Zap.Infow("Existing preference found",
			"id", existingPref.ID,
			"enabled", existingPref.Enabled,
			"deliveryMethod", existingPref.DeliveryMethod,
		)
	}

	// Use upsert to create or update seamlessly
	params := db.UpsertNotificationPreferenceParams{
		UserID:           id,
		NotificationType: req.NotificationType,
		Enabled:          req.Enabled,
		DeliveryMethod:   &req.DeliveryMethod,
	}

	preference, err := queries.UpsertNotificationPreference(c.Request().Context(), params)
	if err != nil {
		h.server.Logger.Zap.Errorw("Failed to upsert notification preference", "error", err.Error())
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to update notification preference: "+err.Error()).JSON(c)
	}

	h.server.Logger.Zap.Infow("Successfully updated notification preference",
		"id", preference.ID,
		"enabled", preference.Enabled,
		"deliveryMethod", preference.DeliveryMethod,
	)

	return responses.NewNotificationPreferenceResponseSuccess(preference).JSON(c)
}

// UpdateNotificationPreferencesBulk updates multiple notification preferences in bulk
// @Summary      Bulk update notification preferences
// @Description  Update multiple notification preferences for the current user in bulk
// @Tags         Notification Preference
// @Accept       json
// @Param        body  body  []requests.NotificationPreferenceRequest  true  "Notification Preferences"
// @Success      200   {object} responses.NotificationPreferencesResponse
// @Failure      400   {object} responses.Error
// @Failure      500   {object} responses.Error
// @Router       /notification-preference/bulk [put]
// @Security     BearerAuth
func (h *NotificationPreferenceHandler) UpdateNotificationPreferencesBulk(c echo.Context) error {
	var reqs []requests.NotificationPreferenceRequest

	// Log the raw request body for debugging
	bodyBytes, err := io.ReadAll(c.Request().Body)
	if err == nil {
		h.server.Logger.Zap.Infow("Raw request body received", "body", string(bodyBytes))
		// Restore the body for binding
		c.Request().Body = io.NopCloser(bytes.NewReader(bodyBytes))
	}

	if err := c.Bind(&reqs); err != nil {
		h.server.Logger.Zap.Errorw("Failed to bind bulk update request", "error", err.Error())
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid request body: "+err.Error()).JSON(c)
	}

	// Log the parsed requests for debugging
	h.server.Logger.Zap.Infow("Parsed requests", "count", len(reqs))
	for i, req := range reqs {
		h.server.Logger.Zap.Infow("Request details",
			"index", i,
			"notificationType", req.NotificationType,
			"enabled", req.Enabled,
			"enabledType", fmt.Sprintf("%T", req.Enabled),
			"deliveryMethod", req.DeliveryMethod,
		)
	}

	userToken := c.Get("user").(*jwt.Token)
	claims := userToken.Claims.(*token.JwtCustomClaims)
	id := claims.ID

	queries := h.server.DB.Queries()
	ctx := c.Request().Context()

	var updatedPreferences []db.NotificationPreference

	// Process each request
	for i, req := range reqs {
		// Log detailed information about the request
		h.server.Logger.Zap.Infow("Processing request",
			"index", i,
			"notificationType", req.NotificationType,
			"enabled", req.Enabled,
			"enabledType", fmt.Sprintf("%T", req.Enabled),
			"enabledNil", req.Enabled == nil,
			"deliveryMethod", req.DeliveryMethod,
		)

		// Validate the request
		if err := req.Validate(); err != nil {
			h.server.Logger.Zap.Errorw("Validation failed for preference",
				"index", i,
				"error", err.Error(),
				"notificationType", req.NotificationType,
				"enabled", req.Enabled,
				"enabledNil", req.Enabled == nil,
			)
			return responses.NewErrorResponse(http.StatusBadRequest, fmt.Sprintf("Validation failed for preference at index %d: %s", i, err.Error())).JSON(c)
		}

		h.server.Logger.Zap.Infow("Validation passed for preference", "index", i, "notificationType", req.NotificationType)

		// Use upsert to create or update seamlessly
		params := db.UpsertNotificationPreferenceParams{
			UserID:           id,
			NotificationType: req.NotificationType,
			Enabled:          req.Enabled,
			DeliveryMethod:   &req.DeliveryMethod,
		}

		preference, err := queries.UpsertNotificationPreference(ctx, params)
		if err != nil {
			h.server.Logger.Zap.Errorw("Failed to upsert notification preference in bulk", "index", i, "notificationType", req.NotificationType, "error", err.Error())
			return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to update notification preferences: "+err.Error()).JSON(c)
		}
		updatedPreferences = append(updatedPreferences, preference)
	}

	h.server.Logger.Zap.Infow("Successfully bulk updated notification preferences", "count", len(updatedPreferences))

	return responses.NewNotificationPreferencesBulkResponse(updatedPreferences).JSON(c)
}
