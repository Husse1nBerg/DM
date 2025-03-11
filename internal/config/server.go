package config

import (
	"os"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

type ServerConfig struct {
	Host       string
	Port       string
	Env        string
	Validator  echo.Validator
	Binder     echo.Binder
	CORSConfig middleware.CORSConfig
}

func GetEchoLogConfig(cfg *Config) middleware.LoggerConfig {
	echoLogConf := middleware.DefaultLoggerConfig
	echoLogConf.CustomTimeFormat = time.RFC3339
	// echoLogConf.Format = fmt.Sprintln(`{"level":"info","source":"echo","id":"${id}","mt":"${method}","uri":"${uri}","st":${status},"e":"${error}","lc":"${latency_human}","ts":"${time_custom}"}`)
	return echoLogConf
}

func LoadServerConfig() ServerConfig {
	return ServerConfig{
		Host:       os.Getenv("HOST"),
		Port:       os.Getenv("PORT"),
		Env:        os.Getenv("ENV"),
		CORSConfig: middleware.DefaultCORSConfig,
	}
}
