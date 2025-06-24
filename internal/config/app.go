package config

import (
	"os"
	"strings"
)

type AppConfig struct {
	AdminEmail              string `mapstructure:"AdminEmail"`
	AdminPassword           string `mapstructure:"AdminPassword"`
	FrontendBaseURL         string `mapstructure:"FrontendBaseURL"`
	InvitationRoute         string `mapstructure:"InvitationRoute"`
	InvitationCustomerRoute string `mapstructure:"InvitationCustomerRoute"`
	CustomerIntakeRoute     string `mapstructure:"CustomerIntakeRoute"`
	PasswordResetRoute      string `mapstructure:"PasswordResetRoute"`
	TermsConditionsRoute    string `mapstructure:"TermsConditionsRoute"`
}

func LoadAppConfig() AppConfig {
	frontendBaseURL := EnvOrDefault("FRONTEND_BASE_URL", "https://app.dockmaster.com")
	invitationRoute := EnvOrDefault("INVITATION_ROUTE", "auth/invitation")
	invitationCustomerRoute := EnvOrDefault("INVITATION_CUSTOMER_ROUTE", "auth/invitation")
	customerIntakeRoute := EnvOrDefault("CUSTOMER_INTAKE_ROUTE", "auth/customer-intake")
	passwordResetRoute := EnvOrDefault("PASSWORD_RESET_ROUTE", "auth/password-reset")
	termsConditionsRoute := EnvOrDefault("TERMS_CONDITIONS_ROUTE", "terms-conditions")
	if frontendBaseURL == "" {
		switch strings.ToLower(os.Getenv("ENV")) {
		case "production", "prod":
			frontendBaseURL = "https://app.dockmaster.com"
		case "beta":
			frontendBaseURL = "https://dmweb-beta.dockmaster.com"
		case "development", "dev":
			frontendBaseURL = "https://dmweb-dev.dockmaster.com"
		}
	}

	return AppConfig{
		AdminEmail:              os.Getenv("ADMIN_EMAIL"),
		AdminPassword:           os.Getenv("ADMIN_PASSWORD"),
		FrontendBaseURL:         frontendBaseURL,
		InvitationRoute:         invitationRoute,
		InvitationCustomerRoute: invitationCustomerRoute,
		CustomerIntakeRoute:     customerIntakeRoute,
		PasswordResetRoute:      passwordResetRoute,
		TermsConditionsRoute:    termsConditionsRoute,
	}
}

func (c *AppConfig) RemoveSlashes(url string) string {
	return strings.TrimSuffix(strings.TrimPrefix(url, "/"), "/")
}

func (c *AppConfig) InvitationURL() string {
	return c.RemoveSlashes(c.FrontendBaseURL) + "/" + c.RemoveSlashes(c.InvitationRoute)
}

func (c *AppConfig) InvitationCustomerURL() string {
	return c.RemoveSlashes(c.FrontendBaseURL) + "/" + c.RemoveSlashes(c.InvitationCustomerRoute)
}

func (c *AppConfig) PasswordResetURL() string {
	return c.RemoveSlashes(c.FrontendBaseURL) + "/" + c.RemoveSlashes(c.PasswordResetRoute)
}

func (c *AppConfig) TermsConditionsURL() string {
	return c.RemoveSlashes(c.FrontendBaseURL) + "/" + c.RemoveSlashes(c.TermsConditionsRoute)
}

func (c *AppConfig) CustomerIntakeURL() string {
	return c.RemoveSlashes(c.FrontendBaseURL) + "/" + c.RemoveSlashes(c.CustomerIntakeRoute)
}

func (c *AppConfig) HomeURL() string {
	return c.RemoveSlashes(c.FrontendBaseURL)
}
