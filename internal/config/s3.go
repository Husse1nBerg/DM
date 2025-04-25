package config

import (
	"os"

	"github.com/dockworks/dm-web-backend/pkg/s3"
)

// LoadS3Config loads S3 configuration from environment variables
func LoadS3Config() s3.S3Config {
	return s3.S3Config{
		Region:          getEnvOrDefault("AWS_REGION", "us-east-1"),
		AccessKeyID:     getEnvOrDefault("AWS_ACCESS_KEY_ID", ""),
		SecretAccessKey: getEnvOrDefault("AWS_SECRET_ACCESS_KEY", ""),
		Bucket:          getEnvOrDefault("AWS_S3_BUCKET", ""),
		BaseURL:         getEnvOrDefault("AWS_S3_BASE_URL", ""),
	}
}

// getEnvOrDefault gets an environment variable or returns a default value
func getEnvOrDefault(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
