package sendgrid

import (
	"github.com/google/uuid"
)

// EmailData represents the basic email data
type EmailData struct {
	To        []string // List of recipient email addresses
	Subject   string   // Email subject
	FromEmail string   // Sender email
	FromName  string   // Sender name
}

// EmailStatus represents the status of an email sending task
type EmailStatus struct {
	ID        uuid.UUID // Unique identifier for the email task
	Status    string    // Status: pending, sent, failed
	Timestamp int64     // Timestamp of the status update
	Error     string    // Error message if any
}

// HTMLEmail contains data for sending a standard HTML email
type HTMLEmail struct {
	EmailData
	HTMLContent string // HTML content of the email
	PlainText   string // Plain text version of the email
}

// TemplateEmail contains data for sending an email using a SendGrid template
type TemplateEmail struct {
	EmailData
	TemplateID   string                 // SendGrid template ID
	TemplateData map[string]interface{} // Dynamic template data
}

// WelcomeTemplateData contains specific fields for the welcome email template
type WelcomeTemplateData struct {
	FirstName   string `json:"first_name"`   // User's first name
	Username    string `json:"username"`     // User's username for login
	LoginURL    string `json:"login_url"`    // URL for the login page
	CompanyName string `json:"company_name"` // Organization/company name
	SupportURL  string `json:"support_url"`  // URL for support/help pages
}

// PasswordResetTemplateData contains specific fields for the password reset template
type PasswordResetTemplateData struct {
	FirstName   string `json:"first_name"`   // User's first name
	ResetURL    string `json:"reset_url"`    // URL with token for password reset
	Token       string `json:"token"`        // Reset token
	Email       string `json:"email"`        // User's email
	ExpiresIn   string `json:"expires_in"`   // Expiration time (e.g., "24 hours")
	CompanyName string `json:"company_name"` // Organization/company name
}
