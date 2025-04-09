package utils

import (
	"crypto/rand"
	"encoding/base64"
	"time"
)

// GenerateRandomToken generates a cryptographically secure random token
// with a specified length in bytes. The resulting string is base64 encoded.
func GenerateRandomToken(length int) (string, error) {
	randomBytes := make([]byte, length)
	_, err := rand.Read(randomBytes)
	if err != nil {
		return "", err
	}

	// Use URL-safe base64 encoding to ensure token is URL-friendly
	return base64.URLEncoding.EncodeToString(randomBytes), nil
}

// GetTokenExpiryTime returns a time.Time set to n hours from now
func GetTokenExpiryTime(hoursValid int) time.Time {
	return time.Now().Add(time.Duration(hoursValid) * time.Hour)
}
