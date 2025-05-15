package config

import (
	"os"
)

// TelgorithmConfig holds Telgorithm API settings for SMS services
type TelgorithmConfig struct {
	BaseURL    string // Telgorithm API URL
	Username   string // Telgorithm username for basic auth
	Password   string // Telgorithm password for basic auth
	FromNumber string // Default sender phone number
}

// LoadTelgorithmConfig loads Telgorithm configuration from environment variables
func LoadTelgorithmConfig() TelgorithmConfig {
	return TelgorithmConfig{
		BaseURL:    getEnvOrDefault("TELGORITHM_BASE_URL", "https://api.telgorithm.com/v1"),
		Username:   os.Getenv("TELGORITHM_USERNAME"),
		Password:   os.Getenv("TELGORITHM_PASSWORD"),
		FromNumber: os.Getenv("TELGORITHM_FROM_NUMBER"),
	}
}
