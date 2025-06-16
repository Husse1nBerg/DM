package config

import (
	"crypto/rand"
	"encoding/base64"
	"log"
	"os"
	"strconv"
)

type AuthConfig struct {
	AccessSecret    string
	RefreshSecret   string
	LoginAttempts   int32
	LockoutDuration int32
}

// generateSecureToken creates a random token for use as a secret key
func generateSecureToken(length int) string {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		log.Printf("Warning: Could not generate secure token: %v", err)
		return "fallback_insecure_default_secret_do_not_use_in_production"
	}
	return base64.StdEncoding.EncodeToString(b)
}

func LoadAuthConfig() AuthConfig {
	accessSecret := os.Getenv("ACCESS_SECRET")
	if accessSecret == "" {
		accessSecret = generateSecureToken(32)
		log.Println("Warning: Using generated ACCESS_SECRET. Consider setting a permanent value in your .env file")
	}

	refreshSecret := os.Getenv("REFRESH_SECRET")
	if refreshSecret == "" {
		refreshSecret = generateSecureToken(32)
		log.Println("Warning: Using generated REFRESH_SECRET. Consider setting a permanent value in your .env file")
	}

	LoginAttemptsValue := int32(6)
	if val := os.Getenv("LOGIN_ATTEMPTS"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 32); err == nil {
			LoginAttemptsValue = int32(parsed)
		}
	}

	LockoutDurationValue := int32(30)
	if val := os.Getenv("LOCKOUT_DURATION"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 32); err == nil {
			LockoutDurationValue = int32(parsed)
		}
	}

	return AuthConfig{
		AccessSecret:    accessSecret,
		RefreshSecret:   refreshSecret,
		LoginAttempts:   LoginAttemptsValue,
		LockoutDuration: LockoutDurationValue,
	}
}
