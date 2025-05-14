package requests

// SendHTMLEmailRequest defines the input for the HTML email sending endpoint
// @Description Send HTML email request payload
// @Schema requests.SendHTMLEmailRequest
type SendHTMLEmailRequest struct {
	To          []string `json:"to" validate:"required,dive,email" example:"user@example.com"`
	Subject     string   `json:"subject" validate:"required" example:"Welcome to Dockmaster"`
	HTMLContent string   `json:"htmlContent" validate:"required" example:"<h1>Hello World</h1><p>This is a test email.</p>"`
	PlainText   string   `json:"plainText" validate:"required" example:"Hello World. This is a test email."`
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
}
