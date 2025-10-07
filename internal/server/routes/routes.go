package routes

import (
	"net/http"
	"strings"
	"time"

	_ "github.com/dockworks/dm-web-backend/docs"
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	pm "github.com/dockworks/dm-web-backend/internal/server/middleware"
	"go.uber.org/zap"

	"github.com/brpaz/echozap"
	"github.com/dockworks/dm-web-backend/internal/responses"
	"github.com/dockworks/dm-web-backend/pkg/token"
	"github.com/golang-jwt/jwt/v5"
	echojwt "github.com/labstack/echo-jwt/v4"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	echoSwagger "github.com/swaggo/echo-swagger"
)

func RegisterRoutes(s *s.Server) {
	// Server config
	s.Echo.Server.ReadTimeout = 10 * time.Second
	s.Echo.Server.WriteTimeout = 30 * time.Second
	s.Echo.Server.IdleTimeout = time.Minute
	s.Echo.Validator = s.Config.Server.Validator
	s.Echo.Binder = s.Config.Server.Binder

	zapLogger := s.Logger.DesugarZap
	// Core handlers for base routes
	genericHandler := h.NewGenericHandler(s)

	// Middlewares
	s.Echo.Use(middleware.RequestID())
	s.Echo.Use(echozap.ZapLogger(s.Logger.DesugarZap))
	s.Echo.Use(middleware.CORSWithConfig(s.Config.Server.CORSConfig))
	s.Echo.Use(middleware.Recover())
	s.Echo.Use(middleware.TimeoutWithConfig(middleware.TimeoutConfig{
		Timeout: 30 * time.Second,
	}))
	s.Echo.Use(middleware.BodyDumpWithConfig(middleware.BodyDumpConfig{
		Skipper: func(c echo.Context) bool {
			path := c.Request().URL.Path
			return strings.HasPrefix(path, "/swagger/")
		},
		Handler: func(c echo.Context, reqBody, resBody []byte) {
			zapLogger.Info("Request Body", zap.String("body", string(reqBody)), zap.String("path", c.Path()), zap.String("method", c.Request().Method), zap.String("query", c.QueryString()), zap.String("remote_ip", c.RealIP()), zap.String("host", c.Request().Host), zap.String("user_agent", c.Request().UserAgent()), zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)))
			zapLogger.Info("Response Body", zap.String("body", string(resBody)), zap.String("path", c.Path()), zap.String("method", c.Request().Method), zap.String("query", c.QueryString()), zap.String("remote_ip", c.RealIP()), zap.String("host", c.Request().Host), zap.String("user_agent", c.Request().UserAgent()), zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)))
		},
	}))

	// Base Routes
	s.Echo.GET("/swagger/*", echoSwagger.WrapHandler)

	// Versioned Routes
	base := s.Echo.Group("/api/v1")

	base.GET("/health", genericHandler.HealthHandler)
	base.GET("/project-details", genericHandler.ProjectDetailsHandler)

	// Authentication routes
	RegisterAuthRoutes(s, base)

	protected := base.Group("")
	// Configure middleware with the custom claims type
	config := echojwt.Config{
		NewClaimsFunc: func(_ echo.Context) jwt.Claims {
			return new(token.JwtCustomClaims)
		},
		SigningKey: []byte(s.Config.Auth.AccessSecret),
		ErrorHandler: func(c echo.Context, err error) error {
			s.Logger.Zap.Error("JWT validation failed", zap.Error(err))
			return responses.NewErrorResponse(http.StatusUnauthorized, "Token validation failed").JSON(c)
		},
		TokenLookup: "header:Authorization:Bearer ",
		ParseTokenFunc: func(c echo.Context, auth string) (interface{}, error) {
			token, err := jwt.ParseWithClaims(auth, &token.JwtCustomClaims{}, func(t *jwt.Token) (interface{}, error) {
				return []byte(s.Config.Auth.AccessSecret), nil
			})
			if err != nil {
				return nil, err
			}
			if !token.Valid {
				return nil, jwt.ErrTokenExpired
			}
			return token, nil
		},
	}
	protected.Use(echojwt.WithConfig(config))

	permissionMiddleware, err := pm.InitializePermissionMiddleware(s.DB.Queries())
	if err != nil {
		s.Logger.Zap.Error("Failed to initialize permission middleware", err)
	}

	permissionProtected := protected.Group("")
	if permissionMiddleware != nil {
		permissionProtected.Use(permissionMiddleware.RequirePermission())
	}

	// Invite routes (public and protected)
	RegisterInviteRoutes(s, base, permissionProtected)

	// E-signature routes (public and protected)
	RegisterEsignRoutes(s, base, permissionProtected)

	// User routes
	RegisterUserRoutes(s, permissionProtected)

	// Customer routes
	RegisterCustomerRoutes(s, base, permissionProtected)

	// DME routes (includes external API and protected routes)
	RegisterDMERoutes(s, base, permissionProtected)

	// Role routes
	RegisterRoleRoutes(s, permissionProtected)

	// Organization routes
	RegisterOrganizationRoutes(s, permissionProtected)

	// Marina routes
	RegisterMarinaRoutes(s, permissionProtected)

	// Criteria routes
	RegisterCriteriaRoutes(s, permissionProtected)

	// Address routes
	RegisterAddressRoutes(s, permissionProtected)

	// Email routes
	RegisterEmailRoutes(s, s.Config, permissionProtected)

	// SMS routes
	RegisterSMSRoutes(s, s.Config, permissionProtected)

	// Boat routes
	RegisterBoatRoutes(s, permissionProtected)

	// Gallery routes
	RegisterGalleryRoutes(s, permissionProtected)

	// Document routes
	RegisterDocumentRoutes(s, base, permissionProtected)

	// Work Order routes
	RegisterWorkOrderRoutes(s, permissionProtected)

	// Estimate routes
	RegisterEstimateRoutes(s, permissionProtected)

	// Service routes
	RegisterServiceRoutes(s, permissionProtected)

	// Invoice routes
	RegisterInvoiceRoutes(s, permissionProtected)

	// Message routes
	RegisterMessageRoutes(s, permissionProtected)

	// AI routes
	RegisterAIRoutes(s, permissionProtected)

	// Marina Usage History routes
	RegisterMarinaUsageHistoryRoutes(s, protected)

	// Plan routes
	RegisterPlanRoutes(s, protected)

	// Test routes (for testing and development)
	RegisterTestRoutes(s, permissionProtected)

	// Notification routes
	RegisterNotificationRoutes(s, permissionProtected)

	// Admin routes
	RegisterAdminRoutes(s, permissionProtected)

	// Notification preference routes
	RegisterNotificationPreferenceRoutes(s, permissionProtected)

	// General DME API routes
	RegisterGeneralRoutes(s, permissionProtected)

	// Inventory routes
	RegisterInventoryRoutes(s, base, permissionProtected)
}
