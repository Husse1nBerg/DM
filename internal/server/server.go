package server

import (
	"github.com/dockworks/dm-web-backend/internal/config"
	db "github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/labstack/echo/v4"
)

type Server struct {
	Echo         *echo.Echo
	Config       *config.Config
	DB           db.DBService
	Logger       *logger.Logger
	S3Service    *s3.S3Service
	ImageService *s3.ImageService
	DME          *dme.Client
}

func NewServer(cfg *config.Config, logger *logger.Logger) *Server {
	dbConn := db.NewConnection(&cfg.DB)
	// Initialize S3 service
	s3Service, err := s3.NewS3Service(cfg.S3)
	if err != nil {
		logger.Zap.Error("Failed to initialize S3 service", err, "initialization")
	}

	// Initialize image service with S3 and BaseURL from config
	imageService := s3.NewImageService(s3Service, cfg.S3.BaseURL)

	// Set the image service in the responses package
	utils.SetImageService(imageService)

	return &Server{
		Config:       cfg,
		Echo:         echo.New(),
		DB:           dbConn,
		Logger:       logger,
		S3Service:    s3Service,
		ImageService: imageService,
		DME:          dme.NewClientFromConfig(cfg, logger, dbConn),
	}
}

func (server *Server) Start(addr string) error {
	return server.Echo.Start(":" + addr)
}
