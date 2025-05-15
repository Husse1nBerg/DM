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
	organizationHandler := h.NewOrganizationHandler(s)
	addressHandler := h.NewAddressHandler(s)
	marinaHandler := h.NewMarinaHandler(s)
	roleHandler := h.NewRoleHandler(s)
	dmeCredentialHandler := h.NewDMECredentialHandler(s)
	dmeSysIDHandler := h.NewDMESysIDHandler(s)
	customerHandler := h.NewCustomerHandler(s)
	emailHandler := h.NewEmailHandler(s, s.Config)
	smsHandler := h.NewSMSHandler(s, s.Config)
	boatHandler := h.NewBoatHandler(s)
	galleryHandler := h.NewGalleryHandler(s)
	workOrderHandler := h.NewWorkOrderHandler(s)

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
	base.GET("/project-details", genericHandler.ProjectDetailsHandler)

	auth := base.Group("/auth")
	auth.POST("/login", authHandler.Login)
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
	users.POST("/reset-password", userHandler.ResetPassword)

	// Password recovery (public endpoints)
	base.POST("/user/forgot-password", userHandler.ForgotPassword)
	base.POST("/user/recover-password", userHandler.RecoverPassword)

	// User by role, organization, marina
	users.GET("/role/:roleId", userHandler.GetUsersByRoleHandler)
	users.GET("/organization/:organizationId", userHandler.GetUsersByOrganizationHandler)
	users.GET("/marina/:marinaId", userHandler.GetUsersByMarinaHandler)
	users.GET("/marina/:marinaId/assigned", userHandler.GetMarinaUsersList)

	// User-marina assignments
	users.POST("/marina/assign", userHandler.AssignUserToMarinaHandler)
	users.POST("/marina/unassign", userHandler.UnassignUserFromMarinaHandler)

	// Role routes
	roles := protected.Group("/role")
	roles.GET("/list", roleHandler.ListRolesHandler)
	roles.POST("", roleHandler.CreateRoleHandler)
	roles.GET("/:roleId", roleHandler.GetRoleHandler)
	roles.PUT("/:roleId", roleHandler.UpdateRoleHandler)
	roles.DELETE("/:roleId", roleHandler.DeleteRoleHandler)
	roles.GET("/name", roleHandler.GetRoleByNameHandler)

	// Organization routes
	organizations := protected.Group("/organizations")
	organizations.POST("", organizationHandler.CreateOrganization)
	organizations.GET("", organizationHandler.GetOrganizationsPaginated)
	organizations.GET("/by-email", organizationHandler.GetOrganizationByEmail)
	organizations.GET("/:id", organizationHandler.GetOrganizationByID)
	organizations.GET("/:id/with-address", organizationHandler.GetOrganizationWithAddress)
	organizations.PUT("/:id", organizationHandler.UpdateOrganization)
	organizations.PUT("/:id/with-address", organizationHandler.UpdateOrgAddress)
	organizations.DELETE("/:id", organizationHandler.DeleteOrganization)

	// Marina routes
	marinas := protected.Group("/marinas")
	marinas.POST("", marinaHandler.CreateMarina)
	marinas.GET("", marinaHandler.GetMarinasPaginated)
	marinas.GET("/by-email", marinaHandler.GetMarinaByEmail)
	marinas.GET("/organization/:organizationId", marinaHandler.GetMarinasByOrganization)
	marinas.GET("/user/:userId", marinaHandler.GetUserMarinas)
	marinas.GET("/user", marinaHandler.GetMyUserMarinas)
	marinas.GET("/:id", marinaHandler.GetMarinaByID)
	marinas.GET("/:id/with-address", marinaHandler.GetMarinaWithAddress)
	marinas.PUT("/:id", marinaHandler.UpdateMarina)
	marinas.PUT("/:id/with-address", marinaHandler.UpdateMarinaWithAddress)
	marinas.DELETE("/:id", marinaHandler.DeleteMarina)

	// Address routes
	addresses := protected.Group("/addresses")
	addresses.POST("", addressHandler.CreateAddress)
	addresses.GET("/:id", addressHandler.GetAddressById)
	addresses.PUT("/:id", addressHandler.UpdateAddress)

	// DME routes
	dme := protected.Group("/dme")

	// DME Credentials routes
	credGroup := dme.Group("/credentials")
	credGroup.POST("", dmeCredentialHandler.CreateDMECredential)
	credGroup.GET("/organization/:organizationId", dmeCredentialHandler.GetDMECredentialByOrgID)
	credGroup.PUT("/organization/:organizationId", dmeCredentialHandler.UpdateDMECredential)
	credGroup.DELETE("/organization/:organizationId", dmeCredentialHandler.DeleteDMECredential)

	// DME System ID routes
	sysidGroup := dme.Group("/sysids")
	sysidGroup.POST("", dmeSysIDHandler.CreateDMESysID)
	sysidGroup.GET("", dmeSysIDHandler.ListDMESysIDs)
	sysidGroup.GET("/:id", dmeSysIDHandler.GetDMESysIDByID)
	sysidGroup.PUT("/:id", dmeSysIDHandler.UpdateDMESysID)
	sysidGroup.DELETE("/:id", dmeSysIDHandler.DeleteDMESysID)
	sysidGroup.GET("/system/:systemId", dmeSysIDHandler.GetDMESysIDBySystemID)
	sysidGroup.GET("/organization/:organizationId", dmeSysIDHandler.GetDMESysIDsByOrgID)
	sysidGroup.GET("/marina/:marinaId", dmeSysIDHandler.GetDMESysIDByMarinaID)
	sysidGroup.PATCH("/:id/link", dmeSysIDHandler.LinkDMESysIDToMarina)
	sysidGroup.PATCH("/:id/unlink", dmeSysIDHandler.UnlinkDMESysIDFromMarina)

	// Customer routes
	customers := protected.Group("/customers")
	customers.GET("/list", customerHandler.ListCustomersByPage)
	customers.GET("/list-short", customerHandler.ListCustomersShortByPage)
	customers.GET("/retrieve", customerHandler.RetrieveCustomer)
	customers.GET("/search", customerHandler.SearchCustomers)
	customers.POST("/update", customerHandler.UpdateCustomer)
	customers.POST("/create", customerHandler.CreateCustomer)
	customers.GET("/settings", customerHandler.GetCustomerSettings)
	customers.POST("/settings", customerHandler.UpdateCustomerSettings)

	// Email routes
	emails := protected.Group("/email")
	emails.POST("/send-html", emailHandler.SendHTMLEmail)
	emails.POST("/send-template", emailHandler.SendTemplateEmail)

	// SMS routes
	sms := protected.Group("/sms")
	sms.POST("/send", smsHandler.SendSMS)
	sms.POST("/send-batch", smsHandler.SendBatchSMS)

	// Boat routes
	boats := protected.Group("/boats")
	boats.GET("/list", boatHandler.ListBoatsByPage)
	boats.GET("/retrieve", boatHandler.RetrieveBoat)
	boats.GET("/customer", boatHandler.RetrieveBoatsForCustomer)
	boats.GET("/search", boatHandler.SearchBoats)
	boats.POST("/update", boatHandler.UpdateBoat)
	boats.POST("/create", boatHandler.CreateBoat)

	// Gallery routes
	gallery := protected.Group("/gallery")
	gallery.POST("/marina", galleryHandler.CreateMarinaGalleryItem)
	gallery.GET("/marina/:marinaId", galleryHandler.GetMarinaGallery)
	gallery.GET("/marina/item/:id", galleryHandler.GetMarinaGalleryItem)
	gallery.PUT("/marina/item/:id", galleryHandler.UpdateMarinaGalleryItem)
	gallery.DELETE("/marina/item/:id", galleryHandler.DeleteMarinaGalleryItem)

	gallery.POST("/boat", galleryHandler.CreateVesselGalleryItem)
	gallery.GET("/boat/:boatId", galleryHandler.GetVesselGallery)
	gallery.GET("/boat/item/:id", galleryHandler.GetVesselGalleryItem)
	gallery.PUT("/boat/item/:id", galleryHandler.UpdateVesselGalleryItem)
	gallery.DELETE("/boat/item/:id", galleryHandler.DeleteVesselGalleryItem)
	// Work Order routes
	workOrders := protected.Group("/work-orders")
	workOrders.GET("/list", workOrderHandler.ListWorkOrdersByPage)
	workOrders.GET("/retrieve", workOrderHandler.RetrieveWorkOrder)
	workOrders.GET("/search", workOrderHandler.SearchWorkOrders)
	workOrders.GET("/customer", workOrderHandler.ListWorkOrdersForCustomer)
	workOrders.GET("/operations", workOrderHandler.RetrieveWorkOrderOperations)
	workOrders.GET("/completed", workOrderHandler.RetrieveCompletedWorkOrders)
	workOrders.POST("/update", workOrderHandler.UpdateWorkOrder)
	workOrders.POST("/create", workOrderHandler.CreateWorkOrder)
	workOrders.POST("/create-from-estimate", workOrderHandler.CreateWorkOrderFromEstimate)
}
