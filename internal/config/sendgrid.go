package config

import (
	"os"
)

// SendGridConfig holds SendGrid API settings and template configurations
type SendGridConfig struct {
	APIKey       string            // SendGrid API key
	FromEmail    string            // Default sender email
	FromName     string            // Default sender name
	TemplatesMap map[string]string // Map of template names to SendGrid template IDs
}

// LoadSendGridConfig loads SendGrid configuration from environment variables
func LoadSendGridConfig() SendGridConfig {
	// Initialize the template map
	templatesMap := make(map[string]string)

	// Add known templates to the map
	// Example: templatesMap["welcome"] = os.Getenv("SENDGRID_TEMPLATE_WELCOME")
	templatesMap["password_reset"] = os.Getenv("SENDGRID_TEMPLATE_PASSWORD_RESET")
	templatesMap["welcome"] = os.Getenv("SENDGRID_TEMPLATE_WELCOME")

	return SendGridConfig{
		APIKey:       os.Getenv("SENDGRID_API_KEY"),
		FromEmail:    getEnvOrDefault("SENDGRID_FROM_EMAIL", "no-reply@dockmaster.com"),
		FromName:     getEnvOrDefault("SENDGRID_FROM_NAME", "Dockmaster"),
		TemplatesMap: templatesMap,
	}
}
