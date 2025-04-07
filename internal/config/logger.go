package config

import (
	"os"
	"strconv"
)

type LoggerConfig struct {
	Format    string `mapstructure:"Format"`
	Level     string `mapstructure:"Level"`
	Directory string `mapstructure:"Directory"`
	Name      string `mapstructure:"Name"`
	Local     bool   `mapstructure:"Local"`
}

func LoadLoggerConfig() LoggerConfig {
	local, err := strconv.ParseBool(os.Getenv("LOGS_LOCAL"))
	if err != nil {
		local = false
	}

	format := os.Getenv("LOGS_FORMAT")
	if format == "" {
		format = "json"
	}

	level := os.Getenv("LOGS_LEVEL")
	if level == "" {
		level = "info"
	}

	directory := os.Getenv("LOGS_DIRECTORY")
	if directory == "" {
		directory = "logs"
	}

	name := os.Getenv("LOGS_NAME")
	if name == "" {
		name = "echo"
	}

	return LoggerConfig{
		Format:    format,
		Level:     level,
		Directory: directory,
		Name:      name,
		Local:     local,
	}
}
