package config

import (
	"os"
)

type AppConfig struct {
	AdminEmail           string `mapstructure:"AdminEmail"`
	AdminPassword        string `mapstructure:"AdminPassword"`
	FrontendBaseURL      string `mapstructure:"FrontendBaseURL"`
	InvitationRoute      string `mapstructure:"InvitationRoute"`
	PasswordResetRoute   string `mapstructure:"PasswordResetRoute"`
	TermsConditionsRoute string `mapstructure:"TermsConditionsRoute"`
}

func LoadAppConfig() AppConfig {
	frontendBaseURL := os.Getenv("FRONTEND_BASE_URL")
	if frontendBaseURL == "" {
		frontendBaseURL = "https://dmweb-dev.dockmaster.com"
	}

	return AppConfig{
		AdminEmail:           os.Getenv("ADMIN_EMAIL"),
		AdminPassword:        os.Getenv("ADMIN_PASSWORD"),
		FrontendBaseURL:      frontendBaseURL,
		InvitationRoute:      os.Getenv("INVITATION_ROUTE"),
		PasswordResetRoute:   os.Getenv("PASSWORD_RESET_ROUTE"),
		TermsConditionsRoute: os.Getenv("TERMS_CONDITIONS_ROUTE"),
	}
}
