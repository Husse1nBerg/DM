package config

import "os"

// DMEConfig holds the configuration for DockMaster API
type DMEConfig struct {
	BaseURL string
	AuthURL string
}

// LoadDMEConfig loads the DME configuration from environment variables
func LoadDMEConfig() DMEConfig {
	baseURL := os.Getenv("DME_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.dmeapi.dev"
	}

	authURL := os.Getenv("DME_AUTH_URL")
	if authURL == "" {
		authURL = "https://auth.dmeapi.dev"
	}

	return DMEConfig{
		BaseURL: baseURL,
		AuthURL: authURL,
	}
}
