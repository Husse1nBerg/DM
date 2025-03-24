package config

import (
	"os"
)

type AppConfig struct {
	AdminEmail    string `mapstructure:"AdminEmail"`
	AdminPassword string `mapstructure:"AdminPassword"`
}

func LoadAppConfig() AppConfig {

	return AppConfig{
		AdminEmail:    os.Getenv("ADMIN_EMAIL"),
		AdminPassword: os.Getenv("ADMIN_PASSWORD"),
	}
}
