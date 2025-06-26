package server

import (
	"github.com/dockworks/dm-web-backend/internal/config"
	db "github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/dockworks/dm-web-backend/pkg/redis"
	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
	"github.com/dockworks/dm-web-backend/pkg/telgorithm"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/labstack/echo/v4"
)

type Server struct {
	Echo            *echo.Echo
	Config          *config.Config
	DB              db.DBService
	Logger          *logger.Logger
	Redis           *redis.Client
	S3Service       *s3.S3Service
	DocumentService *s3.S3Service
	ImageService    *s3.ImageService
	ESignService    *s3.ESignService
	DocURLService   *s3.DocumentService
	DME             *dme.Client
	SendGrid        *sendgrid.Client
	Telgorithm      *telgorithm.Client
}

func NewServer(cfg *config.Config, logger *logger.Logger) *Server {
	dbConn := db.NewConnection(&cfg.DB)

	// Initialize Redis client
	redisClient := redis.NewClient(cfg.Redis, logger)

	// Initialize S3 service for images
	s3Service, err := s3.NewS3Service(cfg.S3)
	if err != nil {
		logger.Zap.Error("Failed to initialize S3 service", err, "initialization")
	}

	// Initialize document storage service
	documentService, err := s3.NewS3Service(cfg.DocumentStorage)
	if err != nil {
		logger.Zap.Error("Failed to initialize document storage service", err, "initialization")
	}

	// Initialize image service with S3 and BaseURL from config
	imageService := s3.NewImageService(s3Service, cfg.S3.BaseURL)

	// Initialize document URL service with document storage and BaseURL from config
	docURLService := s3.NewDocumentService(documentService, cfg.DocumentStorage.BaseURL)

	// Initialize e-signature service with S3 and BaseURL from config
	esignS3Service, err := s3.NewS3Service(cfg.ESign)
	if err != nil {
		logger.Zap.Error("Failed to initialize e-signature service", err, "initialization")
	}
	esignService := s3.NewESignService(esignS3Service, cfg.ESign.BaseURL)

	// Set the image service in the responses package
	utils.SetImageService(imageService)
	utils.SetDocumentService(docURLService)
	utils.SetESignService(esignService)

	return &Server{
		Config:          cfg,
		Echo:            echo.New(),
		DB:              dbConn,
		Logger:          logger,
		Redis:           redisClient,
		S3Service:       s3Service,
		DocumentService: documentService,
		ImageService:    imageService,
		ESignService:    esignService,
		DocURLService:   docURLService,
		DME:             dme.NewClientFromConfig(cfg, logger, dbConn),
		SendGrid:        sendgrid.NewClient(cfg),
		Telgorithm:      telgorithm.NewClient(cfg),
	}
}

func (server *Server) Start(addr string) error {
	return server.Echo.Start(":" + addr)
}
