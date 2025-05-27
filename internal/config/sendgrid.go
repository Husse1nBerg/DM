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
	templatesMap["message"] = os.Getenv("SENDGRID_TEMPLATE_MESSAGE")
	templatesMap["invite"] = os.Getenv("SENDGRID_TEMPLATE_INVITE")
	templatesMap["message_external"] = os.Getenv("SENDGRID_TEMPLATE_MESSAGE_EXTERNAL")
	if templatesMap["message"] == "" {
		templatesMap["message"] = "d-2a2afe50d47d417c99019692dc20079a"
	}
	if templatesMap["password_reset"] == "" {
		templatesMap["password_reset"] = "d-61fbca26090a4668b7279ab1bc2c8980"
	}
	if templatesMap["welcome"] == "" {
		templatesMap["welcome"] = "d-d823f4dd24504a7a9b3bbd7ae96e4510"
	}
	if templatesMap["invite"] == "" {
		templatesMap["invite"] = "d-c9f51b77474a4d3bbf8e3df07af225a2"
	}
	if templatesMap["message_external"] == "" {
		templatesMap["message_external"] = "d-6c7a11bc1f3147018986af40b98ca69a"
	}

	return SendGridConfig{
		APIKey:       os.Getenv("SENDGRID_API_KEY"),
		FromEmail:    getEnvOrDefault("SENDGRID_FROM_EMAIL", "no-reply@dockmaster.com"),
		FromName:     getEnvOrDefault("SENDGRID_FROM_NAME", "Dockmaster"),
		TemplatesMap: templatesMap,
	}
}
