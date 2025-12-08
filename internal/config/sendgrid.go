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
	templatesMap["invite_customer"] = os.Getenv("SENDGRID_TEMPLATE_INVITE_CUSTOMER")
	templatesMap["message_external"] = os.Getenv("SENDGRID_TEMPLATE_MESSAGE_EXTERNAL")
	templatesMap["assigned_to_marina"] = os.Getenv("SENDGRID_TEMPLATE_ASSIGNED_TO_MARINA")
	templatesMap["esign_submission"] = os.Getenv("SENDGRID_TEMPLATE_ESIGN_SUBMISSION")
	templatesMap["notification"] = os.Getenv("SENDGRID_TEMPLATE_NOTIFICATION")
	templatesMap["payment_link"] = os.Getenv("SENDGRID_TEMPLATE_PAYMENT_LINK")
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
		templatesMap["invite"] = "d-4d89d4122eeb473cbc2a59b373976934"
	}
	if templatesMap["invite_customer"] == "" {
		templatesMap["invite_customer"] = "d-c9f51b77474a4d3bbf8e3df07af225a2"
	}
	if templatesMap["message_external"] == "" {
		templatesMap["message_external"] = "d-6c7a11bc1f3147018986af40b98ca69a"
	}
	if templatesMap["assigned_to_marina"] == "" {
		templatesMap["assigned_to_marina"] = "d-a79d23ec7a564423bc15f844c9f25e23"
	}
	if templatesMap["esign_submission"] == "" {
		templatesMap["esign_submission"] = "d-b14c44271dbb4893a34730bc59cb3c62"
	}
	if templatesMap["payment_link"] == "" {
		templatesMap["payment_link"] = "d-d02373e4593b4cdaab42c1022abb189f"
	}
	if templatesMap["notification"] == "" {
		templatesMap["notification"] = "d-e6168ddc800045b6abf536dfaa5681d9"
	}
	return SendGridConfig{
		APIKey:       os.Getenv("SENDGRID_API_KEY"),
		FromEmail:    EnvOrDefault("SENDGRID_FROM_EMAIL", "no-reply@dockmaster.com"),
		FromName:     EnvOrDefault("SENDGRID_FROM_NAME", "Dockmaster"),
		TemplatesMap: templatesMap,
	}
}
