# Testing in Marina Management System

This document describes the testing approach for the Marina Management System backend.

## Test Database Setup

The system uses a dedicated test database configuration, separate from the development and production databases. This ensures that tests can run without affecting real data.

### Environment Variables

To configure the test database, add the following environment variables to your `.env` file:

```
TEST_DB_HOST=localhost
TEST_DB_PORT=5432
TEST_DB_USER=postgres
TEST_DB_PASSWORD=yourpassword
TEST_DB_NAME=marina_test
TEST_DB_SCHEMA=public
```

### Test Suite Structure

The tests use Go's standard testing package with some additional utilities:

1. **TestMain**: The `TestMain` function in `suite_test.go` initializes the test database connection once before all tests run and cleans up after all tests are complete.

2. **Database Utilities**: The `testutil` package provides functions to work with the test database:
   - `InitTestDB()`: Initializes the test database connection
   - `CleanupTestDB()`: Closes the test database connection
   - `WithTestTransaction()`: Runs a test within a transaction that is rolled back after completion

## Writing Tests with Database Access

### Basic Tests

For simple tests that need the database but don't modify data:

```go
func TestYourFunction(t *testing.T) {
    // Skip if testDB is nil (will be handled by TestMain)
    if testDB == nil {
        t.Skip("Test database not initialized")
    }
    
    // Use testDB in your test
    // ...
}
```

### Tests with Data Modifications

For tests that need to modify data, use transactions to ensure test isolation:

```go
func TestDataModification(t *testing.T) {
    testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
        // Create test data
        // Execute operations
        // Verify results
        
        // The transaction will be automatically rolled back after the test
    })
}
```

### Subtests with Transactions

For multiple related tests that should be isolated from each other:

```go
func TestMultipleOperations(t *testing.T) {
    t.Run("First operation", func(t *testing.T) {
        testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
            // First test logic
        })
    })
    
    t.Run("Second operation", func(t *testing.T) {
        testutil.WithTestTransaction(t, func(ctx context.Context, q *db.Queries) {
            // Second test logic - will not see data from first test
        })
    })
}
```

## Running Tests

To run all tests:

```bash
go test ./tests/...
```

To run a specific test:

```bash
go test ./tests -run TestName
```

To run tests with verbose output:

```bash
go test ./tests/... -v
```

## Code Coverage

Code coverage helps identify parts of your codebase that aren't being tested adequately.

### Running Coverage Analysis

You can use the provided Makefile targets to run coverage analysis:

```bash
# Run comprehensive coverage for all packages
make coverage

# Run coverage specifically for handlers
make coverage-handlers
```

Or manually with the Go tools:

```bash
# Generate coverage profile
go test -coverprofile=coverage.out -covermode=atomic ./...

# View coverage stats in terminal
go tool cover -func=coverage.out

# Generate HTML report
go tool cover -html=coverage.out -o coverage.html
```

### Coverage Shell Script

For convenience, you can also use the shell script:

```bash
# Make it executable
chmod +x tests/coverage.sh

# Run it
./tests/coverage.sh
```

### Understanding Coverage Reports

The coverage report shows:

- **Coverage percentage**: The percentage of code that is executed during tests
- **File-by-file breakdown**: Which files have good or poor coverage
- **Line-by-line highlighting**: In the HTML report, lines in green are covered, red are not covered

Aim to have at least 70-80% coverage for critical code paths, especially in the business logic and handlers. 