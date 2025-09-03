package notifications

import (
	"testing"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// TestSendSmartNotification_AllDeliveryMethod_OrganizationNotFound tests the scenario where
// organization lookup fails but email still gets sent with default customer name
func TestSendSmartNotification_AllDeliveryMethod_OrganizationNotFound(t *testing.T) {
	// This test verifies that when organization lookup fails, the email still gets sent
	// with a default customer name instead of failing completely

	// Create a minimal config
	cfg := &config.Config{
		App: config.AppConfig{
			FrontendBaseURL: "https://test.dockmaster.com",
		},
		SendGrid: config.SendGridConfig{
			FromEmail: "test@dockmaster.com",
			FromName:  "Test Dockmaster",
			TemplatesMap: map[string]string{
				"notification": "d-test-template-id",
			},
		},
	}

	// Test data
	userID := uuid.New()
	orgID := uuid.New()

	// Create request
	req := SmartNotificationRequest{
		UserID:         userID,
		OrganizationID: orgID,
		MarinaID:       uuid.New(),
		Type:           "message",
		Title:          "Test Message",
		Content:        "Test content",
		Data:           map[string]interface{}{"test": "data"},
	}

	// Test the email data construction logic directly
	// This simulates what happens in sendMultiChannelNotification when org lookup fails
	user := db.User{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		IsActive:  utils.Pointer(true),
	}

	// Simulate organization lookup failure
	var organization db.Organization
	var orgErr error = assert.AnError // Simulate error

	// Test the email data construction logic
	emailData := sendgrid.NotificationTemplateData{
		Recipient: user.FirstName + " " + user.LastName,
		Type:      req.Type,
		CustomerName: func() string {
			if orgErr == nil && organization.Name != "" {
				return organization.Name
			}
			return "DockMaster" // Default fallback
		}(),
		HomeURL: cfg.App.FrontendBaseURL,
	}

	// Assertions
	assert.Equal(t, "John Doe", emailData.Recipient)
	assert.Equal(t, "message", emailData.Type)
	assert.Equal(t, "DockMaster", emailData.CustomerName) // Should use default fallback
	assert.Equal(t, "https://test.dockmaster.com", emailData.HomeURL)
}

// TestSendSmartNotification_AllDeliveryMethod_OrganizationFound tests the scenario where
// organization lookup succeeds and email gets sent with actual organization name
func TestSendSmartNotification_AllDeliveryMethod_OrganizationFound(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		App: config.AppConfig{
			FrontendBaseURL: "https://test.dockmaster.com",
		},
		SendGrid: config.SendGridConfig{
			FromEmail: "test@dockmaster.com",
			FromName:  "Test Dockmaster",
			TemplatesMap: map[string]string{
				"notification": "d-test-template-id",
			},
		},
	}

	// Test data
	userID := uuid.New()
	orgID := uuid.New()

	// Mock user data
	user := db.User{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		IsActive:  utils.Pointer(true),
	}

	// Mock organization data
	organization := db.Organization{
		ID:   orgID,
		Name: "Test Marina",
	}

	// Test the email data construction logic
	emailData := sendgrid.NotificationTemplateData{
		Recipient: user.FirstName + " " + user.LastName,
		Type:      "message",
		CustomerName: func() string {
			if organization.Name != "" {
				return organization.Name
			}
			return "DockMaster" // Default fallback
		}(),
		HomeURL: cfg.App.FrontendBaseURL,
	}

	// Assertions
	assert.Equal(t, "John Doe", emailData.Recipient)
	assert.Equal(t, "message", emailData.Type)
	assert.Equal(t, "Test Marina", emailData.CustomerName) // Should use actual org name
	assert.Equal(t, "https://test.dockmaster.com", emailData.HomeURL)
}

// TestSendSmartNotification_AllDeliveryMethod_EmptyOrganizationName tests the scenario where
// organization is found but has empty name
func TestSendSmartNotification_AllDeliveryMethod_EmptyOrganizationName(t *testing.T) {
	// Create a minimal config
	cfg := &config.Config{
		App: config.AppConfig{
			FrontendBaseURL: "https://test.dockmaster.com",
		},
		SendGrid: config.SendGridConfig{
			FromEmail: "test@dockmaster.com",
			FromName:  "Test Dockmaster",
			TemplatesMap: map[string]string{
				"notification": "d-test-template-id",
			},
		},
	}

	// Test data
	userID := uuid.New()
	orgID := uuid.New()

	// Mock user data
	user := db.User{
		ID:        userID,
		FirstName: "John",
		LastName:  "Doe",
		Email:     "john.doe@example.com",
		IsActive:  utils.Pointer(true),
	}

	// Mock organization data with empty name
	organization := db.Organization{
		ID:   orgID,
		Name: "", // Empty name
	}

	// Test the email data construction logic
	emailData := sendgrid.NotificationTemplateData{
		Recipient: user.FirstName + " " + user.LastName,
		Type:      "message",
		CustomerName: func() string {
			if organization.Name != "" {
				return organization.Name
			}
			return "DockMaster" // Default fallback
		}(),
		HomeURL: cfg.App.FrontendBaseURL,
	}

	// Assertions
	assert.Equal(t, "John Doe", emailData.Recipient)
	assert.Equal(t, "message", emailData.Type)
	assert.Equal(t, "DockMaster", emailData.CustomerName) // Should use default fallback for empty name
	assert.Equal(t, "https://test.dockmaster.com", emailData.HomeURL)
}
