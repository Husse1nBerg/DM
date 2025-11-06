package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"go.uber.org/zap"
)

func main() {
	// Parse command line flags
	dryRun := flag.Bool("dry-run", false, "Run in dry-run mode (don't make changes)")
	flag.Parse()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize logger
	logger.InitLogger(cfg.Logger)
	defer logger.Zap.Sync()

	// Connect to database
	conn, err := pg.NewConnection(cfg.DB)
	if err != nil {
		logger.Zap.Fatal("Failed to connect to database", zap.Error(err))
	}
	defer conn.Close()

	queries := db.New(conn)

	ctx := context.Background()

	logger.Zap.Info("Starting migration to add 'service' notification preference to existing users")

	// Get all users (process in batches if needed)
	// First, get a batch of users
	limit := int32(10000)
	offset := int32(0)
	users, err := queries.GetAllUsersPaginated(ctx, db.GetAllUsersPaginatedParams{
		Limit:  limit,
		Offset: offset,
	})
	if err != nil {
		logger.Zap.Fatal("Failed to get users", zap.Error(err))
	}

	logger.Zap.Info("Found users to process", zap.Int("count", len(users)))

	successCount := 0
	skipCount := 0
	errorCount := 0

	for _, user := range users {
		// Check if user already has 'service' notification preference
		existingPref, err := queries.GetNotificationPreference(ctx, db.GetNotificationPreferenceParams{
			UserID:           user.ID,
			NotificationType: "service",
		})

		if err == nil && existingPref != nil {
			// Preference already exists, skip
			skipCount++
			logger.Zap.Debug("User already has 'service' notification preference",
				zap.String("user_id", user.ID.String()),
				zap.String("email", user.Email))
			continue
		}

		// Create the preference
		if !*dryRun {
			enabled := true
			deliveryMethod := "all"

			_, err := queries.CreateNotificationPreference(ctx, db.CreateNotificationPreferenceParams{
				UserID:           user.ID,
				NotificationType: "service",
				Enabled:          &enabled,
				DeliveryMethod:   &deliveryMethod,
			})

			if err != nil {
				errorCount++
				logger.Zap.Error("Failed to create 'service' notification preference for user",
					zap.String("user_id", user.ID.String()),
					zap.String("email", user.Email),
					zap.Error(err))
				continue
			}

			successCount++
			logger.Zap.Info("Added 'service' notification preference for user",
				zap.String("user_id", user.ID.String()),
				zap.String("email", user.Email))
		} else {
			successCount++
			logger.Zap.Info("[DRY RUN] Would add 'service' notification preference for user",
				zap.String("user_id", user.ID.String()),
				zap.String("email", user.Email))
		}
	}

	logger.Zap.Info("Migration completed",
		zap.Int("total_users", len(users)),
		zap.Int("successful", successCount),
		zap.Int("skipped", skipCount),
		zap.Int("errors", errorCount),
		zap.Bool("dry_run", *dryRun))

	if *dryRun {
		fmt.Println("\nThis was a dry run. Use without --dry-run flag to apply changes.")
		os.Exit(0)
	}

	if errorCount > 0 {
		fmt.Printf("\nMigration completed with %d errors. Check logs for details.\n", errorCount)
		os.Exit(1)
	}

	fmt.Println("\nMigration completed successfully!")
}
