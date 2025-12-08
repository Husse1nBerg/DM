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
	Subject     string // Subject of the email
	ReplyTo     string // Reply to email address
	ReplyName   string // Reply to name
}

// Attachment represents an email attachment
type Attachment struct {
	Content     []byte // File content
	Filename    string // Filename for the attachment
	Type        string // MIME type (e.g., "application/pdf")
	Disposition string // Content disposition (default: "attachment")
	ContentID   string // Content ID for inline attachments (optional)
}

// TemplateEmail contains data for sending an email using a SendGrid template
type TemplateEmail struct {
	EmailData
	TemplateID   string                 // SendGrid template ID
	TemplateData map[string]interface{} // Dynamic template data
	ReplyTo      string                 // Reply to email address
	ReplyName    string                 // Reply to name
	Subject      string                 // Subject of the email
	Attachments  []Attachment           // Email attachments
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
	Logo            string `json:"logo,omitempty"`   // URL for the customer logo
}

// MessageTemplateData contains specific fields for the message template
type MessageTemplateData struct {
	Content         string `json:"content"`          // Message content
	Recipient       string `json:"recipient"`        // Recipient name
	Sender          string `json:"sender"`           // Sender name
	HomeURL         string `json:"home_url"`         // URL for the home page
	TermsConditions string `json:"terms_conditions"` // URL for the terms and conditions
	Logo            string `json:"logo,omitempty"`   // URL for the customer logo
}

// MessageTemplateData contains specific fields for the message template
type ExternalMessageTemplateData struct {
	Content         string `json:"content"`          // Message content
	Recipient       string `json:"recipient"`        // Recipient name
	Sender          string `json:"sender"`           // Sender name
	ReplyTo         string `json:"reply_to"`         // Reply to email address
	TermsConditions string `json:"terms_conditions"` // URL for the terms and conditions
	Logo            string `json:"logo,omitempty"`   // URL for the customer logo
}

// InviteTemplateData contains specific fields for the invite template
type InviteTemplateData struct {
	UserName        string `json:"user_name"`        // User's username for login
	InviteURL       string `json:"invite_url"`       // URL for the invite
	TermsConditions string `json:"terms_conditions"` // URL for the terms and conditions
	Logo            string `json:"logo,omitempty"`   // URL for the customer logo
}

type InviteCustomerTemplateData struct {
	UserName        string `json:"user_name"`        // User's username for login
	InviteURL       string `json:"invite_url"`       // URL for the invite
	TermsConditions string `json:"terms_conditions"` // URL for the terms and conditions
	Logo            string `json:"logo,omitempty"`   // URL for the customer logo
}

// AssignedToMarinaTemplateData contains specific fields for the assigned_to_marina template
//
//	{
//	    "customer_logo":"sdfs",
//	    "business_name":"Big Marina",
//	    "user_name":"Adam",
//	    "home_url":"sdfsfsd",
//	    "terms_conditions": ""
//	}
type AssignedToMarinaTemplateData struct {
	CustomerLogo    string `json:"customer_logo"`
	BusinessName    string `json:"business_name"`
	UserName        string `json:"user_name"`
	HomeURL         string `json:"home_url"`
	TermsConditions string `json:"terms_conditions"`
}

type ESignSubmissionTemplateData struct {
	Recipient       string `json:"recipient"`                // Recipient name
	Sender          string `json:"sender"`                   // Sender name
	ReplyTo         string `json:"reply_to"`                 // Reply to email address
	ReplyName       string `json:"reply_name"`               // Reply to name
	TermsConditions string `json:"terms_conditions"`         // URL for the terms and conditions
	DocumentURL     string `json:"document_url"`             // URL for the document
	Name            string `json:"name"`                     // Name of the document
	CustomMessage   string `json:"custom_message,omitempty"` // Custom message to the customer
	Logo            string `json:"logo,omitempty"`           // URL for the customer logo
}

type NotificationTemplateData struct {
	Recipient    string `json:"recipient"` // Recipient name
	Type         string `json:"type"`      // Type of notification
	CustomerName string `json:"customer_name"`
	HomeURL      string `json:"home_url"`
	Logo         string `json:"logo,omitempty"` // URL for the customer logo
}

// PaymentLinkTemplateData contains fields for the payment link template
// Expected dynamic data keys in SendGrid template
type PaymentLinkTemplateData struct {
	Recipient       string `json:"recipient"`
	Sender          string `json:"sender"`
	TermsConditions string `json:"terms_conditions"`
	Name            string `json:"name"`
	ReplyName       string `json:"reply_name"`
	CustomMessage   string `json:"custom_message"`
	PaymentURL      string `json:"payment_url"`
	InvoiceID       string `json:"invoice_id"`
	Amount          string `json:"amount"`
	Logo            string `json:"logo,omitempty"` // URL for the customer logo
}
