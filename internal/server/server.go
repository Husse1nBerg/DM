package server

import (
	"github.com/dockworks/dm-web-backend/internal/config"
	db "github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/labstack/echo/v4"
)

type Server struct {
	Echo   *echo.Echo
	Config *config.Config
	DB     db.DBService
	Logger *logger.Logger
	DME    *dme.Client
}

func NewServer(cfg *config.Config, logger *logger.Logger) *Server {
	dbConn := db.NewConnection(&cfg.DB)

	return &Server{
		Config: cfg,
		Echo:   echo.New(),
		DB:     dbConn,
		Logger: logger,
		DME:    dme.NewClientFromConfig(cfg, logger, dbConn),
	}
}

func (server *Server) Start(addr string) error {
	return server.Echo.Start(":" + addr)
}
