package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/internal/server/routes"

	"github.com/dockworks/dm-web-backend/docs"
	"github.com/dockworks/dm-web-backend/pkg/logger"

	_ "github.com/lib/pq"
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

	dbUser := os.Getenv("DB_USER")
	dbPassword := os.Getenv("DB_PASSWORD")
	dbHost := os.Getenv("DB_HOST")
	dbPort := os.Getenv("DB_PORT")
	dbName := os.Getenv("DB_NAME")

	connStr := fmt.Sprintf("postgresql://%s:%s@%s:%s/%s?sslmode=disable",
		dbUser, dbPassword, dbHost, dbPort, dbName)

	dbConn, err := sql.Open("postgres", connStr)
	if err != nil {
		zlog.Zap.Fatalf("Failed to connect to the database: %s", err)
	}
	defer dbConn.Close()

	// Initialize the Queries object with the db connection
	ctx := context.Background()
	queries := db.New(dbConn)

	role, err := queries.GetRoleById(ctx, int32(1))
	if err != nil {
		log.Fatalf("Error getting role by ID, %v.", err)
	}

	fmt.Printf("\nRole: %+v\n", role)

	err = server.Start(cfg.Server.Port)
	if err != nil {
		zlog.Zap.Fatalf("Cannot start server: %s", err)
	}
}
