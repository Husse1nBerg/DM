package handlers

import (
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/requests"
	"github.com/dockworks/dm-web-backend/internal/responses"
	"github.com/dockworks/dm-web-backend/internal/server"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
	"github.com/labstack/echo/v4"
)

// EmailHandler handles email-related requests
type EmailHandler struct {
	server *server.Server
}

// NewEmailHandler creates a new EmailHandler
func NewEmailHandler(srv *server.Server, cfg *config.Config) *EmailHandler {
	return &EmailHandler{
		server: srv,
	}
}

// SendHTMLEmail sends an HTML email
// @Summary Send HTML email
// @Description Send an email with HTML content
// @Tags Email
// @Accept json
// @Produce json
// @Param params body requests.SendHTMLEmailRequest true "Email details"
// @Success 200 {object} responses.EmailSendResponse "Email accepted for delivery"
// @Failure 400 {object} responses.Error "Validation error"
// @Failure 500 {object} responses.Error "Server error"
// @Router /email/send-html [post]
func (h *EmailHandler) SendHTMLEmail(c echo.Context) error {
	logger := h.server.Logger

	// Bind and validate request
	req := new(requests.SendHTMLEmailRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Create email data
	email := &sendgrid.HTMLEmail{
		EmailData: sendgrid.EmailData{
			To:        req.To,
			Subject:   req.Subject,
			FromEmail: req.FromEmail,
			FromName:  req.FromName,
		},
		HTMLContent: req.HTMLContent,
		PlainText:   req.PlainText,
	}

	// Send email asynchronously
	taskID, _ := h.server.SendGrid.SendHTMLEmail(email)

	// Log the task
	logger.Zap.Infow("HTML email queued", "task_id", taskID.String(), "to", req.To)

	// Return response to client
	return responses.NewEmailSendResponse(
		taskID,
		"Email accepted for delivery",
		true,
	).JSON(c)
}

// SendWelcomeEmail sends a welcome email using the welcome template
// @Summary Send welcome email
// @Description Send a welcome email using the predefined template
// @Tags Email
// @Accept json
// @Produce json
// @Param params body requests.WelcomeEmailRequest true "Welcome email details"
// @Success 200 {object} responses.EmailSendResponse "Email accepted for delivery"
// @Failure 400 {object} responses.Error "Validation error"
// @Failure 500 {object} responses.Error "Server error"
// @Router /email/welcome [post]
func (h *EmailHandler) SendWelcomeEmail(c echo.Context) error {
	logger := h.server.Logger

	// Bind and validate request
	req := new(requests.WelcomeEmailRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Create template data
	templateData := sendgrid.WelcomeTemplateData{
		FirstName:   req.FirstName,
		Username:    req.Username,
		LoginURL:    req.LoginURL,
		CompanyName: req.CompanyName,
		SupportURL:  req.SupportURL,
	}

	// Send email using template
	taskID, resultChan, err := h.server.SendGrid.SendWelcomeEmail(
		req.To,
		req.Subject,
		templateData,
	)

	if err != nil {
		logger.Zap.Errorw("Failed to send welcome email", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to send welcome email: "+err.Error()).JSON(c)
	}

	// Log the task
	logger.Zap.Infow("Welcome email queued", "task_id", taskID.String(), "to", req.To)

	// Process the result asynchronously to log success/failure
	go func() {
		result := <-resultChan
		if result.Status == sendgrid.StatusSent {
			logger.Zap.Infow("Welcome email sent successfully",
				"to", req.To,
				"task_id", result.ID.String())
		} else {
			logger.Zap.Errorw("Failed to send welcome email",
				"to", req.To,
				"task_id", result.ID.String(),
				"error", result.Error)
		}
	}()

	// Return response to client
	return responses.NewEmailSendResponse(
		taskID,
		"Welcome email accepted for delivery",
		true,
	).JSON(c)
}

// SendPasswordResetEmail sends a password reset email using the password reset template
// @Summary Send password reset email
// @Description Send a password reset email using the predefined template
// @Tags Email
// @Accept json
// @Produce json
// @Param params body requests.PasswordResetEmailRequest true "Password reset email details"
// @Success 200 {object} responses.EmailSendResponse "Email accepted for delivery"
// @Failure 400 {object} responses.Error "Validation error"
// @Failure 500 {object} responses.Error "Server error"
// @Router /email/password-reset [post]
func (h *EmailHandler) SendPasswordResetEmail(c echo.Context) error {
	logger := h.server.Logger

	// Bind and validate request
	req := new(requests.PasswordResetEmailRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Build the reset URL
	baseURL := "https://app.dockmaster.com" // Default URL
	resetURL := baseURL + "/reset-password?token=" + req.Token + "&email=" + url.QueryEscape(req.Email)

	// Create template data
	templateData := sendgrid.PasswordResetTemplateData{
		FirstName:   req.FirstName,
		ResetURL:    resetURL,
		Token:       req.Token,
		Email:       req.Email,
		ExpiresIn:   req.ExpiresIn,
		CompanyName: req.CompanyName,
	}

	// Send email using template
	taskID, resultChan, err := h.server.SendGrid.SendPasswordResetEmail(
		req.To,
		req.Subject,
		templateData,
	)

	if err != nil {
		logger.Zap.Errorw("Failed to send password reset email", "error", err)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to send password reset email: "+err.Error()).JSON(c)
	}

	// Log the task
	logger.Zap.Infow("Password reset email queued", "task_id", taskID.String(), "to", req.To)

	// Process the result asynchronously to log success/failure
	go func() {
		result := <-resultChan
		if result.Status == sendgrid.StatusSent {
			logger.Zap.Infow("Password reset email sent successfully",
				"to", req.To,
				"task_id", result.ID.String())
		} else {
			logger.Zap.Errorw("Failed to send password reset email",
				"to", req.To,
				"task_id", result.ID.String(),
				"error", result.Error)
		}
	}()

	// Return response to client
	return responses.NewEmailSendResponse(
		taskID,
		"Password reset email accepted for delivery",
		true,
	).JSON(c)
}

// SendTemplateEmail sends an email using a SendGrid template
// @Summary Send template email
// @Description Send an email using a SendGrid dynamic template
// @Tags Email
// @Accept json
// @Produce json
// @Param params body requests.SendTemplateEmailRequest true "Email details"
// @Success 200 {object} responses.EmailSendResponse "Email accepted for delivery"
// @Failure 400 {object} responses.Error "Validation error"
// @Failure 500 {object} responses.Error "Server error"
// @Router /email/send-template [post]
func (h *EmailHandler) SendTemplateEmail(c echo.Context) error {
	logger := h.server.Logger

	// Bind and validate request
	req := new(requests.SendTemplateEmailRequest)
	if err := c.Bind(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	if err := c.Validate(req); err != nil {
		return responses.NewErrorResponse(http.StatusBadRequest, err).JSON(c)
	}

	// Parse template data from JSON string
	var templateData map[string]interface{}
	if err := json.Unmarshal([]byte(req.TemplateData), &templateData); err != nil {
		logger.Zap.Errorw("Failed to parse template data JSON", "error", err)
		return responses.NewErrorResponse(http.StatusBadRequest, "Invalid template data JSON format").JSON(c)
	}

	// Send email using template
	taskID, _, err := h.server.SendGrid.SendTemplateByName(
		req.TemplateName,
		req.To,
		req.Subject,
		templateData,
	)

	if err != nil {
		logger.Zap.Errorw("Failed to send template email", "error", err, "template", req.TemplateName)
		return responses.NewErrorResponse(http.StatusInternalServerError, "Failed to send template email: "+err.Error()).JSON(c)
	}

	// Log the task
	logger.Zap.Infow("Template email queued", "task_id", taskID.String(), "template", req.TemplateName, "to", req.To)

	// Return response to client
	return responses.NewEmailSendResponse(
		taskID,
		"Email accepted for delivery",
		true,
	).JSON(c)
}
