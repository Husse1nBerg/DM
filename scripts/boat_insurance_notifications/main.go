package main

import (
	"context"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	sqlc "github.com/dockworks/dm-web-backend/internal/db"
	conn "github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/dockworks/dm-web-backend/pkg/notifications"
	"github.com/dockworks/dm-web-backend/pkg/redis"
	"github.com/google/uuid"
)

// BoatInsuranceCheck represents a boat that needs insurance notification
type BoatInsuranceCheck struct {
	Boat         *dme.Boat
	Customer     *dme.Customer
	DaysToExpiry int
	ExpDate      time.Time
	MarinaID     uuid.UUID
	OrgID        uuid.UUID
	SystemID     string
}

// NotificationSummary tracks the results of sending notifications
type NotificationSummary struct {
	TotalBoatsChecked         int
	BoatsWithExpiredInsurance int
	Boats10DaysToExpiry       int
	Boats5DaysToExpiry        int
	Boats1DayToExpiry         int
	NotificationsSent         int
	NotificationsFailed       int
	Errors                    []string
}

func main() {
	log.Println("Starting boat insurance expiration notification script...")

	// Load configuration
	cfg := config.New()
	ctx := context.Background()

	// Set up database connection
	db := conn.NewConnection(&cfg.DB)
	q := db.Queries()

	// Set up logger
	logger := logger.NewLogger(cfg.Logger)

	// Set up Redis client for notifications
	redisClient := redis.NewClient(cfg.Redis, logger)

	// Set up notification service
	notificationService := notifications.NewNotificationService(q, redisClient, logger)

	// Set up DME client
	dmeClient := dme.NewClientFromConfig(cfg, logger, db)

	// Run the insurance check
	summary, err := checkInsuranceExpirations(ctx, q, dmeClient, notificationService, logger)
	if err != nil {
		log.Fatalf("Failed to check insurance expirations: %v", err)
	}

	// Print summary
	printSummary(summary)

	log.Println("Boat insurance expiration notification script completed successfully")
}

// checkInsuranceExpirations performs the main logic of checking and notifying
func checkInsuranceExpirations(
	ctx context.Context,
	q *sqlc.Queries,
	dmeClient *dme.Client,
	notificationService *notifications.NotificationService,
	logger *logger.Logger,
) (*NotificationSummary, error) {
	summary := &NotificationSummary{}

	// Get all active marinas
	marinas, err := q.GetAllMarinas(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get marinas: %w", err)
	}

	log.Printf("Found %d marinas to check", len(marinas))

	// Check each marina for boats with expiring insurance
	for _, marina := range marinas {
		// Skip inactive marinas or those without system ID
		if marina.IsActive != nil && !*marina.IsActive {
			log.Printf("Skipping inactive marina: %s", marina.Name)
			continue
		}

		if marina.SystemID == nil {
			log.Printf("Skipping marina without system ID: %s", marina.Name)
			continue
		}

		log.Printf("Checking marina: %s (ID: %s)", marina.Name, marina.ID)

		marinaSummary, err := checkMarinaInsurance(ctx, q, dmeClient, notificationService, marina, logger)
		if err != nil {
			errorMsg := fmt.Sprintf("Error checking marina %s: %v", marina.Name, err)
			log.Printf(errorMsg)
			summary.Errors = append(summary.Errors, errorMsg)
			continue
		}

		// Aggregate summary data
		summary.TotalBoatsChecked += marinaSummary.TotalBoatsChecked
		summary.BoatsWithExpiredInsurance += marinaSummary.BoatsWithExpiredInsurance
		summary.Boats10DaysToExpiry += marinaSummary.Boats10DaysToExpiry
		summary.Boats5DaysToExpiry += marinaSummary.Boats5DaysToExpiry
		summary.Boats1DayToExpiry += marinaSummary.Boats1DayToExpiry
		summary.NotificationsSent += marinaSummary.NotificationsSent
		summary.NotificationsFailed += marinaSummary.NotificationsFailed
		summary.Errors = append(summary.Errors, marinaSummary.Errors...)
	}

	return summary, nil
}

// checkMarinaInsurance checks boats for a specific marina
func checkMarinaInsurance(
	ctx context.Context,
	q *sqlc.Queries,
	dmeClient *dme.Client,
	notificationService *notifications.NotificationService,
	marina sqlc.Marina,
	logger *logger.Logger,
) (*NotificationSummary, error) {
	summary := &NotificationSummary{}

	// Get all boats for this marina (paginated approach)
	page := 1
	pageSize := 100

	for {
		boatList, err := dmeClient.BoatsList(ctx, page, pageSize, marina.OrganizationID, *marina.SystemID)
		if err != nil {
			return nil, fmt.Errorf("failed to get boats list for marina %s: %w", marina.Name, err)
		}

		if len(boatList.Content) == 0 {
			break // No more boats
		}

		log.Printf("Processing page %d with %d boats for marina %s", page, len(boatList.Content), marina.Name)

		// Check each boat for insurance expiration
		for _, boat := range boatList.Content {
			summary.TotalBoatsChecked++

			// Get full boat details to access insurance information
			fullBoat, err := dmeClient.RetrieveBoatByID(ctx, boat.ID, marina.OrganizationID, *marina.SystemID)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to get boat details for %s (%s): %v", boat.Name, boat.ID, err)
				log.Printf(errorMsg)
				summary.Errors = append(summary.Errors, errorMsg)
				continue
			}

			// Check if insurance is expiring
			insuranceCheck, err := checkBoatInsurance(fullBoat, marina.ID, marina.OrganizationID, *marina.SystemID)
			if err != nil {
				errorMsg := fmt.Sprintf("Error checking insurance for boat %s: %v", fullBoat.Name, err)
				log.Printf(errorMsg)
				summary.Errors = append(summary.Errors, errorMsg)
				continue
			}

			// If no insurance check needed, continue
			if insuranceCheck == nil {
				continue
			}

			// Update summary based on expiry days
			switch insuranceCheck.DaysToExpiry {
			case 10:
				summary.Boats10DaysToExpiry++
			case 5:
				summary.Boats5DaysToExpiry++
			case 1:
				summary.Boats1DayToExpiry++
			default:
				if insuranceCheck.DaysToExpiry <= 0 {
					summary.BoatsWithExpiredInsurance++
				}
			}

			// Get customer information
			customer, err := dmeClient.CustomerRetrieve(ctx, fullBoat.OwnerID, marina.OrganizationID, *marina.SystemID)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to get customer info for boat %s owner %s: %v", fullBoat.Name, fullBoat.OwnerID, err)
				log.Printf(errorMsg)
				summary.Errors = append(summary.Errors, errorMsg)
				continue
			}

			insuranceCheck.Customer = customer

			// Send notification
			err = sendInsuranceNotification(ctx, q, notificationService, insuranceCheck, marina, logger)
			if err != nil {
				errorMsg := fmt.Sprintf("Failed to send notification for boat %s: %v", fullBoat.Name, err)
				log.Printf(errorMsg)
				summary.Errors = append(summary.Errors, errorMsg)
				summary.NotificationsFailed++
			} else {
				summary.NotificationsSent++
				log.Printf("Sent insurance notification for boat %s (expires in %d days)", fullBoat.Name, insuranceCheck.DaysToExpiry)
			}
		}

		// Check if we've reached the last page
		if page >= boatList.MaxPages {
			break
		}

		page++
	}

	return summary, nil
}

// checkBoatInsurance checks if a boat's insurance is expiring in 10, 5, or 1 days
func checkBoatInsurance(boat *dme.Boat, marinaID, orgID uuid.UUID, systemID string) (*BoatInsuranceCheck, error) {
	// Check if insurance expiration date exists and is not empty
	if boat.InsuranceExpDate == "" {
		return nil, nil // No insurance date to check
	}

	// Parse the insurance expiration date
	// The date might be in various formats, try common ones
	var expDate time.Time
	var err error

	// Try common date formats
	dateFormats := []string{
		"2006-01-02",           // YYYY-MM-DD
		"01/02/2006",           // MM/DD/YYYY
		"1/2/2006",             // M/D/YYYY
		"02-Jan-2006",          // DD-Mon-YYYY
		"2006-01-02T15:04:05Z", // ISO 8601
		"2006-01-02 15:04:05",  // YYYY-MM-DD HH:MM:SS
	}

	for _, format := range dateFormats {
		expDate, err = time.Parse(format, strings.TrimSpace(boat.InsuranceExpDate))
		if err == nil {
			break
		}
	}

	if err != nil {
		return nil, fmt.Errorf("could not parse insurance expiration date '%s': %w", boat.InsuranceExpDate, err)
	}

	// Calculate days until expiration
	now := time.Now()
	daysUntilExp := int(expDate.Sub(now).Hours() / 24)

	// Only notify for 10, 5, and 1 days before expiration
	if daysUntilExp == 10 || daysUntilExp == 5 || daysUntilExp == 1 {
		return &BoatInsuranceCheck{
			Boat:         boat,
			DaysToExpiry: daysUntilExp,
			ExpDate:      expDate,
			MarinaID:     marinaID,
			OrgID:        orgID,
			SystemID:     systemID,
		}, nil
	}

	return nil, nil // Not a notification day
}

// sendInsuranceNotification sends a notification to users associated with the boat owner
func sendInsuranceNotification(
	ctx context.Context,
	q *sqlc.Queries,
	notificationService *notifications.NotificationService,
	insuranceCheck *BoatInsuranceCheck,
	marina sqlc.Marina,
	logger *logger.Logger,
) error {
	// Find users associated with this customer
	users, err := findUsersForCustomer(ctx, q, insuranceCheck.Customer.ID, insuranceCheck.MarinaID)
	if err != nil {
		return fmt.Errorf("failed to find users for customer %s: %w", insuranceCheck.Customer.ID, err)
	}

	if len(users) == 0 {
		log.Printf("No users found for customer %s (boat: %s)", insuranceCheck.Customer.ID, insuranceCheck.Boat.Name)
		return nil
	}

	// Create notification content
	title := "Boat Insurance Expiring Soon"
	var content string
	var priority string

	switch insuranceCheck.DaysToExpiry {
	case 10:
		content = fmt.Sprintf("Your boat '%s' insurance with %s expires in 10 days on %s. Please renew to avoid any interruptions.",
			insuranceCheck.Boat.Name, insuranceCheck.Boat.InsuranceCompany, insuranceCheck.ExpDate.Format("January 2, 2006"))
		priority = "normal"
	case 5:
		content = fmt.Sprintf("Your boat '%s' insurance with %s expires in 5 days on %s. Please renew soon to avoid any interruptions.",
			insuranceCheck.Boat.Name, insuranceCheck.Boat.InsuranceCompany, insuranceCheck.ExpDate.Format("January 2, 2006"))
		priority = "high"
	case 1:
		content = fmt.Sprintf("URGENT: Your boat '%s' insurance with %s expires tomorrow on %s. Please renew immediately.",
			insuranceCheck.Boat.Name, insuranceCheck.Boat.InsuranceCompany, insuranceCheck.ExpDate.Format("January 2, 2006"))
		priority = "high"
	default:
		content = fmt.Sprintf("Your boat '%s' insurance with %s has expired or is expiring soon. Please renew as soon as possible.",
			insuranceCheck.Boat.Name, insuranceCheck.Boat.InsuranceCompany)
		priority = "high"
	}

	// Send notification to each user
	for _, user := range users {
		notificationData := map[string]interface{}{
			"boatID":           insuranceCheck.Boat.ID,
			"boatName":         insuranceCheck.Boat.Name,
			"insuranceCompany": insuranceCheck.Boat.InsuranceCompany,
			"expirationDate":   insuranceCheck.ExpDate.Format("2006-01-02"),
			"daysToExpiry":     insuranceCheck.DaysToExpiry,
			"customerID":       insuranceCheck.Customer.ID,
		}

		smartNotificationReq := notifications.SmartNotificationRequest{
			UserID:         user.ID,
			OrganizationID: insuranceCheck.OrgID,
			MarinaID:       insuranceCheck.MarinaID,
			Type:           "system",
			Title:          title,
			Content:        content,
			Data:           notificationData,
			Priority:       &priority,
		}

		// Send the notification
		result, err := notificationService.SendSmartNotification(ctx, smartNotificationReq)
		if err != nil {
			return fmt.Errorf("failed to send notification to user %s: %w", user.ID, err)
		}

		if result != nil && len(result.Errors) > 0 {
			log.Printf("Warning: Some notification delivery methods failed for user %s: %v", user.ID, result.Errors)
		}

		log.Printf("Sent insurance notification to user %s (%s) for boat %s", user.Username, user.Email, insuranceCheck.Boat.Name)
	}

	return nil
}

// findUsersForCustomer finds users associated with a customer ID in a marina
func findUsersForCustomer(ctx context.Context, q *sqlc.Queries, customerID string, marinaID uuid.UUID) ([]sqlc.User, error) {
	// Try to find users by customer ID in user_marinas table
	users, err := q.GetMarinaUsersList(ctx, sqlc.GetMarinaUsersListParams{
		MarinaID:   marinaID,
		IsCustomer: nil, // Get all users, we'll filter by customer ID
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get marina users: %w", err)
	}

	// Filter users by customer ID
	var matchingUsers []sqlc.User
	for _, user := range users {
		// Check if user has matching customer ID
		if user.CustomerID != nil && *user.CustomerID == customerID {
			matchingUsers = append(matchingUsers, sqlc.User{
				ID:             user.ID,
				Username:       user.Username,
				FirstName:      user.FirstName,
				LastName:       user.LastName,
				Email:          user.Email,
				Phone:          user.Phone,
				OrganizationID: user.OrganizationID,
				MarinaID:       user.MarinaID,
				RoleID:         user.RoleID,
				IsActive:       user.IsActive,
				CustomerID:     user.CustomerID,
				IsCustomer:     user.IsCustomer,
			})
		}
	}

	return matchingUsers, nil
}

// printSummary prints a summary of the insurance check results
func printSummary(summary *NotificationSummary) {
	log.Println("=== BOAT INSURANCE NOTIFICATION SUMMARY ===")
	log.Printf("Total boats checked: %d", summary.TotalBoatsChecked)
	log.Printf("Boats with expired insurance: %d", summary.BoatsWithExpiredInsurance)
	log.Printf("Boats expiring in 10 days: %d", summary.Boats10DaysToExpiry)
	log.Printf("Boats expiring in 5 days: %d", summary.Boats5DaysToExpiry)
	log.Printf("Boats expiring in 1 day: %d", summary.Boats1DayToExpiry)
	log.Printf("Notifications sent: %d", summary.NotificationsSent)
	log.Printf("Notifications failed: %d", summary.NotificationsFailed)

	if len(summary.Errors) > 0 {
		log.Printf("Errors encountered: %d", len(summary.Errors))
		for i, err := range summary.Errors {
			if i < 10 { // Limit error output
				log.Printf("  Error %d: %s", i+1, err)
			}
		}
		if len(summary.Errors) > 10 {
			log.Printf("  ... and %d more errors", len(summary.Errors)-10)
		}
	}
	log.Println("=============================================")
}
