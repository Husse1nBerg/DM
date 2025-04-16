package dme

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// ClientConfig holds configuration options for the DME client
type ClientConfig struct {
	BaseURL    string
	AuthURL    string
	APIVersion string
	IsOldAPI   bool
	Logger     *logger.Logger
	DB         pg.DBService
}

// Client represents a DockMaster API client
type Client struct {
	config     *ClientConfig
	logger     *logger.Logger
	httpClient *http.Client
	db         pg.DBService
}

// SystemID represents a system connection available to the user
type SystemID struct {
	Name      string `json:"name"`
	SystemID  string `json:"systemId"`
	IsDefault bool   `json:"isDefault"`
}

// NewClient creates a new DME API client
func NewClient(cfg *ClientConfig) *Client {
	return &Client{
		config:     cfg,
		logger:     cfg.Logger,
		httpClient: &http.Client{Timeout: 10 * time.Second},
		db:         cfg.DB,
	}
}

// Authenticate authenticates with the DME API for a specific organization
func (c *Client) Authenticate(ctx context.Context, organizationID uuid.UUID) error {
	queries := c.db.Queries()

	// Get credentials from database
	credential, err := queries.GetDMECredentialsByOrgID(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("failed to get DME credentials for organization %s: %w", organizationID, err)
	}

	// Check if we need to authenticate or refresh token
	if credential.AccessToken == nil || credential.RefreshToken == nil ||
		credential.ExpiryDate.Time.Before(time.Now().Add(30*time.Second)) {
		// Need to authenticate
		if err := c.authenticateOrg(ctx, credential); err != nil {
			return fmt.Errorf("failed to authenticate organization %s: %w", organizationID, err)
		}
	}

	return nil
}

// authenticateOrg performs authentication for an organization and updates the tokens in the database
func (c *Client) authenticateOrg(ctx context.Context, credential db.DmeCredential) error {
	var (
		authToken    string
		refreshToken string
		expiration   time.Time
		systemIDs    []SystemID
		err          error
	)

	if credential.IsOldApi != nil && *credential.IsOldApi {
		authToken, expiration, err = c.getOldAPIToken(ctx, credential.Username, *credential.Password)
		if err != nil {
			return err
		}
	} else {
		authToken, refreshToken, expiration, systemIDs, err = c.getToken(ctx, credential.Username, *credential.Password)
		if err != nil {
			return err
		}
	}

	// Update tokens in database
	expiryTimestamp := pgtype.Timestamp{Time: expiration, Valid: true}

	// Update the token in the database
	params := db.UpdateDMETokenParams{
		OrganizationID: credential.OrganizationID,
		AccessToken:    &authToken,
		RefreshToken:   &refreshToken,
		ExpiryDate:     expiryTimestamp,
	}
	// Update the system IDs in the database
	for _, systemID := range systemIDs {
		params := db.UpdateDMESysIDParams{
			OrganizationID: credential.OrganizationID,
			SystemID:       systemID.SystemID,
		}
		_, err = c.db.Queries().UpdateDMESysID(ctx, params)
		if err != nil {
			return fmt.Errorf("failed to update DME system ID in database: %w", err)
		}
	}

	_, err = c.db.Queries().UpdateDMEToken(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update DME token in database: %w", err)
	}

	return nil
}

// getToken authenticates with the new API version
func (c *Client) getToken(ctx context.Context, username, password string) (string, string, time.Time, []SystemID, error) {
	reqBody, err := json.Marshal(map[string]string{
		"UserName": username,
		"Password": password,
	})
	if err != nil {
		return "", "", time.Time{}, nil, fmt.Errorf("failed to marshal auth request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.AuthURL+"/token", bytes.NewReader(reqBody))
	if err != nil {
		return "", "", time.Time{}, nil, fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", "", time.Time{}, nil, fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", "", time.Time{}, nil, fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	var authResp struct {
		AuthToken            string `json:"authToken"`
		RefreshToken         string `json:"refreshToken"`
		AuthExpiration       string `json:"authExpiration"`
		AvailableConnections []struct {
			Name      string `json:"name"`
			SystemID  string `json:"systemId"`
			IsDefault bool   `json:"isDefault"`
		} `json:"availableConnections"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", "", time.Time{}, nil, fmt.Errorf("failed to decode auth response: %w", err)
	}

	expiration, err := time.Parse(time.RFC3339, authResp.AuthExpiration)
	if err != nil {
		return "", "", time.Time{}, nil, fmt.Errorf("failed to parse expiration time: %w", err)
	}

	var systemIDs []SystemID
	for _, conn := range authResp.AvailableConnections {
		systemIDs = append(systemIDs, SystemID{
			Name:      conn.Name,
			SystemID:  conn.SystemID,
			IsDefault: conn.IsDefault,
		})
	}

	return authResp.AuthToken, authResp.RefreshToken, expiration, systemIDs, nil
}

// getOldAPIToken authenticates with the old API version
func (c *Client) getOldAPIToken(ctx context.Context, username, password string) (string, time.Time, error) {
	data := "grant_type=password&username=" + username + "&password=" + password

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.AuthURL+"/token", bytes.NewBufferString(data))
	if err != nil {
		return "", time.Time{}, fmt.Errorf("failed to create auth request: %w", err)
	}
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("auth request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", time.Time{}, fmt.Errorf("authentication failed with status: %d", resp.StatusCode)
	}

	var authResp struct {
		AccessToken string `json:"access_token"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&authResp); err != nil {
		return "", time.Time{}, fmt.Errorf("failed to decode auth response: %w", err)
	}

	expiration := time.Now().Add(time.Duration(authResp.ExpiresIn) * time.Second)

	return authResp.AccessToken, expiration, nil
}

// refreshAuthToken refreshes the authentication token for an organization
func (c *Client) refreshAuthToken(ctx context.Context, organizationID uuid.UUID) error {
	// Get credentials from database
	credential, err := c.db.Queries().GetDMECredentialsByOrgID(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("failed to get DME credentials: %w", err)
	}

	if credential.IsOldApi != nil && *credential.IsOldApi {
		// Old API doesn't support refresh, so re-authenticate
		return c.authenticateOrg(ctx, credential)
	}

	if credential.AccessToken == nil || credential.RefreshToken == nil {
		return fmt.Errorf("missing access or refresh token for organization %s", organizationID)
	}

	reqBody, err := json.Marshal(map[string]string{
		"expiredAuthToken": *credential.AccessToken,
		"refreshToken":     *credential.RefreshToken,
	})
	if err != nil {
		return fmt.Errorf("failed to marshal refresh request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.config.AuthURL+"/token/refresh", bytes.NewReader(reqBody))
	if err != nil {
		return fmt.Errorf("failed to create refresh request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("refresh request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		// If refresh failed, try to re-authenticate
		return c.authenticateOrg(ctx, credential)
	}

	var refreshResp struct {
		AuthToken      string `json:"authToken"`
		RefreshToken   string `json:"refreshToken"`
		AuthExpiration string `json:"authExpiration"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&refreshResp); err != nil {
		return fmt.Errorf("failed to decode refresh response: %w", err)
	}

	expiration, err := time.Parse(time.RFC3339, refreshResp.AuthExpiration)
	if err != nil {
		return fmt.Errorf("failed to parse expiration time: %w", err)
	}

	// Update token in database
	expiryTimestamp := pgtype.Timestamp{Time: expiration, Valid: true}

	params := db.UpdateDMETokenParams{
		OrganizationID: credential.OrganizationID,
		AccessToken:    &refreshResp.AuthToken,
		RefreshToken:   &refreshResp.RefreshToken,
		ExpiryDate:     expiryTimestamp,
	}

	_, err = c.db.Queries().UpdateDMEToken(ctx, params)
	if err != nil {
		return fmt.Errorf("failed to update DME token in database: %w", err)
	}

	return nil
}

// ensureValidToken ensures the auth token is still valid, refreshing if needed
func (c *Client) ensureValidToken(ctx context.Context, organizationID uuid.UUID) error {
	// Get the token from the database
	credential, err := c.db.Queries().GetDMECredentialsByOrgID(ctx, organizationID)
	if err != nil {
		return fmt.Errorf("failed to get DME credentials: %w", err)
	}

	// Check if token exists and is valid
	if credential.AccessToken == nil ||
		credential.ExpiryDate.Time.Before(time.Now().Add(30*time.Second)) {
		return c.refreshAuthToken(ctx, organizationID)
	}
	return nil
}

// getHeaders returns the headers needed for API requests
func (c *Client) getHeaders(ctx context.Context, organizationID uuid.UUID, systemID string) (http.Header, error) {
	// Get credentials from database
	credential, err := c.db.Queries().GetDMECredentialsByOrgID(ctx, organizationID)
	if err != nil {
		return nil, fmt.Errorf("failed to get DME credentials: %w", err)
	}

	if credential.AccessToken == nil {
		return nil, fmt.Errorf("missing access token for organization %s", organizationID)
	}

	headers := http.Header{}
	headers.Set("Authorization", "Bearer "+*credential.AccessToken)
	headers.Set("Content-Type", "application/json")

	if systemID != "" && (credential.IsOldApi == nil || !*credential.IsOldApi) {
		headers.Set("X-DM_SYSTEM_ID", systemID)
	}

	return headers, nil
}

// DoRequest makes an HTTP request to the DME API with the specified organization and system ID
func (c *Client) DoRequest(ctx context.Context, method, endpoint string, body interface{}, organizationID uuid.UUID, systemID string) (*http.Response, error) {
	if err := c.ensureValidToken(ctx, organizationID); err != nil {
		return nil, fmt.Errorf("failed to ensure valid token: %w", err)
	}

	var reqBody []byte
	var err error

	if body != nil {
		reqBody, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
	}

	url := c.config.BaseURL
	if c.config.APIVersion != "" {
		url += c.config.APIVersion
	}
	url += endpoint

	req, err := http.NewRequestWithContext(ctx, method, url, bytes.NewReader(reqBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Set headers
	headers, err := c.getHeaders(ctx, organizationID, systemID)
	if err != nil {
		return nil, err
	}

	for key, values := range headers {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	return c.httpClient.Do(req)
}

// DoJSONRequest makes an HTTP request and decodes the JSON response
func (c *Client) DoJSONRequest(ctx context.Context, method, endpoint string, body, result interface{}, organizationID uuid.UUID, systemID string) error {
	resp, err := c.DoRequest(ctx, method, endpoint, body, organizationID, systemID)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("request failed with status: %d", resp.StatusCode)
	}

	if result != nil {
		if err := json.NewDecoder(resp.Body).Decode(result); err != nil {
			return fmt.Errorf("failed to decode response: %w", err)
		}
	}

	return nil
}

// NewClientFromConfig creates a new DME API client from app config
func NewClientFromConfig(cfg *config.Config, logger *logger.Logger, db pg.DBService) *Client {
	clientCfg := &ClientConfig{
		BaseURL:    cfg.DME.BaseURL,
		AuthURL:    cfg.DME.AuthURL,
		APIVersion: cfg.DME.APIVersion,
		IsOldAPI:   cfg.DME.IsOldAPI,
		Logger:     logger,
		DB:         db,
	}

	return NewClient(clientCfg)
}
