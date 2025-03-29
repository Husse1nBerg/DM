package routes

import (
	"time"

	_ "github.com/dockworks/dm-web-backend/docs"
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"go.uber.org/zap"

	"github.com/brpaz/echozap"
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
	// Handlers creation
	genericHandler := h.NewGenericHandler(s)
	authHandler := h.NewAuthHandler(s)
	userHandler := h.NewUserHandler(s)

	// Middlewares
	s.Echo.Use(middleware.RequestID())
	s.Echo.Use(echozap.ZapLogger(s.Logger.DesugarZap))
	s.Echo.Use(middleware.CORSWithConfig(s.Config.Server.CORSConfig))
	s.Echo.Use(middleware.Recover())
	s.Echo.Use(middleware.Timeout())
	s.Echo.Use(middleware.BodyDump(func(c echo.Context, reqBody, resBody []byte) {
		zapLogger.Info("Request Body", zap.String("body", string(reqBody)), zap.String("path", c.Path()), zap.String("method", c.Request().Method), zap.String("query", c.QueryString()), zap.String("remote_ip", c.RealIP()), zap.String("host", c.Request().Host), zap.String("user_agent", c.Request().UserAgent()), zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)))
		zapLogger.Info("Response Body", zap.String("body", string(resBody)), zap.String("path", c.Path()), zap.String("method", c.Request().Method), zap.String("query", c.QueryString()), zap.String("remote_ip", c.RealIP()), zap.String("host", c.Request().Host), zap.String("user_agent", c.Request().UserAgent()), zap.String("request_id", c.Response().Header().Get(echo.HeaderXRequestID)))
	}))

	// Base Routes
	s.Echo.GET("/swagger/*", echoSwagger.WrapHandler)

	// Versioned Routes
	base := s.Echo.Group("/api/v1")

	base.GET("/health", genericHandler.HealthHandler)

	auth := base.Group("/auth")
	auth.POST("/login", authHandler.Login)
	auth.POST("/register", authHandler.Register)
	auth.POST("/refresh", authHandler.RefreshToken)

	protected := base.Group("")
	// Configure middleware with the custom claims type
	config := echojwt.Config{
		NewClaimsFunc: func(_ echo.Context) jwt.Claims {
			return new(token.JwtCustomClaims)
		},
		SigningKey: []byte(s.Config.Auth.AccessSecret),
	}
	protected.Use(echojwt.WithConfig(config))

	// User routes
	users := protected.Group("/user")
	users.GET("/profile", userHandler.GetMyUserHandler)
	users.GET("/list", userHandler.ListUsersHandler)
	users.POST("", userHandler.CreateUserHandler)
	users.GET("/:userId", userHandler.GetUserHandler)
	users.PUT("/:userId", userHandler.UpdateUserHandler)
	users.DELETE("/:userId", userHandler.DeleteUserHandler)

	// User by role, organization, marina
	users.GET("/role/:roleId", userHandler.GetUsersByRoleHandler)
	users.GET("/organization/:organizationId", userHandler.GetUsersByOrganizationHandler)
	users.GET("/marina/:marinaId", userHandler.GetUsersByMarinaHandler)
	users.GET("/marina/:marinaId/assigned", userHandler.GetMarinaUsersList)

	// User-marina assignments
	users.POST("/marina/assign", userHandler.AssignUserToMarinaHandler)
	users.POST("/marina/unassign", userHandler.UnassignUserFromMarinaHandler)
}
