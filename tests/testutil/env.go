package testutil

import (
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

// SetupTestEnv sets up the test environment variables required for tests
// This function should be called before any tests that require database connections
func SetupTestEnv() {
	// Try to find and load .env file first
	root, err := FindProjectRoot()
	if err == nil {
		envPath := filepath.Join(root, ".env")
		if _, err := os.Stat(envPath); err == nil {
			err = godotenv.Load(envPath)
			if err == nil {
				log.Println("Test environment loaded from .env file")
			} else {
				log.Printf("Warning: Error loading .env file: %v", err)
			}
		} else {
			log.Println("No .env file found, using default test values")
		}
	}

	// Set test database environment variables if not already set
	// These are default values for testing and can be overridden by actual env vars
	if os.Getenv("TEST_DB_HOST") == "" {
		os.Setenv("TEST_DB_HOST", "localhost")
	}

	if os.Getenv("TEST_DB_PORT") == "" {
		os.Setenv("TEST_DB_PORT", "5432")
	}

	if os.Getenv("TEST_DB_NAME") == "" {
		os.Setenv("TEST_DB_NAME", "marina_test")
	}

	if os.Getenv("TEST_DB_USER") == "" {
		os.Setenv("TEST_DB_USER", "postgres")
	}

	if os.Getenv("TEST_DB_PASSWORD") == "" {
		os.Setenv("TEST_DB_PASSWORD", "postgres")
	}

	if os.Getenv("TEST_DB_SCHEMA") == "" {
		os.Setenv("TEST_DB_SCHEMA", "public")
	}

	// Print current test database configuration for debugging
	log.Printf("Test database config: %s@%s:%s/%s",
		os.Getenv("TEST_DB_USER"),
		os.Getenv("TEST_DB_HOST"),
		os.Getenv("TEST_DB_PORT"),
		os.Getenv("TEST_DB_NAME"))
}

// FindProjectRoot tries to locate the project root directory
// This is used to find the .env file
func FindProjectRoot() (string, error) {
	// Start from the current directory
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}

	// Walk up the directory tree until we find a file or directory that
	// indicates we're at the project root (like go.mod)
	for {
		// Check if go.mod exists in this directory
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir, nil
		}

		// Move up one directory
		parentDir := filepath.Dir(dir)

		// If we've reached the root of the file system and haven't found
		// what we're looking for, return an error
		if parentDir == dir {
			return "", os.ErrNotExist
		}

		dir = parentDir
	}
}
