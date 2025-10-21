package main

import (
	"context"
	"log"
	"os"
	"strings"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	conn "github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/pkg/dme"
	"github.com/dockworks/dm-web-backend/pkg/logger"
	"github.com/dockworks/dm-web-backend/pkg/utils"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func main() {
	log.Println("Starting customer name backfill script...")

	// Load configuration
	cfg := config.New()
	ctx := context.Background()

	// Set up logger
	logger := logger.NewLogger(cfg.Logger)

	// Connect directly to database for raw SQL access
	dbPool, err := pgxpool.New(ctx, cfg.DB.Addr())
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer dbPool.Close()

	// Set up database connection for DME client
	db := conn.NewConnection(&cfg.DB)
	queries := db.Queries()

	// Initialize DME client
	dmeClient := dme.NewClientFromConfig(cfg, logger, db)

	log.Println("Fetching submissions that need customer name backfill...")

	// Use raw SQL query to find submissions with customer_id but without customer_name
	rows, err := dbPool.Query(ctx, `
		SELECT id, customer_id, organization_id, marina_id
		FROM esign_submissions
		WHERE customer_id IS NOT NULL 
		AND customer_id != ''
		AND customer_name IS NULL
		AND deleted_at IS NULL
	`)
	if err != nil {
		log.Fatalf("Failed to query submissions: %v", err)
	}
	defer rows.Close()

	var submissions []struct {
		ID             uuid.UUID
		CustomerID     string
		OrganizationID uuid.UUID
		MarinaID       uuid.UUID
	}

	for rows.Next() {
		var sub struct {
			ID             uuid.UUID
			CustomerID     string
			OrganizationID uuid.UUID
			MarinaID       uuid.UUID
		}
		if err := rows.Scan(&sub.ID, &sub.CustomerID, &sub.OrganizationID, &sub.MarinaID); err != nil {
			log.Printf("Error scanning row: %v", err)
			continue
		}
		submissions = append(submissions, sub)
	}

	if err := rows.Err(); err != nil {
		log.Fatalf("Error iterating rows: %v", err)
	}

	log.Printf("Found %d submissions to backfill", len(submissions))

	if len(submissions) == 0 {
		log.Println("No submissions to backfill. Exiting.")
		return
	}

	// Process each submission
	successCount := 0
	errorCount := 0
	skippedCount := 0

	for i, sub := range submissions {
		// Rate limiting: wait 100ms between requests to avoid overwhelming DME API
		if i > 0 {
			time.Sleep(100 * time.Millisecond)
		}

		log.Printf("[%d/%d] Processing submission %s with customer_id %s",
			i+1, len(submissions), sub.ID, sub.CustomerID)

		// Get marina to retrieve system ID
		marina, err := queries.GetMarinaByID(ctx, sub.MarinaID)
		if err != nil {
			log.Printf("  ⚠️  Error fetching marina: %v", err)
			errorCount++
			continue
		}

		if marina.SystemID == nil {
			log.Printf("  ⚠️  Marina has no system ID configured, skipping")
			skippedCount++
			continue
		}

		// Fetch customer from DME
		customer, err := dmeClient.CustomerRetrieve(ctx, sub.CustomerID, sub.OrganizationID, *marina.SystemID)
		if err != nil {
			log.Printf("  ❌ Failed to fetch customer from DME: %v", err)
			errorCount++
			continue
		}

		if customer.Name == "" {
			log.Printf("  ⚠️  Customer has no name in DME, skipping")
			skippedCount++
			continue
		}

		// Transform name from "lastname, firstname" to "firstname lastname"
		formattedName := utils.FormatCustomerName(customer.Name)

		// Update submission with customer name
		_, err = dbPool.Exec(ctx, `
			UPDATE esign_submissions
			SET customer_name = $1, updated_at = CURRENT_TIMESTAMP
			WHERE id = $2
		`, formattedName, sub.ID)

		if err != nil {
			log.Printf("  ❌ Failed to update submission: %v", err)
			errorCount++
			continue
		}

		log.Printf("  ✅ Updated with customer name: %s (formatted from: %s)", formattedName, customer.Name)
		successCount++
	}

	// Print summary
	log.Println("\n" + strings.Repeat("=", 60))
	log.Println("Backfill Summary:")
	log.Printf("  Total submissions: %d", len(submissions))
	log.Printf("  ✅ Successfully updated: %d", successCount)
	log.Printf("  ❌ Errors: %d", errorCount)
	log.Printf("  ⚠️  Skipped: %d", skippedCount)
	log.Println(strings.Repeat("=", 60))

	if errorCount > 0 {
		os.Exit(1)
	}
}
