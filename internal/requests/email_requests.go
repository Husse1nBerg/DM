package requests

// SendHTMLEmailRequest defines the input for the HTML email sending endpoint
// @Description Send HTML email request payload
// @Schema requests.SendHTMLEmailRequest
type SendHTMLEmailRequest struct {
	To          []string `json:"to" validate:"required,dive,email" example:"user@example.com"`
	Subject     string   `json:"subject" validate:"required" example:"Welcome to Dockmaster"`
	HTMLContent string   `json:"htmlContent" validate:"required" example:"<h1>Hello World</h1><p>This is a test email.</p>"`
	PlainText   string   `json:"plainText" validate:"required" example:"Hello World. This is a test email."`
	FromName    string   `json:"fromName,omitempty" example:"Dockmaster"`
	FromEmail   string   `json:"fromEmail,omitempty" example:"no-reply@dockmaster.com"`
}

// WelcomeEmailRequest defines the input for sending a welcome email
// @Description Send welcome email request payload
// @Schema requests.WelcomeEmailRequest
type WelcomeEmailRequest struct {
	To          []string `json:"to" validate:"required,dive,email" example:"user@example.com"`
	Subject     string   `json:"subject" validate:"required" example:"Welcome to Dockmaster"`
	FirstName   string   `json:"firstName" validate:"required" example:"John"`
	Username    string   `json:"username" validate:"required" example:"john.doe"`
	LoginURL    string   `json:"loginUrl" validate:"required,url" example:"https://app.dockmaster.com/login"`
	CompanyName string   `json:"companyName" example:"Dockmaster"`
	SupportURL  string   `json:"supportUrl" example:"https://support.dockmaster.com"`
	FromName    string   `json:"fromName,omitempty" example:"Dockmaster"`
	FromEmail   string   `json:"fromEmail,omitempty" example:"no-reply@dockmaster.com"`
}

// PasswordResetEmailRequest defines the input for sending a password reset email
// @Description Send password reset email request payload
// @Schema requests.PasswordResetEmailRequest
type PasswordResetEmailRequest struct {
	To          []string `json:"to" validate:"required,dive,email" example:"user@example.com"`
	Subject     string   `json:"subject" validate:"required" example:"Reset Your Password"`
	FirstName   string   `json:"firstName" validate:"required" example:"John"`
	Token       string   `json:"token" validate:"required" example:"a1b2c3d4e5f6g7h8i9j0"`
	Email       string   `json:"email" validate:"required,email" example:"john.doe@example.com"`
	ExpiresIn   string   `json:"expiresIn" example:"24 hours"`
	CompanyName string   `json:"companyName" example:"Dockmaster"`
	FromName    string   `json:"fromName,omitempty" example:"Dockmaster"`
	FromEmail   string   `json:"fromEmail,omitempty" example:"no-reply@dockmaster.com"`
}

// SendTemplateEmailRequest defines the input for the template email sending endpoint
// This is kept for backward compatibility and general-purpose template sending
// @Description Send template email request payload
// @Schema requests.SendTemplateEmailRequest
type SendTemplateEmailRequest struct {
	To           []string `json:"to" validate:"required,dive,email" example:"user@example.com"`
	Subject      string   `json:"subject" validate:"required" example:"Welcome to Dockmaster"`
	TemplateName string   `json:"templateName" validate:"required" example:"welcome"`
	TemplateData string   `json:"templateData" validate:"required,json" example:"{\"name\":\"John Doe\"}"`
	FromName     string   `json:"fromName,omitempty" example:"Dockmaster"`
	FromEmail    string   `json:"fromEmail,omitempty" example:"no-reply@dockmaster.com"`
}
