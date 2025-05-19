package telgorithm

import (
	"time"

	"github.com/google/uuid"
)

// SMSData represents the basic SMS data
type SMSData struct {
	To         []string   // List of recipient phone numbers
	Message    string     // SMS message content
	FromNumber string     // Sender phone number
	MediaURLs  []string   // Optional media URLs for MMS
	Priority   string     // Optional priority
	ExpiresOn  *time.Time // Optional expiration time
}

// BatchSMSData represents a batch of SMS messages to send
type BatchSMSData struct {
	Messages   []SMSData // Individual SMS messages to send
	FromNumber string    // Default sender phone number (optional)
}

// BatchSMSResult represents the results of a batch send operation
type BatchSMSResult struct {
	TaskID       uuid.UUID   // Unique identifier for the batch task
	Results      []SMSStatus // Individual results for each message
	SuccessCount int         // Number of successfully sent messages
	FailureCount int         // Number of failed messages
	Timestamp    int64       // Timestamp of the operation
}

// SMSStatus represents the status of an SMS sending task
type SMSStatus struct {
	ID           uuid.UUID // Unique identifier for the SMS task
	Status       string    // Status: pending, sent, failed
	MessageID    string    // Telgorithm message ID for tracking
	Timestamp    int64     // Timestamp of the status update
	Error        string    // Error message if any
	SegmentCount int       // Number of segments in the SMS
	Recipient    string    // The recipient of this specific message
}

// SMSResponse represents the response from the Telgorithm API
type SMSResponse struct {
	SID          string `json:"sid"`
	SegmentCount int    `json:"segmentCount"`
}

// SendSMSRequest represents the request payload to Telgorithm API
type SendSMSRequest struct {
	From                 string   `json:"from"`
	To                   string   `json:"to,omitempty"`
	Recipients           []string `json:"recipients,omitempty"`
	Text                 string   `json:"text"`
	Priority             string   `json:"priority,omitempty"`
	MediaURLs            []string `json:"mediaUrls,omitempty"`
	CallbackURLOverride  string   `json:"callbackUrlOverride,omitempty"`
	ExpiresOn            string   `json:"expiresOn,omitempty"`
	ConversationMetadata string   `json:"conversationMetadata,omitempty"`
	IgnoreSendingTimes   bool     `json:"ignoreSendingTimes,omitempty"`
}
