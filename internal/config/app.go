package config

import (
	"os"
	"strings"
)

type AppConfig struct {
	AdminEmail           string `mapstructure:"AdminEmail"`
	AdminPassword        string `mapstructure:"AdminPassword"`
	FrontendBaseURL      string `mapstructure:"FrontendBaseURL"`
	InvitationRoute      string `mapstructure:"InvitationRoute"`
	CustomerIntakeRoute  string `mapstructure:"CustomerIntakeRoute"`
	PasswordResetRoute   string `mapstructure:"PasswordResetRoute"`
	TermsConditionsRoute string `mapstructure:"TermsConditionsRoute"`
}

func LoadAppConfig() AppConfig {
	frontendBaseURL := os.Getenv("FRONTEND_BASE_URL")
	invitationRoute := EnvOrDefault("INVITATION_ROUTE", "auth/customer-portal-access")
	customerIntakeRoute := EnvOrDefault("CUSTOMER_INTAKE_ROUTE", "auth/customer-intake")
	passwordResetRoute := EnvOrDefault("PASSWORD_RESET_ROUTE", "auth/password-reset")
	termsConditionsRoute := EnvOrDefault("TERMS_CONDITIONS_ROUTE", "terms-conditions")
	if frontendBaseURL == "" {
		frontendBaseURL = "https://dmweb-dev.dockmaster.com"
	}

	return AppConfig{
		AdminEmail:           os.Getenv("ADMIN_EMAIL"),
		AdminPassword:        os.Getenv("ADMIN_PASSWORD"),
		FrontendBaseURL:      frontendBaseURL,
		InvitationRoute:      invitationRoute,
		CustomerIntakeRoute:  customerIntakeRoute,
		PasswordResetRoute:   passwordResetRoute,
		TermsConditionsRoute: termsConditionsRoute,
	}
}

func (c *AppConfig) RemoveSlashes(url string) string {
	return strings.TrimSuffix(strings.TrimPrefix(url, "/"), "/")
}

func (c *AppConfig) InvitationURL() string {
	return c.RemoveSlashes(c.FrontendBaseURL) + "/" + c.RemoveSlashes(c.InvitationRoute)
}

func (c *AppConfig) PasswordResetURL() string {
	return c.RemoveSlashes(c.FrontendBaseURL) + "/" + c.RemoveSlashes(c.PasswordResetRoute)
}

func (c *AppConfig) TermsConditionsURL() string {
	return c.RemoveSlashes(c.FrontendBaseURL) + "/" + c.RemoveSlashes(c.TermsConditionsRoute)
}
