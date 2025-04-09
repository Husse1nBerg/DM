package tests

import (
	"context"
	"testing"

	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/tests/testutil"
	"github.com/stretchr/testify/assert"
)

// TestDatabaseTransaction demonstrates how to use a transaction for database tests
func TestDatabaseTransaction(t *testing.T) {
	// Skip if testDB is nil (will be handled by TestMain)
	if testDB == nil {
		t.Skip("Test database not initialized")
	}

	// Run test in a transaction that will be rolled back
	testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
		// Here you would typically:
		// 1. Set up test data
		// 2. Execute the operation being tested
		// 3. Verify the results

		// Example (assuming you have these queries in your SQLC definitions):
		// Create a test organization
		// org, err := q.CreateOrganization(ctx, db.CreateOrganizationParams{
		//     Name: "Test Organization",
		//     Email: "test@example.com",
		// })
		// assert.NoError(t, err)
		// assert.NotEmpty(t, org.ID)

		// This is just a placeholder assertion since we're not executing actual operations
		assert.True(t, true, "Database transaction was executed")
	})
}

// ExampleMultipleTestsWithTransactions shows how to use transactions for test isolation
func TestMultipleOperationsWithTransactions(t *testing.T) {
	// Skip if testDB is nil (will be handled by TestMain)
	if testDB == nil {
		t.Skip("Test database not initialized")
	}

	// First transaction - completely isolated from other tests
	t.Run("First operation", func(t *testing.T) {
		testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
			// Create data specific to this test
			// The data will not be visible to other tests and won't persist after the test

			// Example (pseudocode):
			// _, err := q.CreateUser(ctx, db.CreateUserParams{...})
			// assert.NoError(t, err)

			assert.True(t, true, "First operation completed")
		})
	})

	// Second transaction - completely isolated from the first
	t.Run("Second operation", func(t *testing.T) {
		testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
			// Data created in the first test is not visible here
			// Any data created here will be rolled back after the test

			// Example (pseudocode):
			// users, err := q.ListUsers(ctx)
			// assert.NoError(t, err)
			// assert.Empty(t, users, "No users should exist from previous test")

			assert.True(t, true, "Second operation completed")
		})
	})
}
