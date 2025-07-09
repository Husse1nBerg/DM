package routes

import (
	"time"

	_ "github.com/dockworks/dm-web-backend/docs"
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	pm "github.com/dockworks/dm-web-backend/internal/server/middleware"
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
	contactHandler := h.NewContactHandler(s)
	messageHandler := h.NewMessageHandler(s)
	documentHandler := h.NewDocumentHandler(s)
	inviteHandler := h.NewInviteHandler(s)
	marinaUsageHistoryHandler := h.NewMarinaUsageHistoryHandler(s)
	planHandler := h.NewPlanHandler(s)
	permissionTestHandler := h.NewPermissionTestHandler(s)
	redisHandler := h.NewRedisHandler(s)
	notificationHandler := h.NewNotificationHandler(s)
	esignHandler := h.NewEsignHandler(s)
	criteriaHandler := h.NewCriteriaHandler(s)
	adminHandler := h.NewAdminHandler(s)
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

	// Invite routes (public endpoints)
	invite := base.Group("/invite")
	invite.GET("/confirm", inviteHandler.ConfirmToken)
	invite.POST("/accept", inviteHandler.AcceptInvitation)

	protected := base.Group("")
	// Configure middleware with the custom claims type
	config := echojwt.Config{
		NewClaimsFunc: func(_ echo.Context) jwt.Claims {
			return new(token.JwtCustomClaims)
		},
		SigningKey: []byte(s.Config.Auth.AccessSecret),
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

	// Protected invite routes
	protectedInvite := permissionProtected.Group("/invite")
	protectedInvite.POST("/refresh", inviteHandler.RefreshInvite)

	// User routes
	users := permissionProtected.Group("/user")
	users.GET("/profile", userHandler.GetMyUserHandler)
	users.GET("/list", userHandler.ListUsersHandler)
	users.POST("", userHandler.CreateUserHandler)
	users.GET("/:userId", userHandler.GetUserHandler)
	users.PUT("/:userId", userHandler.UpdateUserHandler)
	users.DELETE("/:userId", userHandler.DeleteUserHandler)
	users.POST("/reset-password", userHandler.ResetPassword)
	users.POST("/customer-portal", userHandler.CreateCustomerUserHandler)
	users.GET("/customer-portal/:customerId", userHandler.GetUsersByCustomerIDHandler)
	users.POST("/invite", userHandler.CreateUserWithInvitationHandler)

	// Password recovery (public endpoints)
	base.POST("/user/forgot-password", userHandler.ForgotPassword)
	base.POST("/user/recover-password", userHandler.RecoverPassword)

	// Customer intake (public endpoint)
	base.POST("/customer-intake", customerHandler.CustomerIntake)

	// Public E-signature routes
	publicEsign := base.Group("/public/esign")
	publicEsign.GET("/submissions/:id", esignHandler.GetEsignSubmissionPublic)
	publicEsign.PUT("/submissions/:id", esignHandler.UpdateEsignSubmissionPublic)
	publicEsign.POST("/dme/documents", esignHandler.CreateEsignDocumentDME)

	// User by role, organization, marina
	users.GET("/role/:roleId", userHandler.GetUsersByRoleHandler)
	users.GET("/organization/:organizationId", userHandler.GetUsersByOrganizationHandler)
	users.GET("/marina/:marinaId", userHandler.GetUsersByMarinaHandler)
	users.GET("/marina/:marinaId/assigned", userHandler.GetMarinaUsersList)
	users.GET("/marina/:marinaId/not-assigned", userHandler.GetUsersNotAssignedToMarinaHandler)

	// User-marina assignments
	users.POST("/marina/assign", userHandler.AssignUserToMarinaHandler)
	users.POST("/marina/unassign", userHandler.UnassignUserFromMarinaHandler)

	// Role routes
	roles := permissionProtected.Group("/role")
	roles.GET("/list", roleHandler.ListRolesHandler)
	roles.POST("", roleHandler.CreateRoleHandler)
	roles.GET("/:roleId", roleHandler.GetRoleHandler)
	roles.PUT("/:roleId", roleHandler.UpdateRoleHandler)
	roles.DELETE("/:roleId", roleHandler.DeleteRoleHandler)
	roles.GET("/name", roleHandler.GetRoleByNameHandler)

	// Organization routes
	organizations := permissionProtected.Group("/organizations")
	organizations.POST("", organizationHandler.CreateOrganization)
	organizations.GET("", organizationHandler.GetOrganizationsPaginated)
	organizations.GET("/by-email", organizationHandler.GetOrganizationByEmail)
	organizations.GET("/:id", organizationHandler.GetOrganizationByID)
	organizations.GET("/:id/with-address", organizationHandler.GetOrganizationWithAddress)
	organizations.PUT("/:id", organizationHandler.UpdateOrganization)
	organizations.PUT("/:id/with-address", organizationHandler.UpdateOrgAddress)
	organizations.DELETE("/:id", organizationHandler.DeleteOrganization)

	// Marina routes
	marinas := permissionProtected.Group("/marinas")
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
	marinas.GET("/:id/contacts", contactHandler.ListContacts)
	marinas.POST("/:id/contacts", contactHandler.CreateContact)
	marinas.PUT("/:id/contacts/:contactId", contactHandler.UpdateContact)
	marinas.DELETE("/:id/contacts/:contactId", contactHandler.DeleteContact)

	// Criteria routes
	criteria := permissionProtected.Group("/criteria")
	criteria.GET("", criteriaHandler.ListCriteria)
	criteria.GET("/paginated", criteriaHandler.ListCriteriaPaginated)
	criteria.POST("", criteriaHandler.CreateCriteria)
	criteria.GET("/search", criteriaHandler.SearchCriteria)
	criteria.GET("/search/paginated", criteriaHandler.SearchCriteriaPaginated)
	criteria.GET("/:criteriaId", criteriaHandler.GetCriteria)
	criteria.PUT("/:criteriaId", criteriaHandler.UpdateCriteria)
	criteria.DELETE("/:criteriaId", criteriaHandler.DeleteCriteria)
	criteria.POST("/:criteriaId/duplicate", criteriaHandler.DuplicateCriteria)

	// Address routes
	addresses := permissionProtected.Group("/addresses")
	addresses.POST("", addressHandler.CreateAddress)
	addresses.GET("/:id", addressHandler.GetAddressById)
	addresses.PUT("/:id", addressHandler.UpdateAddress)

	// DME routes
	dme := permissionProtected.Group("/dme")

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
	customers := permissionProtected.Group("/customers")
	customers.GET("/list", customerHandler.ListCustomersByPage)
	customers.GET("/list-short", customerHandler.ListCustomersShortByPage)
	customers.GET("/retrieve", customerHandler.RetrieveCustomer)
	customers.GET("/search", customerHandler.SearchCustomers)
	customers.POST("/update", customerHandler.UpdateCustomer)
	customers.POST("/create", customerHandler.CreateCustomer)
	customers.GET("/settings", customerHandler.GetCustomerSettings)
	customers.POST("/settings", customerHandler.UpdateCustomerSettings)

	// Email routes
	emails := permissionProtected.Group("/email")
	emails.POST("/send-html", emailHandler.SendHTMLEmail)
	emails.POST("/send-template", emailHandler.SendTemplateEmail)

	// SMS routes
	sms := permissionProtected.Group("/sms")
	sms.POST("/send", smsHandler.SendSMS)

	// Boat routes
	boats := permissionProtected.Group("/boats")
	boats.GET("/list", boatHandler.ListBoatsByPage)
	boats.GET("/retrieve", boatHandler.RetrieveBoat)
	boats.GET("/customer", boatHandler.RetrieveBoatsForCustomer)
	boats.GET("/search", boatHandler.SearchBoats)
	boats.POST("/update", boatHandler.UpdateBoat)
	boats.POST("/create", boatHandler.CreateBoat)

	// Gallery routes
	gallery := permissionProtected.Group("/gallery")
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

	// Document routes
	documents := permissionProtected.Group("/documents")
	documents.POST("/customer", documentHandler.CustomerUploadDocument)
	documents.GET("/customer", documentHandler.CustomerGetDocumentsByEntity)
	documents.POST("/boat", documentHandler.BoatUploadDocument)
	documents.GET("/boat", documentHandler.BoatGetDocumentsByEntity)
	documents.POST("/user", documentHandler.UserUploadDocument)
	documents.GET("/user", documentHandler.UserGetDocumentsByEntity)
	documents.GET("/:id", documentHandler.GetDocument)
	documents.DELETE("/:id", documentHandler.DeleteDocument)

	// E-signature routes
	esign := permissionProtected.Group("/esign")
	esign.GET("/templates", esignHandler.ListEsignTemplates)
	esign.GET("/templates/:id", esignHandler.GetEsignTemplate)
	esign.POST("/templates", esignHandler.CreateEsignTemplate)
	esign.PUT("/templates/:id", esignHandler.UpdateEsignTemplate)
	esign.DELETE("/templates/:id", esignHandler.DeleteEsignTemplate)
	esign.POST("/documents", esignHandler.CreateEsignDocument)
	esign.GET("/documents", esignHandler.ListEsignDocuments)
	esign.GET("/documents/:id", esignHandler.GetEsignDocument)
	esign.PUT("/documents/:id", esignHandler.UpdateEsignDocument)
	esign.DELETE("/documents/:id", esignHandler.DeleteEsignDocument)
	esign.GET("/documents/:documentId/submissions", esignHandler.ListEsignSubmissionsByDocument)
	esign.POST("/submissions", esignHandler.CreateEsignSubmission)
	esign.GET("/submissions", esignHandler.ListEsignSubmissions)
	esign.GET("/submissions/:id", esignHandler.GetEsignSubmission)
	esign.PUT("/submissions/:id", esignHandler.UpdateEsignSubmission)
	esign.DELETE("/submissions/:id", esignHandler.DeleteEsignSubmission)
	esign.GET("/submissions/status", esignHandler.ListEsignSubmissionsByStatus)

	// Work Order routes
	workOrders := permissionProtected.Group("/work-orders")
	workOrders.GET("/list", workOrderHandler.ListWorkOrdersByPage)
	workOrders.GET("/retrieve", workOrderHandler.RetrieveWorkOrder)
	workOrders.GET("/search", workOrderHandler.SearchWorkOrders)
	workOrders.GET("/customer", workOrderHandler.ListWorkOrdersForCustomer)
	workOrders.GET("/operations", workOrderHandler.RetrieveWorkOrderOperations)
	workOrders.GET("/completed", workOrderHandler.RetrieveCompletedWorkOrders)
	workOrders.POST("/update", workOrderHandler.UpdateWorkOrder)
	workOrders.POST("/create", workOrderHandler.CreateWorkOrder)
	workOrders.POST("/create-from-estimate", workOrderHandler.CreateWorkOrderFromEstimate)
	workOrders.POST("/delete-operation", workOrderHandler.DeleteWorkOrderOperation)

	// Message routes
	messages := permissionProtected.Group("/message")
	messages.POST("/customer", messageHandler.CreateMessageHandler)
	messages.POST("/marina", messageHandler.CreateMessageMarinaHandler)
	messages.GET("/marina", messageHandler.ListMessagesMarinaHandler)
	messages.GET("/customer", messageHandler.ListMessagesCustomerHandler)
	messages.GET("/get", messageHandler.GetMessageByIDHandler)
	messages.PUT("/customer", messageHandler.UpdateCustomerMessageHandler)
	messages.PUT("/marina", messageHandler.UpdateMarinaMessageHandler)
	messages.DELETE("/customer", messageHandler.DeleteCustomerMessageHandler)
	messages.DELETE("/marina", messageHandler.DeleteMarinaMessageHandler)

	// Marina Usage History routes
	marinaUsageHistory := protected.Group("/marina-usage-history")
	marinaUsageHistory.GET("/:id", marinaUsageHistoryHandler.GetMarinaUsageHistoryByID)
	marinaUsageHistory.GET("/marina/:marinaId", marinaUsageHistoryHandler.GetMarinaUsageHistoryByMarinaID)
	marinaUsageHistory.GET("/marina/:marinaId/range", marinaUsageHistoryHandler.GetMarinaUsageHistoryByDateRange)
	marinaUsageHistory.GET("/marina/:marinaId/latest", marinaUsageHistoryHandler.GetLatestMarinaUsageHistory)
	marinaUsageHistory.GET("/marina/:marinaId/month", marinaUsageHistoryHandler.GetMarinaUsageHistoryByMonth)

	// Plan routes
	plans := protected.Group("/plans")

	// Notes and Messages Plans
	notesMessagesPlans := plans.Group("/notes-messages")
	notesMessagesPlans.GET("", planHandler.ListNotesMessagesPlans)
	notesMessagesPlans.GET("/:planId", planHandler.GetNotesMessagesPlan)

	// Storage Plans
	storagePlans := plans.Group("/storage")
	storagePlans.GET("", planHandler.ListStoragePlans)
	storagePlans.GET("/:planId", planHandler.GetStoragePlan)

	// Document Plans
	documentPlans := plans.Group("/document")
	documentPlans.GET("", planHandler.ListDocumentPlans)
	documentPlans.GET("/:planId", planHandler.GetDocumentPlan)

	// Test routes (for testing the permission system)
	test := permissionProtected.Group("/test")
	test.POST("/permissions", permissionTestHandler.TestPermission)
	test.GET("/permissions/user", permissionTestHandler.GetUserPermissions)
	test.GET("/permissions/routes", permissionTestHandler.GetRoutePermissions)

	// Redis test routes (for testing Redis connection and operations)
	redisTest := test.Group("/redis") // Redis tests under the test group
	redisTest.GET("/ping", redisHandler.Ping)
	redisTest.GET("/connection", redisHandler.TestRedisConnection)
	redisTest.POST("/operations", redisHandler.TestRedisOperations)
	redisTest.GET("/dual-databases", redisHandler.TestDualDatabaseConnection)

	// Cache operations (DB 0)
	redisCache := redisTest.Group("/cache")
	redisCache.POST("/operations", redisHandler.TestCacheOperations)
	redisCache.GET("/key/:key", redisHandler.GetCacheKey)
	redisCache.DELETE("/key/:key", redisHandler.DeleteCacheKey)

	// Task operations (DB 1)
	redisTask := redisTest.Group("/task")
	redisTask.POST("/operations", redisHandler.TestTaskOperations)
	redisTask.GET("/key/:key", redisHandler.GetTaskKey)
	redisTask.DELETE("/key/:key", redisHandler.DeleteTaskKey)

	// Queue operations (DB 1)
	redisQueue := redisTest.Group("/queue")
	redisQueue.POST("/operations", redisHandler.TestQueueOperations)
	redisQueue.GET("/:queue/length", redisHandler.GetQueueLength)

	// Notification routes - READ-ONLY API (notifications are created internally by services)
	notifications := permissionProtected.Group("/notifications")
	notifications.GET("/list", notificationHandler.ListNotificationsHandler)
	notifications.GET("/unread-count", notificationHandler.GetUnreadCountHandler)
	notifications.GET("/stream", notificationHandler.NotificationStreamHandler)
	notifications.GET("/:id", notificationHandler.GetNotificationHandler)
	notifications.PUT("/:id/read", notificationHandler.MarkAsReadHandler)
	notifications.PUT("/mark-all-read", notificationHandler.MarkAllAsReadHandler)
	notifications.GET("/type/:type", notificationHandler.GetNotificationsByTypeHandler)
	notifications.DELETE("/:id", notificationHandler.DeleteNotificationHandler)

	// Admin routes
	admin := permissionProtected.Group("/admin")
	admin.GET("/user/marina/:marinaId", adminHandler.GetUsersByMarinaHandler)
	admin.GET("/role/list", adminHandler.ListRolesHandler)
}
