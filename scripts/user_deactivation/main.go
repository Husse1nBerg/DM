package main

import (
	"context"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	conn "github.com/dockworks/dm-web-backend/internal/pg"
)

// DeactivateInactiveUsers deactivates users who haven't logged in for the specified number of days
func DeactivateInactiveUsers() {
	cfg := config.New()
	ctx := context.Background()
	db := conn.NewConnection(&cfg.DB)
	q := db.Queries()

	// Get the number of days from environment variable
	inactiveDaysStr := os.Getenv("INACTIVE_USER_BLOCK_DAYS")
	if inactiveDaysStr == "" {
		log.Fatal("INACTIVE_USER_BLOCK_DAYS environment variable is not set")
	}

	inactiveDays, err := strconv.Atoi(inactiveDaysStr)
	if err != nil {
		log.Fatalf("Invalid INACTIVE_USER_BLOCK_DAYS value: %v", err)
	}

	// Calculate the cutoff date
	cutoffDate := time.Now().UTC().AddDate(0, 0, -inactiveDays)

	// Get all active users
	users, err := q.GetAllUsers(ctx)
	if err != nil {
		log.Fatalf("Failed to get users: %v", err)
	}

	deactivatedCount := 0
	for _, user := range users {
		// Skip if user is already inactive
		if user.IsActive != nil && !*user.IsActive {
			continue
		}

		// Skip if user has never logged in
		if user.LastLogin.Time.IsZero() {
			continue
		}

		// Check if user's last login is before the cutoff date
		if user.LastLogin.Time.Before(cutoffDate) {
			// Deactivate the user
			_, err := q.DeactivateUser(ctx, user.ID)
			if err != nil {
				log.Printf("Failed to deactivate user %s: %v", user.Email, err)
				continue
			}
			deactivatedCount++
			log.Printf("Deactivated user: %s (Last login: %s)", user.Email, user.LastLogin.Time.Format(time.RFC3339))
		}
	}

	log.Printf("Deactivation complete. Deactivated %d users.", deactivatedCount)
}

func main() {
	log.Println("Starting user deactivation...")
	DeactivateInactiveUsers()
	log.Println("User deactivation completed successfully")
}
