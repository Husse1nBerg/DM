package main

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/guard"
	"github.com/dockworks/dm-web-backend/internal/pg"
)

func main() {
	fmt.Println("🧪 Testing Permission Service Concurrent Access...")

	// Load configuration
	cfg := config.New()

	// Use test database if available, otherwise use main database
	var dbConfig *config.DBConfig
	if cfg.TestDB.Host != "" {
		dbConfig = &cfg.TestDB
		fmt.Println("Using test database")
	} else {
		dbConfig = &cfg.DB
		fmt.Println("Using main database")
	}

	// Create database connection
	dbService := pg.NewConnection(dbConfig)
	defer dbService.Close()

	// Get queries instance
	queries := dbService.Queries()
	ctx := context.Background()

	// Get real data from database
	fmt.Println("📊 Fetching test data from database...")
	userID1, userID2, marinaID, err := getTestData(ctx, queries)
	if err != nil {
		log.Fatalf("Failed to get test data: %v", err)
	}

	fmt.Printf("✅ Found test data:\n")
	fmt.Printf("   User1: %s\n", userID1)
	fmt.Printf("   User2: %s\n", userID2)
	fmt.Printf("   Marina: %s\n", marinaID)

	// Create permission service
	fmt.Println("🔐 Creating permission service...")
	permissionService, err := guard.NewPermissionService(queries)
	if err != nil {
		log.Fatalf("Failed to create permission service: %v", err)
	}

	// Run concurrent access test
	fmt.Println("🚀 Running concurrent access test...")
	err = runConcurrencyTest(ctx, permissionService, userID1, userID2, marinaID)
	if err != nil {
		log.Fatalf("Concurrency test failed: %v", err)
	}

	// Run race detection test
	fmt.Println("🏁 Running race detection test...")
	err = runRaceDetectionTest(ctx, permissionService, userID1, marinaID)
	if err != nil {
		log.Fatalf("Race detection test failed: %v", err)
	}

	fmt.Println("✅ All tests passed! No race conditions detected.")
}

func getTestData(ctx context.Context, queries *db.Queries) (userID1, userID2, marinaID string, err error) {
	// Get all active users
	users, err := queries.GetAllUsers(ctx)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get users: %w", err)
	}

	if len(users) == 0 {
		return "", "", "", fmt.Errorf("no users found in database")
	}

	// Get at least 2 users for testing
	userID1 = users[0].ID.String()
	if len(users) > 1 {
		userID2 = users[1].ID.String()
	} else {
		// Use the same user twice if only one exists
		userID2 = userID1
		fmt.Println("⚠️  Only one user found, using same user for both test cases")
	}

	// Get all active marinas
	marinas, err := queries.GetAllMarinas(ctx)
	if err != nil {
		return "", "", "", fmt.Errorf("failed to get marinas: %w", err)
	}

	if len(marinas) == 0 {
		return "", "", "", fmt.Errorf("no marinas found in database")
	}

	// Use the first marina
	marinaID = marinas[0].ID.String()

	// Verify that users are associated with the marina
	// Check if user1 has access to the marina
	user1Marina, err := queries.GetUserRoleInMarina(ctx, db.GetUserRoleInMarinaParams{
		ID:       users[0].ID,
		MarinaID: marinas[0].ID,
	})
	if err != nil {
		// User might not be directly associated, but let's try anyway
		fmt.Printf("⚠️  User1 might not be associated with marina (will test anyway): %v\n", err)
	} else {
		fmt.Printf("✅ User1 has role in marina: %s\n", user1Marina)
	}

	return userID1, userID2, marinaID, nil
}

func runConcurrencyTest(ctx context.Context, permissionService *guard.PermissionService, userID1, userID2, marinaID string) error {
	fmt.Println("   Testing with 50 goroutines, 100 permission checks each...")

	const numGoroutines = 50
	const numChecksPerGoroutine = 100

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error
	var totalChecks int
	var successfulChecks int

	startTime := time.Now()

	// Launch goroutines
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			// Alternate between users
			userID := userID1
			if goroutineID%2 == 0 {
				userID = userID2
			}

			localSuccessCount := 0

			// Perform permission checks
			for j := 0; j < numChecksPerGoroutine; j++ {
				objects := []string{"boats", "work_orders", "dockage", "fuel"}
				actions := []string{"read", "write", "create", "delete"}

				// Test a few different combinations
				objectIdx := (goroutineID + j) % len(objects)
				actionIdx := (goroutineID + j) % len(actions)

				_, err := permissionService.CanAccess(
					ctx,
					userID,
					marinaID,
					objects[objectIdx],
					actions[actionIdx],
				)

				if err != nil {
					mu.Lock()
					errors = append(errors, fmt.Errorf("goroutine %d, check %d: %w", goroutineID, j, err))
					mu.Unlock()
					return
				}

				localSuccessCount++
			}

			mu.Lock()
			totalChecks += numChecksPerGoroutine
			successfulChecks += localSuccessCount
			mu.Unlock()
		}(i)
	}

	// Wait for completion with timeout
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		duration := time.Since(startTime)

		if len(errors) > 0 {
			fmt.Printf("❌ Concurrency test failed with %d errors:\n", len(errors))
			for i, err := range errors {
				if i < 5 { // Show first 5 errors
					fmt.Printf("   %v\n", err)
				}
			}
			if len(errors) > 5 {
				fmt.Printf("   ... and %d more errors\n", len(errors)-5)
			}
			return fmt.Errorf("concurrency test failed with %d errors", len(errors))
		}

		expectedChecks := numGoroutines * numChecksPerGoroutine
		fmt.Printf("✅ Concurrency test passed!\n")
		fmt.Printf("   Total checks: %d/%d\n", successfulChecks, expectedChecks)
		fmt.Printf("   Duration: %v\n", duration)
		fmt.Printf("   Rate: %.0f checks/second\n", float64(successfulChecks)/duration.Seconds())

		return nil

	case <-time.After(60 * time.Second):
		return fmt.Errorf("concurrency test timed out after 60 seconds")
	}
}

func runRaceDetectionTest(ctx context.Context, permissionService *guard.PermissionService, userID, marinaID string) error {
	fmt.Println("   Testing with 100 goroutines doing rapid checks...")

	const numGoroutines = 100
	const checksPerGoroutine = 100

	var wg sync.WaitGroup
	var mu sync.Mutex
	var errors []error

	startTime := time.Now()

	// Launch many goroutines doing rapid permission checks
	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < checksPerGoroutine; j++ {
				// Rapid-fire permission checks
				_, err := permissionService.CanAccess(
					ctx,
					userID,
					marinaID,
					"boats",
					"read",
				)

				if err != nil {
					mu.Lock()
					errors = append(errors, err)
					mu.Unlock()
					return
				}

				// No sleep - maximum concurrency stress
			}
		}(i)
	}

	wg.Wait()
	duration := time.Since(startTime)

	if len(errors) > 0 {
		fmt.Printf("❌ Race detection test failed with %d errors\n", len(errors))
		return fmt.Errorf("race detection test failed")
	}

	totalChecks := numGoroutines * checksPerGoroutine
	fmt.Printf("✅ Race detection test passed!\n")
	fmt.Printf("   Total rapid checks: %d\n", totalChecks)
	fmt.Printf("   Duration: %v\n", duration)
	fmt.Printf("   Rate: %.0f checks/second\n", float64(totalChecks)/duration.Seconds())

	return nil
}
