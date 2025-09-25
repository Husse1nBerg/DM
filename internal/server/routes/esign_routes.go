package routes

import (
	s "github.com/dockworks/dm-web-backend/internal/server"
	h "github.com/dockworks/dm-web-backend/internal/server/handlers"
	"github.com/labstack/echo/v4"
)

// RegisterEsignRoutes registers all e-signature related routes (public and protected)
func RegisterEsignRoutes(server *s.Server, base *echo.Group, permissionProtected *echo.Group) {
	esignHandler := h.NewEsignHandler(server)

	// Public E-signature routes
	publicEsign := base.Group("/public/esign")
	publicEsign.GET("/submissions/:id", esignHandler.GetEsignSubmissionPublic)
	publicEsign.PUT("/submissions/:id", esignHandler.UpdateEsignSubmissionPublic)

	// Protected E-signature routes
	esign := permissionProtected.Group("/esign")

	// Template management
	esign.GET("/templates", esignHandler.ListEsignTemplates)
	esign.GET("/templates/:id", esignHandler.GetEsignTemplate)
	esign.POST("/templates", esignHandler.CreateEsignTemplate)
	esign.PUT("/templates/:id", esignHandler.UpdateEsignTemplate)
	esign.DELETE("/templates/:id", esignHandler.DeleteEsignTemplate)

	// Document management
	esign.POST("/documents", esignHandler.CreateEsignDocument)
	esign.GET("/documents", esignHandler.ListEsignDocuments)
	esign.GET("/documents/:id", esignHandler.GetEsignDocument)
	esign.PUT("/documents/:id", esignHandler.UpdateEsignDocument)
	esign.DELETE("/documents/:id", esignHandler.DeleteEsignDocument)
	esign.GET("/documents/:documentId/submissions", esignHandler.ListEsignSubmissionsByDocument)

	// Submission management
	esign.POST("/submissions", esignHandler.CreateEsignSubmission)
	esign.GET("/submissions", esignHandler.ListEsignSubmissions)
	esign.GET("/submissions/:id", esignHandler.GetEsignSubmission)
	esign.PUT("/submissions/:id", esignHandler.UpdateEsignSubmission)
	esign.DELETE("/submissions/:id", esignHandler.DeleteEsignSubmission)
	esign.GET("/submissions/status", esignHandler.ListEsignSubmissionsByStatus)

	// Multiple signature submission management
	esign.POST("/submissions/multiple", esignHandler.CreateMultipleEsignSubmission)
	esign.GET("/submissions/:id/signers", esignHandler.GetEsignSubmissionSigners)
	esign.GET("/submissions/:id/with-signers", esignHandler.GetEsignSubmissionWithSigners)
	esign.PUT("/submissions/signers/:id", esignHandler.UpdateEsignSubmissionSigner)
}
