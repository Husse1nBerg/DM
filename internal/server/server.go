package server

import (
	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Server struct {
	Echo   *echo.Echo
	Config *config.Config
	Logger *logger.Logger
}

func NewServer(cfg *config.Config, logger *logger.Logger) *Server {
	return &Server{
		Config: cfg,
		Echo:   echo.New(),
		Logger: logger,
	}
}

func (server *Server) Start(addr string) error {
	return server.Echo.Start(":" + addr)
}
