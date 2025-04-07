package config

import (
	"fmt"
	"os"
	"strconv"
)

type RedisConfig struct {
	Host      string `mapstructure:"Host"`
	Port      int    `mapstructure:"Port"`
	Password  string `mapstructure:"Password"`
	KeyPrefix string `mapstructure:"KeyPrefix"`
	MainDB    int    `mapstructure:"MainDB"`
	TaskDB    int    `mapstructure:"TaskDB"`
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

	// Default database indices
	mainDB := 0
	mainDBStr := os.Getenv("REDIS_DB")
	if mainDBStr != "" {
		parsedMainDB, err := strconv.Atoi(mainDBStr)
		if err == nil {
			mainDB = parsedMainDB
		}
	}

	taskDB := 1
	taskDBStr := os.Getenv("REDIS_TASK_DB")
	if taskDBStr != "" {
		parsedTaskDB, err := strconv.Atoi(taskDBStr)
		if err == nil {
			taskDB = parsedTaskDB
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

	return RedisConfig{
		Host:      host,
		Port:      port,
		Password:  os.Getenv("REDIS_PASSWORD"),
		KeyPrefix: keyPrefix,
		MainDB:    mainDB,
		TaskDB:    taskDB,
	}
}

func (a *RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%d", a.Host, a.Port)
}
