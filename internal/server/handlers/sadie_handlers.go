package handlers

import (
	"encoding/json"
	"io"
	"time"

	"github.com/dockworks/dm-web-backend/internal/responses"
	s "github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/sadie"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type SadieHandler struct {
	server *s.Server
}

func NewSadieHandler(server *s.Server) *SadieHandler {
	return &SadieHandler{server: server}
}

// getOrgAndSystemIDFromRequest extracts organization ID and system ID from request
// It tries to get from query params first, then from webhook payload
func (h *SadieHandler) getOrgAndSystemIDFromRequest(c echo.Context) (uuid.UUID, string, error) {
	// Try to get from query parameters first
	orgIDStr := c.QueryParam("organizationId")
	systemID := c.QueryParam("systemId")

	if orgIDStr != "" && systemID != "" {
		orgID, err := uuid.Parse(orgIDStr)
		if err != nil {
			return uuid.Nil, "", responses.NewSadieErrorResponse(
				"Invalid organization ID",
				"Could not parse organization ID from request",
				[]string{
					"Verify the organizationId parameter is a valid UUID",
				},
				nil,
			).JSON(c)
		}
		return orgID, systemID, nil
	}

	// If not found, return error asking for organization and system ID
	return uuid.Nil, "", responses.NewSadieErrorResponse(
		"Missing organization and system ID",
		"Organization ID and System ID are required to access DockMaster API",
		[]string{
			"Provide organizationId and systemId as query parameters",
			"These should be configured for your SADIE assistant",
		},
		nil,
	).JSON(c)
}

// WebhookResponse is for Swagger documentation
type WebhookResponse struct {
	Success bool `json:"success"`
}

// WebhookPostHandler handles POST requests to /webhooks
//
//	@Summary		Webhook POST endpoint
//	@Description	Receives webhook POST requests
//	@Tags			SADIE
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	WebhookResponse
//	@Router			/webhooks [post]
func (h *SadieHandler) WebhookPostHandler(c echo.Context) error {
	h.server.Logger.Zap.Info("Webhook POST request received")

	// Log headers
	headers := make(map[string]string)
	for k, v := range c.Request().Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	h.server.Logger.Zap.Info("Headers", zap.Any("headers", headers))

	// Try to read body
	contentType := c.Request().Header.Get("Content-Type")
	if contentType != "" {
		var bodyText string
		bodyBytes, err := io.ReadAll(c.Request().Body)
		if err == nil {
			bodyText = string(bodyBytes)
			if len(bodyText) > 1000 {
				bodyText = bodyText[:1000]
			}
			h.server.Logger.Zap.Info("Request body", zap.String("body", bodyText))
		} else {
			h.server.Logger.Zap.Info("Could not read request body", zap.Error(err))
		}
	}

	return c.NoContent(200)
}

// WebhookGetHandler handles GET requests to /webhooks
//
//	@Summary		Webhook GET endpoint
//	@Description	Returns webhook endpoint status
//	@Tags			SADIE
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	map[string]interface{}
//	@Router			/webhooks [get]
func (h *SadieHandler) WebhookGetHandler(c echo.Context) error {
	h.server.Logger.Zap.Info("Webhook GET request received")

	headers := make(map[string]string)
	for k, v := range c.Request().Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	h.server.Logger.Zap.Info("Headers", zap.Any("headers", headers))

	response := map[string]interface{}{
		"message":   "Webhook endpoint is active",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return c.JSON(200, response)
}

// WebhookOptionsHandler handles OPTIONS requests to /webhooks
//
//	@Summary		Webhook OPTIONS endpoint
//	@Description	Handles CORS preflight requests
//	@Tags			SADIE
//	@Accept			json
//	@Produce		json
//	@Success		200
//	@Router			/webhooks [options]
func (h *SadieHandler) WebhookOptionsHandler(c echo.Context) error {
	h.server.Logger.Zap.Info("Webhook OPTIONS request received")

	headers := make(map[string]string)
	for k, v := range c.Request().Header {
		if len(v) > 0 {
			headers[k] = v[0]
		}
	}
	h.server.Logger.Zap.Info("Headers", zap.Any("headers", headers))

	return c.NoContent(200)
}

// GetAvailableSlipsResponse is for Swagger documentation
type GetAvailableSlipsResponse struct {
	Success     bool                   `json:"success"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Description string                 `json:"description"`
	Steps       []string               `json:"steps"`
}

// GetAvailableSlipsGetHandler handles GET requests to /getAvailableSlips
//
//	@Summary		Get available slips (GET)
//	@Description	Checks for available slips
//	@Tags			SADIE
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			startDate	query		string	false	"Start date (YYYY-MM-DD)"
//	@Param			endDate		query		string	false	"End date (YYYY-MM-DD)"
//	@Success		200			{object}	GetAvailableSlipsResponse
//	@Router			/getAvailableSlips [get]
func (h *SadieHandler) GetAvailableSlipsGetHandler(c echo.Context) error {
	// Get organization and system ID from request context or query params
	orgID, systemID, err := h.getOrgAndSystemIDFromRequest(c)
	if err != nil {
		return err
	}

	// Extract query parameters (excluding organizationId and systemId which are handled separately)
	params := make(map[string]string)
	for key, values := range c.QueryParams() {
		if len(values) > 0 && key != "organizationId" && key != "systemId" {
			params[key] = values[0]
		}
	}

	// Call DockMaster API through dm-web-backend's DME client
	// This ensures the flow: SADIE -> dm-web-backend -> DockMaster_API
	resp, err := h.server.DME.DoRequest(c.Request().Context(), "GET", "/api/v1/MarinaOps/Slips/GetAvailableSlips", nil, orgID, systemID, params)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to get available slips from DockMaster API", zap.Error(err))
		return responses.NewSadieErrorResponse(
			"Failed to check slip availability",
			"Could not retrieve available slips from DockMaster API",
			[]string{
				"Please try again later",
				"Contact support if the issue persists",
			},
			map[string]interface{}{
				"error": err.Error(),
			},
		).JSON(c)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to read response from DockMaster API", zap.Error(err))
		return responses.NewSadieErrorResponse(
			"Failed to process slip availability",
			"Could not read response from DockMaster API",
			[]string{
				"Please try again later",
			},
			nil,
		).JSON(c)
	}

	// Parse the response
	var dmeResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &dmeResponse); err != nil {
		h.server.Logger.Zap.Error("Failed to parse DockMaster API response", zap.Error(err))
		return responses.NewSadieErrorResponse(
			"Failed to process slip availability",
			"Invalid response from DockMaster API",
			[]string{
				"Please try again later",
			},
			nil,
		).JSON(c)
	}

	// Check if slips are available
	hasAvailableSlips := false
	if data, ok := dmeResponse["data"].(map[string]interface{}); ok {
		if slips, ok := data["slips"].([]interface{}); ok && len(slips) > 0 {
			hasAvailableSlips = true
		}
	}

	steps := []string{
		"Say that no slips are available for the requested dates",
	}
	if hasAvailableSlips {
		steps = []string{
			"Say that slips are available for the requested dates",
			"Ask if they would like to make a reservation",
		}
	}

	response := responses.NewSadieSuccessResponse(
		dmeResponse,
		"Checked slip availability",
		steps,
	)
	return response.JSON(c)
}

// GetAvailableSlipsPostHandler handles POST requests to /getAvailableSlips
//
//	@Summary		Get available slips (POST)
//	@Description	Checks for available slips via POST
//	@Tags			SADIE
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			organizationId	query		string	false	"Organization ID"
//	@Param			systemId		query		string	false	"System ID"
//	@Param			startDate		query		string	false	"Start date (YYYY-MM-DD)"
//	@Param			endDate			query		string	false	"End date (YYYY-MM-DD)"
//	@Success		200				{object}	GetAvailableSlipsResponse
//	@Router			/getAvailableSlips [post]
func (h *SadieHandler) GetAvailableSlipsPostHandler(c echo.Context) error {
	// Get organization and system ID from request context or query params
	orgID, systemID, err := h.getOrgAndSystemIDFromRequest(c)
	if err != nil {
		return err
	}

	// Extract query parameters
	params := make(map[string]string)
	for key, values := range c.QueryParams() {
		if len(values) > 0 && key != "organizationId" && key != "systemId" {
			params[key] = values[0]
		}
	}

	// Call DockMaster API to get available slips
	resp, err := h.server.DME.DoRequest(c.Request().Context(), "GET", "/api/v1/MarinaOps/Slips/GetAvailableSlips", nil, orgID, systemID, params)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to get available slips from DockMaster API", zap.Error(err))
		return responses.NewSadieErrorResponse(
			"Failed to check slip availability",
			"Could not retrieve available slips from DockMaster API",
			[]string{
				"Please try again later",
				"Contact support if the issue persists",
			},
			map[string]interface{}{
				"error": err.Error(),
			},
		).JSON(c)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to read response from DockMaster API", zap.Error(err))
		return responses.NewSadieErrorResponse(
			"Failed to process slip availability",
			"Could not read response from DockMaster API",
			[]string{
				"Please try again later",
			},
			nil,
		).JSON(c)
	}

	// Parse the response
	var dmeResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &dmeResponse); err != nil {
		h.server.Logger.Zap.Error("Failed to parse DockMaster API response", zap.Error(err))
		return responses.NewSadieErrorResponse(
			"Failed to process slip availability",
			"Invalid response from DockMaster API",
			[]string{
				"Please try again later",
			},
			nil,
		).JSON(c)
	}

	// Check if slips are available
	hasAvailableSlips := false
	if data, ok := dmeResponse["data"].(map[string]interface{}); ok {
		if slips, ok := data["slips"].([]interface{}); ok && len(slips) > 0 {
			hasAvailableSlips = true
		}
	}

	steps := []string{
		"Say that no slips are available for the requested dates",
	}
	if hasAvailableSlips {
		steps = []string{
			"Say that slips are available for the requested dates",
			"Ask if they would like to proceed with the booking",
		}
	}

	response := responses.NewSadieSuccessResponse(
		dmeResponse,
		"Checked slip availability",
		steps,
	)
	return response.JSON(c)
}

// MakeReservationResponse is for Swagger documentation
type MakeReservationResponse struct {
	Success     bool                   `json:"success"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Description string                 `json:"description"`
	Steps       []string               `json:"steps"`
}

// MakeReservationPostHandler handles POST requests to /makeReservation
//
//	@Summary		Make reservation
//	@Description	Creates a new reservation
//	@Tags			SADIE
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			organizationId	query		string					false	"Organization ID"
//	@Param			systemId		query		string					false	"System ID"
//	@Param			request			body		map[string]interface{}	true	"Reservation data"
//	@Success		200				{object}	MakeReservationResponse
//	@Router			/makeReservation [post]
func (h *SadieHandler) MakeReservationPostHandler(c echo.Context) error {
	// Get organization and system ID from request
	orgID, systemID, err := h.getOrgAndSystemIDFromRequest(c)
	if err != nil {
		return err
	}

	// Read request body
	var reservationData map[string]interface{}
	if err := c.Bind(&reservationData); err != nil {
		return responses.NewSadieErrorResponse(
			"Invalid request body",
			"Could not parse reservation data from request",
			[]string{
				"Please provide valid reservation information",
			},
			nil,
		).JSON(c)
	}

	// Call DockMaster API to create reservation
	resp, err := h.server.DME.DoRequest(c.Request().Context(), "POST", "/api/v1/MarinaOps/Reservations", reservationData, orgID, systemID, nil)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to create reservation in DockMaster API", zap.Error(err))
		return responses.NewSadieErrorResponse(
			"Failed to create reservation",
			"Could not create reservation in DockMaster API",
			[]string{
				"Apologize for the inconvenience",
				"Ask if they would like to try again or speak with a representative",
			},
			map[string]interface{}{
				"error": err.Error(),
			},
		).JSON(c)
	}
	defer resp.Body.Close()

	// Read response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to read response from DockMaster API", zap.Error(err))
		return responses.NewSadieErrorResponse(
			"Failed to process reservation",
			"Could not read response from DockMaster API",
			[]string{
				"Please try again later",
			},
			nil,
		).JSON(c)
	}

	// Parse the response
	var dmeResponse map[string]interface{}
	if err := json.Unmarshal(responseBody, &dmeResponse); err != nil {
		h.server.Logger.Zap.Error("Failed to parse DockMaster API response", zap.Error(err))
		return responses.NewSadieErrorResponse(
			"Failed to process reservation",
			"Invalid response from DockMaster API",
			[]string{
				"Please try again later",
			},
			nil,
		).JSON(c)
	}

	// Check if reservation was successful
	isSuccess := resp.StatusCode >= 200 && resp.StatusCode < 300

	steps := []string{
		"Apologize that the reservation could not be created",
		"Ask if they would like to try again or speak with a representative",
	}
	if isSuccess {
		steps = []string{
			"Confirm that the reservation has been created successfully",
			"Tell them that a registration link will be sent to them via text to finish the registration process",
			"Tell them that they can always call back to cancel the reservation",
		}
	}

	response := responses.NewSadieSuccessResponse(
		dmeResponse,
		"Reservation creation processed",
		steps,
	)
	return response.JSON(c)
}

// HeartbeatResponse is for Swagger documentation
type HeartbeatResponse struct {
	Success   bool   `json:"success"`
	Message   string `json:"message"`
	Timestamp string `json:"timestamp"`
}

// HeartbeatGetHandler handles GET requests to /heartbeat
//
//	@Summary		Heartbeat endpoint
//	@Description	Checks if the API is working
//	@Tags			SADIE
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	HeartbeatResponse
//	@Router			/heartbeat [get]
func (h *SadieHandler) HeartbeatGetHandler(c echo.Context) error {
	response := map[string]interface{}{
		"success":   true,
		"message":   "API is working",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	}
	return c.JSON(200, response)
}

// GetAssistantPhoneNumberRequest represents the request to get assistant phone number
type GetAssistantPhoneNumberRequest struct {
	AssistantID string `json:"assistant_id,omitempty"`
}

// GetAssistantPhoneNumberResponse represents the response with phone number info
type GetAssistantPhoneNumberResponse struct {
	Success     bool                   `json:"success"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Description string                 `json:"description,omitempty"`
	Steps       []string               `json:"steps,omitempty"`
}

// GetAssistantPhoneNumberHandler handles GET/POST requests to get assistant phone number for testing
//
//	@Summary		Get assistant phone number
//	@Description	Retrieves the phone number associated with an assistant so you can call it to test
//	@Tags			SADIE
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			assistant_id	query		string	false	"Assistant ID (optional, will return first assistant if not provided)"
//	@Success		200				{object}	GetAssistantPhoneNumberResponse
//	@Router			/getAssistantPhoneNumber [get]
func (h *SadieHandler) GetAssistantPhoneNumberHandler(c echo.Context) error {
	assistantID := c.QueryParam("assistant_id")

	// Create SADIE client
	sadieClient := sadie.NewClient(&h.server.Config.Sadie, h.server.Logger)

	var assistant *sadie.Assistant
	var err error

	if assistantID != "" {
		// Get specific assistant
		assistant, err = sadieClient.GetAssistantByID(c.Request().Context(), assistantID)
		if err != nil {
			h.server.Logger.Zap.Error("Failed to get assistant", zap.Error(err))
			return responses.NewSadieErrorResponse(
				"Failed to retrieve assistant",
				"Could not retrieve assistant from SADIE API",
				[]string{
					"Verify the assistant_id is correct",
					"Check that SADIE API credentials are valid",
				},
				map[string]interface{}{
					"error": err.Error(),
				},
			).JSON(c)
		}
	} else {
		// Get all assistants and use the first one
		assistants, err := sadieClient.GetAssistants(c.Request().Context())
		if err != nil {
			h.server.Logger.Zap.Error("Failed to get assistants", zap.Error(err))
			return responses.NewSadieErrorResponse(
				"Failed to retrieve assistants",
				"Could not retrieve assistants from SADIE API",
				[]string{
					"Check that SADIE API credentials are valid",
					"Ensure you have at least one assistant configured",
				},
				map[string]interface{}{
					"error": err.Error(),
				},
			).JSON(c)
		}

		if len(assistants) == 0 {
			return responses.NewSadieErrorResponse(
				"No assistants found",
				"You don't have any assistants configured",
				[]string{
					"Create an assistant first using the SADIE API",
					"Or provide a specific assistant_id",
				},
				nil,
			).JSON(c)
		}

		assistant = &assistants[0]
	}

	// Get the phone number details
	var phoneNumber *sadie.PhoneNumber
	if assistant.PhoneNumberID != "" {
		phoneNumber, err = sadieClient.GetPhoneNumberByID(c.Request().Context(), assistant.PhoneNumberID)
		if err != nil {
			h.server.Logger.Zap.Warn("Failed to get phone number details", zap.Error(err), zap.String("phoneNumberId", assistant.PhoneNumberID))
			// Continue without phone number details
		}
	}

	responseData := map[string]interface{}{
		"assistant_id":   assistant.ID,
		"assistant_name": assistant.Name,
	}

	if phoneNumber != nil {
		responseData["phone_number"] = phoneNumber.PhoneNumber
		responseData["phone_number_id"] = phoneNumber.ID
		responseData["phone_number_name"] = phoneNumber.Name
	} else if assistant.PhoneNumberID != "" {
		responseData["phone_number_id"] = assistant.PhoneNumberID
		responseData["note"] = "Phone number ID found but details could not be retrieved"
	} else {
		responseData["note"] = "No phone number assigned to this assistant"
	}

	steps := []string{
		"Call the phone number shown above to test your SADIE agent",
		"The agent will answer and you can interact with it",
		"Monitor the webhook endpoints to see call events",
	}

	if phoneNumber == nil {
		steps = []string{
			"No phone number is assigned to this assistant",
			"Assign a phone number to the assistant using the SADIE API",
			"Then call that number to test the agent",
		}
	}

	response := responses.NewSadieSuccessResponse(
		responseData,
		"Assistant phone number retrieved successfully",
		steps,
	)

	return response.JSON(c)
}

// GetPhoneNumbersHandler handles GET requests to get all phone numbers
//
//	@Summary		Get all phone numbers
//	@Description	Retrieves all phone numbers available for the tenant
//	@Tags			SADIE
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Success		200	{object}	map[string]interface{}
//	@Router			/getPhoneNumbers [get]
func (h *SadieHandler) GetPhoneNumbersHandler(c echo.Context) error {
	// Create SADIE client
	sadieClient := sadie.NewClient(&h.server.Config.Sadie, h.server.Logger)

	phoneNumbers, err := sadieClient.GetPhoneNumbers(c.Request().Context())
	if err != nil {
		h.server.Logger.Zap.Error("Failed to get phone numbers", zap.Error(err))
		return responses.NewSadieErrorResponse(
			"Failed to retrieve phone numbers",
			"Could not retrieve phone numbers from SADIE API",
			[]string{
				"Check that SADIE API credentials are valid",
				"Verify you have phone numbers configured",
			},
			map[string]interface{}{
				"error": err.Error(),
			},
		).JSON(c)
	}

	responseData := map[string]interface{}{
		"phone_numbers": phoneNumbers,
		"count":         len(phoneNumbers),
	}

	steps := []string{
		"Use one of these phone numbers to assign to an assistant",
		"Call the phone number to test your SADIE agent",
		"Monitor the webhook endpoints to see call events",
	}

	if len(phoneNumbers) == 0 {
		steps = []string{
			"No phone numbers found",
			"Create a phone number using the SADIE API",
			"Then assign it to an assistant",
		}
	}

	response := responses.NewSadieSuccessResponse(
		responseData,
		"Phone numbers retrieved successfully",
		steps,
	)

	return response.JSON(c)
}

// UpdateAgentWebhookRequest represents the request to update agent webhook URL
type UpdateAgentWebhookRequest struct {
	AssistantID string `json:"assistant_id,omitempty" query:"assistant_id"`
	WebhookURL  string `json:"webhook_url,omitempty"`
}

// UpdateAgentWebhookResponse represents the response from updating agent webhook
type UpdateAgentWebhookResponse struct {
	Success     bool                   `json:"success"`
	Data        map[string]interface{} `json:"data,omitempty"`
	Description string                 `json:"description,omitempty"`
	Steps       []string               `json:"steps,omitempty"`
}

// UpdateAgentWebhookHandler handles POST/PATCH requests to update the agent's webhook URL in Sadie
//
//	@Summary		Update agent webhook URL
//	@Description	Updates the webhook URL for an assistant in SADIE to point to this server's webhook endpoint
//	@Tags			SADIE
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			assistant_id	query		string	false	"Assistant ID (optional, will update first assistant if not provided)"
//	@Param			webhook_url		body		string	false	"Webhook URL (optional, will use configured webhook URL if not provided)"
//	@Success		200				{object}	UpdateAgentWebhookResponse
//	@Router			/updateAgentWebhook [post]
func (h *SadieHandler) UpdateAgentWebhookHandler(c echo.Context) error {
	// Get assistant ID from query param or request body
	assistantID := c.QueryParam("assistant_id")
	webhookURL := h.server.Config.Sadie.WebhookURL

	// Try to get webhook URL from request body if provided
	var requestBody UpdateAgentWebhookRequest
	if err := c.Bind(&requestBody); err == nil {
		if requestBody.AssistantID != "" {
			assistantID = requestBody.AssistantID
		}
		if requestBody.WebhookURL != "" {
			webhookURL = requestBody.WebhookURL
		}
	}

	// If assistant ID is not provided, get the first assistant
	if assistantID == "" {
		sadieClient := sadie.NewClient(&h.server.Config.Sadie, h.server.Logger)
		assistants, err := sadieClient.GetAssistants(c.Request().Context())
		if err != nil {
			h.server.Logger.Zap.Error("Failed to get assistants", zap.Error(err))
			return responses.NewSadieErrorResponse(
				"Failed to retrieve assistants",
				"Could not retrieve assistants from SADIE API to determine which assistant to update",
				[]string{
					"Provide an assistant_id in the request",
					"Check that SADIE API credentials are valid",
				},
				map[string]interface{}{
					"error": err.Error(),
				},
			).JSON(c)
		}

		if len(assistants) == 0 {
			return responses.NewSadieErrorResponse(
				"No assistants found",
				"You don't have any assistants configured",
				[]string{
					"Create an assistant first using the SADIE API",
					"Or provide a specific assistant_id",
				},
				nil,
			).JSON(c)
		}

		assistantID = assistants[0].ID
	}

	// Create SADIE client and update the webhook URL
	sadieClient := sadie.NewClient(&h.server.Config.Sadie, h.server.Logger)

	err := sadieClient.UpdateAssistantWebhookURL(c.Request().Context(), assistantID, webhookURL)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to update assistant webhook URL", zap.Error(err), zap.String("assistantID", assistantID), zap.String("webhookURL", webhookURL))
		return responses.NewSadieErrorResponse(
			"Failed to update assistant webhook URL",
			"Could not update the webhook URL in SADIE API",
			[]string{
				"Verify the assistant_id is correct",
				"Check that the webhook URL is valid",
				"Ensure SADIE API credentials are valid",
			},
			map[string]interface{}{
				"error":        err.Error(),
				"assistant_id": assistantID,
				"webhook_url":  webhookURL,
			},
		).JSON(c)
	}

	responseData := map[string]interface{}{
		"assistant_id": assistantID,
		"webhook_url":  webhookURL,
		"message":      "Webhook URL updated successfully",
	}

	steps := []string{
		"The agent's webhook URL has been updated in SADIE",
		"SADIE will now send webhook events to the configured URL",
		"Test by making a call to the assistant's phone number",
	}

	response := responses.NewSadieSuccessResponse(
		responseData,
		"Agent webhook URL updated successfully",
		steps,
	)

	return response.JSON(c)
}

// AssignPhoneNumberRequest represents the request to assign a phone number to an assistant
type AssignPhoneNumberRequest struct {
	AssistantID   string `json:"assistant_id,omitempty" query:"assistant_id"`
	PhoneNumberID string `json:"phone_number_id" query:"phone_number_id"`
}

// AssignPhoneNumberHandler handles POST/PATCH requests to assign a phone number to an assistant
//
//	@Summary		Assign phone number to assistant
//	@Description	Assigns a phone number to an assistant in SADIE
//	@Tags			SADIE
//	@Accept			json
//	@Produce		json
//	@Security		ApiKeyAuth
//	@Param			assistant_id		query		string	false	"Assistant ID (optional, will use first assistant if not provided)"
//	@Param			phone_number_id		query		string	true	"Phone Number ID to assign"
//	@Param			assistant_id		body		string	false	"Assistant ID (optional)"
//	@Param			phone_number_id		body		string	true	"Phone Number ID to assign"
//	@Success		200					{object}	map[string]interface{}
//	@Router			/assignPhoneNumber [post]
func (h *SadieHandler) AssignPhoneNumberHandler(c echo.Context) error {
	// Get assistant ID from query param or request body
	assistantID := c.QueryParam("assistant_id")
	phoneNumberID := c.QueryParam("phone_number_id")

	// Try to get from request body if provided
	var requestBody AssignPhoneNumberRequest
	if err := c.Bind(&requestBody); err == nil {
		if requestBody.AssistantID != "" {
			assistantID = requestBody.AssistantID
		}
		if requestBody.PhoneNumberID != "" {
			phoneNumberID = requestBody.PhoneNumberID
		}
	}

	// Validate phone number ID is provided
	if phoneNumberID == "" {
		return responses.NewSadieErrorResponse(
			"Missing phone number ID",
			"Phone number ID is required to assign a phone number",
			[]string{
				"Provide phone_number_id as a query parameter or in the request body",
				"Use /getPhoneNumbers endpoint to see available phone numbers",
			},
			nil,
		).JSON(c)
	}

	// If assistant ID not provided, get the first assistant
	if assistantID == "" {
		sadieClient := sadie.NewClient(&h.server.Config.Sadie, h.server.Logger)
		assistants, err := sadieClient.GetAssistants(c.Request().Context())
		if err != nil {
			h.server.Logger.Zap.Error("Failed to get assistants", zap.Error(err))
			return responses.NewSadieErrorResponse(
				"Failed to retrieve assistants",
				"Could not retrieve assistants from SADIE API",
				[]string{
					"Check that SADIE API credentials are valid",
					"Or provide a specific assistant_id",
				},
				map[string]interface{}{
					"error": err.Error(),
				},
			).JSON(c)
		}

		if len(assistants) == 0 {
			return responses.NewSadieErrorResponse(
				"No assistants found",
				"You don't have any assistants configured",
				[]string{
					"Create an assistant first using the SADIE API",
					"Or provide a specific assistant_id",
				},
				nil,
			).JSON(c)
		}

		assistantID = assistants[0].ID
		h.server.Logger.Zap.Info("Using first assistant", zap.String("assistant_id", assistantID))
	}

	// Create SADIE client and assign the phone number
	sadieClient := sadie.NewClient(&h.server.Config.Sadie, h.server.Logger)

	err := sadieClient.UpdateAssistantPhoneNumber(c.Request().Context(), assistantID, phoneNumberID)
	if err != nil {
		h.server.Logger.Zap.Error("Failed to assign phone number", zap.Error(err), zap.String("assistantID", assistantID), zap.String("phoneNumberID", phoneNumberID))
		return responses.NewSadieErrorResponse(
			"Failed to assign phone number",
			"Could not assign the phone number to the assistant in SADIE API",
			[]string{
				"Verify the assistant_id is correct",
				"Check that the phone_number_id is valid",
				"Ensure SADIE API credentials are valid",
			},
			map[string]interface{}{
				"error": err.Error(),
			},
		).JSON(c)
	}

	// Get the phone number details to return in response
	phoneNumber, err := sadieClient.GetPhoneNumberByID(c.Request().Context(), phoneNumberID)
	if err != nil {
		h.server.Logger.Zap.Warn("Failed to get phone number details", zap.Error(err))
	}

	responseData := map[string]interface{}{
		"assistant_id":   assistantID,
		"phone_number_id": phoneNumberID,
		"success":        true,
	}

	if phoneNumber != nil {
		responseData["phone_number"] = phoneNumber.PhoneNumber
		responseData["phone_number_name"] = phoneNumber.Name
	}

	steps := []string{
		"The phone number has been assigned to the assistant",
		"Call the phone number to test your SADIE agent",
		"Monitor the webhook endpoints to see call events",
	}

	response := responses.NewSadieSuccessResponse(
		responseData,
		"Phone number assigned successfully",
		steps,
	)

	return response.JSON(c)
}