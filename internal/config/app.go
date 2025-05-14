package config

import (
	"os"
)

type AppConfig struct {
	AdminEmail      string `mapstructure:"AdminEmail"`
	AdminPassword   string `mapstructure:"AdminPassword"`
	FrontendBaseURL string `mapstructure:"FrontendBaseURL"`
}

func LoadAppConfig() AppConfig {
	frontendBaseURL := os.Getenv("FRONTEND_BASE_URL")
	if frontendBaseURL == "" {
		frontendBaseURL = "https://dmweb-dev.dockmaster.com"
	}

	return AppConfig{
		AdminEmail:      os.Getenv("ADMIN_EMAIL"),
		AdminPassword:   os.Getenv("ADMIN_PASSWORD"),
		FrontendBaseURL: frontendBaseURL,
	}
}
