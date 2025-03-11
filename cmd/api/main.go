package main

import (
	"fmt"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/internal/server/routes"

	"github.com/dockworks/dm-web-backend/docs"
	"github.com/dockworks/dm-web-backend/pkg/logger"
)

//	@title			DockMaster Web API
//	@version		0.0.1
//	@description	This is a API Server.

//	@contact.name	Andrew Sameh
//	@contact.url	https://andrewsam.xyz
//	@contact.email	g.andrewsameh@gmail.com

//	@securityDefinitions.apiKey ApiKeyAuth
//	@in							header
//	@name						Authorization

// @BasePath	/api/v1
func main() {
	cfg := config.New()

	zlog := logger.NewLogger(cfg.Logger)
	if zlog.Zap != nil {
		defer zlog.Zap.Sync()
	}

	server := server.NewServer(cfg, zlog)
	routes.RegisterRoutes(server)

	docs.SwaggerInfo.Host = fmt.Sprintf("%s:%s", cfg.Server.Host, cfg.Server.Port)
	zlog.Zap.Infof("Service URL: http://localhost:%s/swagger/index.html", cfg.Server.Port)

	err := server.Start(cfg.Server.Port)
	if err != nil {
		zlog.Zap.Fatalf("Cannot start server: %s", err)
	}
}
