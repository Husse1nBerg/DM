package config

import (
	"fmt"
	"os"
)

type DBConfig struct {
	User     string
	Password string
	Driver   string
	Name     string
	Host     string
	Port     string
	Schema   string
}

func LoadDBConfig() DBConfig {
	port := os.Getenv("DB_PORT")
	if port == "" {
		port = "5432"
	}

	schema := os.Getenv("DB_SCHEMA")
	if schema == "" {
		schema = "public"
	}

	return DBConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     port,
		Schema:   schema,
	}
}

func LoadTestDBConfig() DBConfig {
	return DBConfig{
		User:     os.Getenv("TEST_DB_USER"),
		Password: os.Getenv("TEST_DB_PASSWORD"),
		Name:     os.Getenv("TEST_DB_NAME"),
		Host:     os.Getenv("TEST_DB_HOST"),
		Port:     os.Getenv("TEST_DB_PORT"),
		Schema:   os.Getenv("TEST_DB_SCHEMA"),
	}
}

func (a *DBConfig) Addr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", a.User, a.Password, a.Host, a.Port, a.Name, a.Schema)
}
