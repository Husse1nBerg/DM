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
}

func New() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	return &Config{
		Server: LoadServerConfig(),
		Logger: LoadLoggerConfig(),
		Auth:   LoadAuthConfig(),
		DB:     LoadDBConfig(),
	}
}
