package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	pm "github.com/dockworks/dm-web-backend/internal/server/middleware"
	"github.com/labstack/echo/v4"
)

// RegisterDMERoutes registers all DME-related routes (credentials and system IDs)
func RegisterDMERoutes(server *s.Server, base *echo.Group, permissionProtected *echo.Group) {
	dmeCredentialHandler := h.NewDMECredentialHandler(server)
	dmeSysIDHandler := h.NewDMESysIDHandler(server)
	esignHandler := h.NewEsignHandler(server)

	// DME E-signature routes (external API with API key authentication)
	dmeRoutes := base.Group("/external/dme")
	dmeRoutes.Use(pm.RequireAPIKey(server.Config))
	dmeRoutes.POST("/esign/documents", esignHandler.CreateEsignDocumentDME)

	// Protected DME routes
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
}
