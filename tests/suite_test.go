package tests

import (
	"fmt"
	"os"
	"testing"

	"github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/dockworks/dm-web-backend/tests/testutil"
)

var (
	// testDB holds the database service for all tests
	testDB pg.DBService
)

// TestMain is the entry point for all tests in this package
func TestMain(m *testing.M) {
	// Setup test environment
	setup()

	// Run tests
	code := m.Run()

	// Teardown test environment
	teardown()

	// Exit with the test result code
	os.Exit(code)
}

// setup initializes the test environment
func setup() {
	fmt.Println("Setting up test environment...")

	// Setup test environment variables
	testutil.SetupTestEnv()

	// Initialize the test database
	testDB = testutil.InitTestDB()

	fmt.Println("Test environment setup complete")
}

// teardown cleans up the test environment
func teardown() {
	fmt.Println("Tearing down test environment...")

	// Cleanup the test database
	testutil.CleanupTestDB()

	fmt.Println("Test environment teardown complete")
}

// TestDatabaseConnection verifies the test database connection works
func TestDatabaseConnection(t *testing.T) {
	// Check if the test database is initialized
	if testDB == nil {
		t.Fatal("Test database not initialized")
	}

	// Check if the database is healthy
	health := testDB.Health()
	status, ok := health["status"]

	if !ok || status != "up" {
		t.Fatalf("Database health check failed: %v", health)
	}

	t.Log("Database connection is healthy")
}
