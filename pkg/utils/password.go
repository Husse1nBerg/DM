package utils

import (
	"github.com/jackc/pgx/v5/pgtype"
	"golang.org/x/crypto/bcrypt"
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
