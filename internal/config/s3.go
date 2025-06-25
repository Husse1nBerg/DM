package config

import (
	"github.com/dockworks/dm-web-backend/pkg/s3"
)

// LoadS3Config loads S3 configuration from environment variables
func LoadS3Config() s3.S3Config {
	return s3.S3Config{
		Region:          EnvOrDefault("AWS_REGION", "us-east-1"),
		AccessKeyID:     EnvOrDefault("AWS_ACCESS_KEY_ID", ""),
		SecretAccessKey: EnvOrDefault("AWS_SECRET_ACCESS_KEY", ""),
		Bucket:          EnvOrDefault("AWS_S3_BUCKET", ""),
		BaseURL:         EnvOrDefault("AWS_S3_BASE_URL", ""),
	}
}

// LoadDocumentStorageConfig loads document storage S3 configuration from environment variables
func LoadDocumentStorageConfig() s3.S3Config {
	return s3.S3Config{
		Region:          EnvOrDefault("AWS_REGION", "us-east-1"),
		AccessKeyID:     EnvOrDefault("AWS_ACCESS_KEY_ID", ""),
		SecretAccessKey: EnvOrDefault("AWS_SECRET_ACCESS_KEY", ""),
		Bucket:          EnvOrDefault("AWS_S3_STORAGE_BUCKET", ""),
		BaseURL:         EnvOrDefault("AWS_S3_STORAGE_BASE_URL", ""),
	}
}

func LoadESignConfig() s3.S3Config {
	return s3.S3Config{
		Region:          EnvOrDefault("AWS_REGION", "us-east-1"),
		AccessKeyID:     EnvOrDefault("AWS_ACCESS_KEY_ID", ""),
		SecretAccessKey: EnvOrDefault("AWS_SECRET_ACCESS_KEY", ""),
		Bucket:          EnvOrDefault("AWS_S3_ESIGN_BUCKET", ""),
		BaseURL:         EnvOrDefault("AWS_S3_ESIGN_BASE_URL", ""),
	}
}
