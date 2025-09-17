package config

import (
	"os"
)

// AdyenConfig holds Adyen API settings and configuration
type AdyenConfig struct {
	APIKey                string   // Adyen API key
	MerchantAccount       string   // Merchant account name
	Environment           string   // Test or Live
	LiveEndpointURL       string   // Live endpoint URL prefix (only used in live environment)
	WebhookHMACKey        string   // HMAC key for webhook verification
	DefaultStoreID        string   // Default store ID for payments
	AllowedPaymentMethods []string // List of allowed payment methods
}

// LoadAdyenConfig loads Adyen configuration from environment variables
func LoadAdyenConfig() AdyenConfig {
	environment := EnvOrDefault("ADYEN_ENVIRONMENT", "test")

	return AdyenConfig{
		APIKey:                os.Getenv("ADYEN_API_KEY"),
		MerchantAccount:       EnvOrDefault("ADYEN_MERCHANT_ACCOUNT", ""),
		Environment:           environment,
		LiveEndpointURL:       EnvOrDefault("ADYEN_LIVE_ENDPOINT_URL", ""),
		WebhookHMACKey:        os.Getenv("ADYEN_WEBHOOK_HMAC_KEY"),
		DefaultStoreID:        EnvOrDefault("ADYEN_DEFAULT_STORE_ID", ""),
		AllowedPaymentMethods: []string{"visa", "mc", "amex", "discover", "ach"},
	}
}
