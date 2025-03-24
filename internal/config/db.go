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
	return DBConfig{
		User:     os.Getenv("DB_USER"),
		Password: os.Getenv("DB_PASSWORD"),
		Name:     os.Getenv("DB_NAME"),
		Host:     os.Getenv("DB_HOST"),
		Port:     os.Getenv("DB_PORT"),
		Schema:   os.Getenv("DB_SCHEMA"),
	}
}

func (a *DBConfig) Addr() string {
	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=%s", a.User, a.Password, a.Host, a.Port, a.Name, a.Schema)
}
