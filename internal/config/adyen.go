package config

import (
	"os"
)

type AdyenConfig struct {
	APIKey          string
	Environment     string
	MerchantAccount string
	ClientKey       string
	HMACKey         string
}

func LoadAdyenConfig() AdyenConfig {
	return AdyenConfig{
		APIKey:          os.Getenv("ADYEN_API_KEY"),
		Environment:     EnvOrDefault("ADYEN_ENVIRONMENT", "test"),
		MerchantAccount: os.Getenv("ADYEN_MERCHANT_ACCOUNT"),
		ClientKey:       os.Getenv("ADYEN_CLIENT_KEY"),
		HMACKey:         os.Getenv("ADYEN_HMAC_KEY"),
	}
}
