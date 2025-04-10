package utils

import (
	"errors"

	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
)

var (
	ErrPasswordReused = errors.New("password was recently used, please choose a different password")
)

// HashPassword hashes a password using bcrypt with the default cost
func HashPassword(password string) (string, error) {
	hashedBytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(hashedBytes), nil
}

// VerifyPassword checks if the provided password matches the hashed password
func VerifyPassword(hashedPassword, password string) error {
	return bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(password))
}

// UpdatePasswordFields updates password hash and resets timestamp
// Returns the new password hash and reset timestamp
func UpdatePasswordFields(password string) (string, pgtype.Timestamp, error) {
	passwordHash, err := HashPassword(password)
	if err != nil {
		return "", pgtype.Timestamp{}, err
	}

	return passwordHash, PgTimeNow(), nil
}

// IsPasswordInHistory checks if the provided password matches any of the provided historical hashes
func IsPasswordInHistory(newPassword string, historyHashes []string) (bool, error) {
	for _, hash := range historyHashes {
		err := VerifyPassword(hash, newPassword)
		if err == nil {
			// If no error, passwords match, so password is in history
			return true, nil
		}
		// If error is bcrypt.ErrMismatchedHashAndPassword, passwords don't match
		// Any other error should be returned
		if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, err
		}
	}
	return false, nil
}
