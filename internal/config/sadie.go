package config

import (
	"fmt"
	"os"
	"strings"
)

// SadieConfig holds the configuration for SADIE Core API
type SadieConfig struct {
	BaseURL        string
	APIKey         string
	TenantID       string
	ClientSecret   string
	WebhookURL     string
}

// LoadSadieConfig loads the SADIE configuration from environment variables
func LoadSadieConfig() SadieConfig {
	baseURL := os.Getenv("SADIE_BASE_URL")
	if baseURL == "" {
		baseURL = "https://core-api.heysadie.ai"
	}

	apiKey := os.Getenv("SADIE_API_KEY")
	tenantID := os.Getenv("SADIE_TENANT_ID")
	clientSecret := os.Getenv("SADIE_CLIENT_SECRET")
	webhookURL := os.Getenv("SADIE_WEBHOOK_URL")
	
	// Use defaults from provided credentials if not set
	if apiKey == "" {
		apiKey = "eyJ0aWQiOiIwMTk1YzkyMC1iYTY0LTczNzgtYTEwYy02YTdkNjM0OTZjMWYiLCJqdGkiOiI2YzkxZGVmODIyNmU5ZjgxMzk2MDg4Y2Y3NTc1ZGM2ZiJ9.8c253a5fe8de303c9e7e8466ba7ece0451a138a84d9217e9fc44a01d3ef9f7a1"
	}
	if tenantID == "" {
		tenantID = "0195c920-ba64-7378-a10c-6a7d63496c1f"
	}
	if clientSecret == "" {
		clientSecret = "dev_dockmaster_0195c925-ff69-7298-81a7-80f9348b52e6"
	}

	// If webhook URL is not set, try to construct it from server config
	if webhookURL == "" {
		serverHost := os.Getenv("HOST")
		serverPort := os.Getenv("PORT")
		env := os.Getenv("ENV")
		
		if serverHost == "" {
			serverHost = "localhost"
		}
		if serverPort == "" {
			serverPort = "8080"
		}
		
		// Construct webhook URL based on environment
		if env == "production" || env == "prod" {
			// For production, use https and the host (assuming it's a full domain)
			if strings.HasPrefix(serverHost, "http") {
				webhookURL = strings.TrimSuffix(serverHost, "/") + "/webhooks"
			} else {
				webhookURL = fmt.Sprintf("https://%s/webhooks", serverHost)
			}
		} else {
			// For dev/local, use http with port
			if strings.HasPrefix(serverHost, "http") {
				webhookURL = strings.TrimSuffix(serverHost, "/") + "/webhooks"
			} else {
				webhookURL = fmt.Sprintf("http://%s:%s/webhooks", serverHost, serverPort)
			}
		}
	}

	return SadieConfig{
		BaseURL:      baseURL,
		APIKey:       apiKey,
		TenantID:     tenantID,
		ClientSecret: clientSecret,
		WebhookURL:   webhookURL,
	}
}

