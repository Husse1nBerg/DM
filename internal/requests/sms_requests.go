package requests

import (
	"time"
)

// SendSMSRequest represents a request to send SMS
type SendSMSRequest struct {
	To         []string   `json:"to" validate:"required,dive,e164"`
	Message    string     `json:"message" validate:"required"`
	FromNumber string     `json:"from_number" validate:"omitempty,e164"`
	MediaURLs  []string   `json:"media_urls" validate:"omitempty,dive,url"`
	Priority   string     `json:"priority" validate:"omitempty,oneof=Urgent High Normal Low"`
	ExpiresOn  *time.Time `json:"expires_on" validate:"omitempty"`
}

// BatchSMSMessageRequest represents a single message in a batch
type BatchSMSMessageRequest struct {
	To        []string   `json:"to" validate:"required,dive,e164"`
	Message   string     `json:"message" validate:"required"`
	MediaURLs []string   `json:"media_urls" validate:"omitempty,dive,url"`
	Priority  string     `json:"priority" validate:"omitempty,oneof=Urgent High Normal Low"`
	ExpiresOn *time.Time `json:"expires_on" validate:"omitempty"`
}

// BatchSMSRequest represents a request to send multiple SMS messages in a batch
type BatchSMSRequest struct {
	Messages   []BatchSMSMessageRequest `json:"messages" validate:"required,dive"`
	FromNumber string                   `json:"from_number" validate:"omitempty,e164"`
}
