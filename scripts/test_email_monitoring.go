package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/dockworks/dm-web-backend/pkg/notifications"
	"github.com/dockworks/dm-web-backend/pkg/redis"
	"github.com/dockworks/dm-web-backend/pkg/sendgrid"
	"github.com/google/uuid"
)

func main() {
	fmt.Println("🧪 Testing SendSmartNotification with detailed email monitoring...")

	// Load configuration
	cfg := config.New()

	// Initialize logger
	logger := logger.NewLogger(cfg.Logger)

	// Initialize database connection
	dbConn := pg.NewConnection(&cfg.DB)
	defer dbConn.Close()

	// Initialize Redis client
	redisClient := redis.NewClient(cfg.Redis, logger)
	defer redisClient.Close()

	// Initialize SendGrid client
	sendgridClient := sendgrid.NewClient(cfg)

	// Initialize notification service
	notificationService := notifications.NewNotificationService(
		dbConn.Queries(),
		redisClient,
		logger,
		sendgridClient,
		cfg,
	)

	// Create a request that will trigger email-only delivery
	req := notifications.SmartNotificationRequest{
		UserID:         uuid.New(),
		OrganizationID: uuid.New(),
		MarinaID:       uuid.New(),
		Type:           "message",
		Title:          "Test Email - SendSmartNotification Fix",
		Content:        "This is a test email to verify that the SendSmartNotification function is working correctly with the 'all' delivery method.",
		Data: map[string]interface{}{
			"test":      "data",
			"timestamp": time.Now().Format(time.RFC3339),
		},
		EmailData: &notifications.EmailNotificationData{
			To:      []string{"abiezer.matos@dockmaster.com"},
			Subject: "Test Email - SendSmartNotification Fix Verification",
		},
	}

	// Create context with longer timeout
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("📧 Sending test email to abiezer.matos@dockmaster.com...")
	fmt.Printf("User ID: %s\n", req.UserID)
	fmt.Printf("Organization ID: %s\n", req.OrganizationID)
	fmt.Printf("Marina ID: %s\n", req.MarinaID)

	// Send the notification
	result, err := notificationService.SendSmartNotification(ctx, req)
	if err != nil {
		log.Fatalf("❌ Failed to send smart notification: %v", err)
	}

	fmt.Println("\n📊 === Notification Result ===")
	fmt.Printf("User ID: %s\n", result.UserID)
	fmt.Printf("System Delivered: %t\n", result.SystemDelivered)
	fmt.Printf("Email Delivered: %t\n", result.EmailDelivered)
	if result.NotificationID != nil {
		fmt.Printf("Notification ID: %s\n", *result.NotificationID)
	}
	if len(result.Errors) > 0 {
		fmt.Println("❌ Errors:")
		for _, err := range result.Errors {
			fmt.Printf("  - %s\n", err)
		}
	}

	if result.EmailDelivered {
		fmt.Println("\n✅ Email was successfully queued for delivery!")
		fmt.Println("📬 Please check your email at abiezer.matos@dockmaster.com")
		fmt.Println("   Subject: Test Email - SendSmartNotification Fix Verification")

		// Wait a bit and check for any delivery status updates
		fmt.Println("\n⏳ Waiting 10 seconds for delivery status updates...")
		time.Sleep(10 * time.Second)

		// Check SendGrid configuration
		fmt.Println("\n🔧 SendGrid Configuration Check:")
		fmt.Printf("API Key configured: %t\n", cfg.SendGrid.APIKey != "")
		fmt.Printf("From Email: %s\n", cfg.SendGrid.FromEmail)
		fmt.Printf("From Name: %s\n", cfg.SendGrid.FromName)
		fmt.Printf("Notification Template ID: %s\n", cfg.SendGrid.TemplatesMap["notification"])

		if cfg.SendGrid.APIKey == "" {
			fmt.Println("❌ WARNING: SendGrid API Key is not configured!")
		}
		if cfg.SendGrid.TemplatesMap["notification"] == "" {
			fmt.Println("❌ WARNING: Notification template ID is not configured!")
		}
	} else {
		fmt.Println("\n❌ Email delivery failed. Check the errors above.")
	}

	fmt.Println("\n🔍 This test verifies that the SendSmartNotification function")
	fmt.Println("   can handle the 'all' delivery method correctly, even when")
	fmt.Println("   database operations fail due to non-existent IDs.")
	fmt.Println("   The email should still be sent even if system notification fails.")
}
