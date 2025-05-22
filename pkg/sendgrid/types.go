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
	UserName        string `json:"user_name"`        // User's username for login
	HomeURL         string `json:"home_url"`         // URL for the login page
	BusinessName    string `json:"business_name"`    // Organization/company name
	CustomerLogo    string `json:"customer_logo"`    // URL for the customer logo
	TermsConditions string `json:"terms_conditions"` // URL for the terms and conditions
}

// PasswordResetTemplateData contains specific fields for the password reset template
type PasswordResetTemplateData struct {
	UserName        string `json:"user_name"`        // User's username for login
	ResetURL        string `json:"reset_url"`        // URL for the login page
	TermsConditions string `json:"terms_conditions"` // URL for the terms and conditions
}

// MessageTemplateData contains specific fields for the message template
type MessageTemplateData struct {
	Content         string `json:"content"`          // Message content
	Recipient       string `json:"recipient"`        // Recipient name
	Sender          string `json:"sender"`           // Sender name
	TermsConditions string `json:"terms_conditions"` // URL for the terms and conditions
}

// InviteTemplateData contains specific fields for the invite template
type InviteTemplateData struct {
	UserName        string `json:"user_name"`        // User's username for login
	InviteURL       string `json:"invite_url"`       // URL for the invite
	TermsConditions string `json:"terms_conditions"` // URL for the terms and conditions
}
