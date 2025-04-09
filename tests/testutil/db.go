package testutil

import (
	"context"
	"log"
	"sync"
	"testing"

	"github.com/dockworks/dm-web-backend/internal/config"
	"github.com/dockworks/dm-web-backend/internal/db"
	"github.com/dockworks/dm-web-backend/internal/pg"
	"github.com/jackc/pgx/v5"
)

var (
	testDB       pg.DBService
	testDBConfig *config.DBConfig
	once         sync.Once
)

// InitTestDB initializes the test database connection
func InitTestDB() pg.DBService {
	once.Do(func() {
		// Ensure test environment is set up
		SetupTestEnv()

		// Load configuration
		cfg := config.New()
		testDBConfig = &cfg.TestDB

		// Ensure test database is properly configured
		if testDBConfig.Host == "" || testDBConfig.Name == "" {
			log.Fatalf("Test database not configured. Please check TEST_DB_* environment variables")
		}

		// Attempt to create connection
		testDB = pg.NewConnection(testDBConfig)

		// Basic health check
		health := testDB.Health()
		if status, ok := health["status"]; !ok || status != "up" {
			log.Printf("Warning: Test database health check failed: %v", health)
			log.Printf("Using test database: %s@%s:%s/%s",
				testDBConfig.User, testDBConfig.Host, testDBConfig.Port, testDBConfig.Name)
		} else {
			log.Printf("Test database connection successful")
		}
	})

	return testDB
}

// GetTestQueries returns the sqlc queries instance for the test database
func GetTestQueries() *db.Queries {
	if testDB == nil {
		InitTestDB()
	}
	return testDB.Queries()
}

// CleanupTestDB closes the test database connection
func CleanupTestDB() {
	if testDB != nil {
		testDB.Close()
		testDB = nil
	}
}

// SetupTestSuite sets up the test database for a test suite
func SetupTestSuite(t *testing.T) pg.DBService {
	t.Helper()
	return InitTestDB()
}

// TeardownTestSuite cleans up the test database after a test suite
func TeardownTestSuite(t *testing.T) {
	t.Helper()
	CleanupTestDB()
}

// WithTestTransaction runs a function within a database transaction that is rolled back
// after the function completes, ensuring tests don't affect each other
func WithTestTransaction(t *testing.T, fn func(ctx context.Context, q *db.Queries)) {
	t.Helper()

	if testDB == nil {
		InitTestDB()
	}

	// Get a connection directly from the database using the test config
	ctx := context.Background()
	conn, err := pgx.Connect(ctx, testDBConfig.Addr())
	if err != nil {
		t.Fatalf("Failed to connect to database: %v", err)
	}
	defer conn.Close(ctx)

	// Begin transaction
	tx, err := conn.Begin(ctx)
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}

	// Create queries with transaction
	q := db.New(tx)

	// Run the test function
	fn(ctx, q)

	// Always rollback - this ensures test isolation
	err = tx.Rollback(ctx)
	if err != nil {
		t.Fatalf("Failed to rollback transaction: %v", err)
	}
}
