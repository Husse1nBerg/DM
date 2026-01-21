package sadie

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"go.uber.org/zap"
)

// Client represents a SADIE Core API client
type Client struct {
	config     *config.SadieConfig
	logger     *logger.Logger
	httpClient *http.Client
}

// NewClient creates a new SADIE API client
func NewClient(cfg *config.SadieConfig, logger *logger.Logger) *Client {
	return &Client{
		config:     cfg,
		logger:     logger,
		httpClient: &http.Client{Timeout: 30 * time.Second},
	}
}

// getHeaders returns the headers needed for API requests
func (c *Client) getHeaders() http.Header {
	headers := http.Header{}
	headers.Set("Content-Type", "application/json")
	// SADIE API uses "ApiKey <key>" format, not "Bearer <key>"
	headers.Set("Authorization", "ApiKey "+c.config.APIKey)
	return headers
}

// DoRequest makes an HTTP request to the SADIE API
func (c *Client) DoRequest(ctx context.Context, method, endpoint string, body interface{}) (*http.Response, error) {
	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	requestURL := c.config.BaseURL + endpoint
	req, err := http.NewRequestWithContext(ctx, method, requestURL, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	headers := c.getHeaders()
	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	c.logger.DesugarZap.Debug("SADIE API request",
		zap.String("method", method),
		zap.String("endpoint", endpoint),
		zap.String("url", requestURL))

	return c.httpClient.Do(req)
}

// DoJSONRequest makes an HTTP request and decodes the JSON response
func (c *Client) DoJSONRequest(ctx context.Context, method, endpoint string, body, result interface{}) error {
	resp, err := c.DoRequest(ctx, method, endpoint, body)
	if err != nil {
		c.logger.DesugarZap.Error("SADIE API request failed",
			zap.String("method", method),
			zap.String("endpoint", endpoint),
			zap.Error(err))
		return fmt.Errorf("SADIE API request failed: %w", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		c.logger.DesugarZap.Error("SADIE API request failed to read response body",
			zap.String("method", method),
			zap.String("endpoint", endpoint),
			zap.Int("status", resp.StatusCode),
			zap.Error(err))
		return fmt.Errorf("SADIE API request failed to read response body: %w", err)
	}

	// Log the response for debugging
	c.logger.DesugarZap.Debug("SADIE API response received",
		zap.String("method", method),
		zap.String("endpoint", endpoint),
		zap.Int("status", resp.StatusCode),
		zap.String("response", string(responseBody)))

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		c.logger.DesugarZap.Error("SADIE API request failed",
			zap.String("method", method),
			zap.String("endpoint", endpoint),
			zap.Int("status", resp.StatusCode),
			zap.String("response", string(responseBody)))
		return fmt.Errorf("SADIE API request failed with status %d: %s", resp.StatusCode, string(responseBody))
	}

	// Only decode if we expect a result and have content
	if result != nil {
		if len(responseBody) == 0 {
			c.logger.DesugarZap.Warn("SADIE API returned empty response body",
				zap.String("method", method),
				zap.String("endpoint", endpoint))
			return fmt.Errorf("SADIE API returned empty response body")
		}

		if err := json.Unmarshal(responseBody, result); err != nil {
			c.logger.DesugarZap.Error("SADIE API request failed to decode response",
				zap.String("method", method),
				zap.String("endpoint", endpoint),
				zap.String("response", string(responseBody)),
				zap.Error(err))
			return fmt.Errorf("SADIE API request failed to decode response: %w", err)
		}
	}

	return nil
}

// Assistant represents an assistant from the SADIE API
type Assistant struct {
	ID            string `json:"id"`
	Name          string `json:"name"`
	PhoneNumberID string `json:"phoneNumberId,omitempty"`
}

// AssistantsResponse represents the response from getting assistants
type AssistantsResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// GetAssistants retrieves all assistants for the tenant
func (c *Client) GetAssistants(ctx context.Context) ([]Assistant, error) {
	var result AssistantsResponse

	endpoint := "/assistants"

	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get assistants: %w", err)
	}

	// Parse the data field which contains the assistants array
	assistants := []Assistant{}
	if data, ok := result.Data.(map[string]interface{}); ok {
		if assistantsList, ok := data["assistants"].([]interface{}); ok {
			for _, a := range assistantsList {
				if assistantMap, ok := a.(map[string]interface{}); ok {
					assistant := Assistant{}
					if id, ok := assistantMap["id"].(string); ok {
						assistant.ID = id
					}
					if name, ok := assistantMap["name"].(string); ok {
						assistant.Name = name
					}
					if phoneNumberID, ok := assistantMap["phoneNumberId"].(string); ok {
						assistant.PhoneNumberID = phoneNumberID
					}
					assistants = append(assistants, assistant)
				}
			}
		}
	}

	return assistants, nil
}

// GetAssistantByID retrieves a specific assistant by ID
func (c *Client) GetAssistantByID(ctx context.Context, assistantID string) (*Assistant, error) {
	var result AssistantsResponse

	endpoint := fmt.Sprintf("/assistants/%s", assistantID)

	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get assistant: %w", err)
	}

	// Parse the data field
	assistant := &Assistant{}
	if data, ok := result.Data.(map[string]interface{}); ok {
		if id, ok := data["id"].(string); ok {
			assistant.ID = id
		}
		if name, ok := data["name"].(string); ok {
			assistant.Name = name
		}
		if phoneNumberID, ok := data["phoneNumberId"].(string); ok {
			assistant.PhoneNumberID = phoneNumberID
		}
	}

	return assistant, nil
}

// PhoneNumber represents a phone number from the SADIE API
type PhoneNumber struct {
	ID          string `json:"id"`
	PhoneNumber string `json:"phoneNumber"`
	Name        string `json:"name,omitempty"`
}

// PhoneNumbersResponse represents the response from getting phone numbers
type PhoneNumbersResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// GetPhoneNumbers retrieves all phone numbers for the tenant
func (c *Client) GetPhoneNumbers(ctx context.Context) ([]PhoneNumber, error) {
	var result PhoneNumbersResponse

	endpoint := "/phone-numbers"

	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get phone numbers: %w", err)
	}

	// Parse the data field which contains the phone numbers array
	phoneNumbers := []PhoneNumber{}
	if data, ok := result.Data.(map[string]interface{}); ok {
		if phoneNumbersList, ok := data["phoneNumbers"].([]interface{}); ok {
			for _, pn := range phoneNumbersList {
				if phoneNumberMap, ok := pn.(map[string]interface{}); ok {
					phoneNumber := PhoneNumber{}
					if id, ok := phoneNumberMap["id"].(string); ok {
						phoneNumber.ID = id
					}
					if pnStr, ok := phoneNumberMap["phoneNumber"].(string); ok {
						phoneNumber.PhoneNumber = pnStr
					}
					if name, ok := phoneNumberMap["name"].(string); ok {
						phoneNumber.Name = name
					}
					phoneNumbers = append(phoneNumbers, phoneNumber)
				}
			}
		}
	}

	return phoneNumbers, nil
}

// GetPhoneNumberByID retrieves a phone number by ID
func (c *Client) GetPhoneNumberByID(ctx context.Context, phoneNumberID string) (*PhoneNumber, error) {
	var result PhoneNumbersResponse

	endpoint := fmt.Sprintf("/phone-numbers/%s", phoneNumberID)

	err := c.DoJSONRequest(ctx, http.MethodGet, endpoint, nil, &result)
	if err != nil {
		return nil, fmt.Errorf("failed to get phone number: %w", err)
	}

	// Parse the data field
	phoneNumber := &PhoneNumber{}
	if data, ok := result.Data.(map[string]interface{}); ok {
		if phoneNumberData, ok := data["phoneNumber"].(map[string]interface{}); ok {
			if id, ok := phoneNumberData["id"].(string); ok {
				phoneNumber.ID = id
			}
			if pn, ok := phoneNumberData["phoneNumber"].(string); ok {
				phoneNumber.PhoneNumber = pn
			}
			if name, ok := phoneNumberData["name"].(string); ok {
				phoneNumber.Name = name
			}
		}
	}

	return phoneNumber, nil
}

// UpdateAssistantWebhookURLRequest represents the request body for updating assistant webhook URL
type UpdateAssistantWebhookURLRequest struct {
	WebhookURL string `json:"webhookUrl"`
}

// UpdateAssistantWebhookURLResponse represents the response from updating assistant webhook URL
type UpdateAssistantWebhookURLResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// UpdateAssistantWebhookURL updates the webhook URL for a specific assistant
func (c *Client) UpdateAssistantWebhookURL(ctx context.Context, assistantID, webhookURL string) error {
	endpoint := fmt.Sprintf("/assistants/%s", assistantID)

	requestBody := UpdateAssistantWebhookURLRequest{
		WebhookURL: webhookURL,
	}

	var result UpdateAssistantWebhookURLResponse
	err := c.DoJSONRequest(ctx, http.MethodPatch, endpoint, requestBody, &result)
	if err != nil {
		return fmt.Errorf("failed to update assistant webhook URL: %w", err)
	}

	return nil
}

// UpdateAssistantPhoneNumberRequest represents the request body for updating assistant phone number
type UpdateAssistantPhoneNumberRequest struct {
	PhoneNumberID string `json:"phoneNumberId"`
}

// UpdateAssistantPhoneNumberResponse represents the response from updating assistant phone number
type UpdateAssistantPhoneNumberResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data"`
}

// UpdateAssistantPhoneNumber assigns a phone number to a specific assistant
func (c *Client) UpdateAssistantPhoneNumber(ctx context.Context, assistantID, phoneNumberID string) error {
	endpoint := fmt.Sprintf("/assistants/%s", assistantID)

	requestBody := UpdateAssistantPhoneNumberRequest{
		PhoneNumberID: phoneNumberID,
	}

	var result UpdateAssistantPhoneNumberResponse
	err := c.DoJSONRequest(ctx, http.MethodPatch, endpoint, requestBody, &result)
	if err != nil {
		return fmt.Errorf("failed to update assistant phone number: %w", err)
	}

	return nil
}