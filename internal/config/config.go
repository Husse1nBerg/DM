package config

import (
	"log"

	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/joho/godotenv"
)

type Config struct {
	Server ServerConfig
	Logger LoggerConfig
	Auth   AuthConfig
	DB     DBConfig
	TestDB DBConfig
	App    AppConfig
	Redis  RedisConfig
	DME    DMEConfig
	S3     s3.S3Config
}

func New() *Config {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
		log.Println(err)
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
		TestDB: LoadTestDBConfig(),
		App:    LoadAppConfig(),
		Redis:  LoadRedisConfig(),
		DME:    LoadDMEConfig(),
		S3:     LoadS3Config(),
	}
}
