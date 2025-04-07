package config

import (
	"log"

	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	Logger LoggerConfig
	Auth   AuthConfig
	DB     DBConfig
	App    AppConfig
	Redis  RedisConfig
}

func New() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	// Load all configurations
	dbConfig := LoadDBConfig()

	// Validate DB configuration - these fields are required
	if dbConfig.Host == "" || dbConfig.User == "" ||
		dbConfig.Password == "" || dbConfig.Name == "" {
		log.Fatalf("Database configuration missing. Please check your .env file for DB_HOST, DB_USER, DB_PASSWORD, and DB_NAME")
	}

	return &Config{
		Server: LoadServerConfig(),
		Logger: LoadLoggerConfig(),
		Auth:   LoadAuthConfig(),
		DB:     dbConfig,
		App:    LoadAppConfig(),
		Redis:  LoadRedisConfig(),
	}
}
