package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

type RedisConfig struct {
	Host       string `mapstructure:"Host"`
	Port       int    `mapstructure:"Port"`
	Username   string `mapstructure:"Username"`
	Password   string `mapstructure:"Password"`
	KeyPrefix  string `mapstructure:"KeyPrefix"`
	TLSEnabled bool   `mapstructure:"TLSEnabled"`
}

func LoadRedisConfig() RedisConfig {
	// Default port for Redis
	portStr := os.Getenv("REDIS_PORT")
	port := 6379 // Default Redis port
	if portStr != "" {
		parsedPort, err := strconv.Atoi(portStr)
		if err == nil {
			port = parsedPort
		}
	}

	// Default host and prefix
	host := os.Getenv("REDIS_HOST")
	if host == "" {
		host = "localhost"
	}

	keyPrefix := os.Getenv("REDIS_KEY_PREFIX")
	if keyPrefix == "" {
		keyPrefix = "dm:"
	}

	// TLS configuration
	tlsEnabled := true // Default to true for AWS ElastiCache
	tlsEnabledStr := os.Getenv("REDIS_TLS_ENABLED")
	if tlsEnabledStr != "" {
		// Only disable TLS if explicitly set to false or 0
		if strings.ToLower(tlsEnabledStr) == "false" || tlsEnabledStr == "0" {
			tlsEnabled = false
		}
	}

	return RedisConfig{
		Host:       host,
		Port:       port,
		Username:   os.Getenv("REDIS_USERNAME"),
		Password:   os.Getenv("REDIS_PASSWORD"),
		KeyPrefix:  keyPrefix,
		TLSEnabled: tlsEnabled,
	}
}

func (a *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}
