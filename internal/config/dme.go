package config

import "os"

// DMEConfig holds the configuration for DockMaster API
type DMEConfig struct {
	BaseURL    string
	AuthURL    string
	APIVersion string
	IsOldAPI   bool
}

// LoadDMEConfig loads the DME configuration from environment variables
func LoadDMEConfig() DMEConfig {
	baseURL := os.Getenv("DME_BASE_URL")
	if baseURL == "" {
		baseURL = "https://api.dockmaster.com"
	}

	authURL := os.Getenv("DME_AUTH_URL")
	if authURL == "" {
		authURL = "https://auth.dockmaster.com"
	}

	apiVersion := os.Getenv("DME_API_VERSION")
	if apiVersion == "" {
		apiVersion = "/api/v1"
	}

	isOldAPI := os.Getenv("DME_IS_OLD_API") == "true"

	return DMEConfig{
		BaseURL:    baseURL,
		AuthURL:    authURL,
		APIVersion: apiVersion,
		IsOldAPI:   isOldAPI,
	}
}
