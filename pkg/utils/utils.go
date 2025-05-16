package utils

import (
	"strings"
	"time"

	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var imageService *s3.ImageService

// Now returns the current time as a pgtype.Timestamp UTC.
func PgTimeNow() pgtype.Timestamp {
	return pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
}

func PgTimeNowLocal() pgtype.Timestamp {
	return pgtype.Timestamp{Time: time.Now().Local(), Valid: true}
}

func Pointer[T any](d T) *T {
	return &d
}

func PgTimeToTimePtr(pgTime pgtype.Timestamp) *time.Time {
	if !pgTime.Valid {
		return nil
	}
	t := pgTime.Time.UTC()
	return &t
}

// SetImageService sets the image service for use in response formatting
func SetImageService(service *s3.ImageService) {
	imageService = service
}

// GetFullImageURL converts an image path to a full URL using the image service
func GetFullImageURL(imagePath *string) *string {
	if imageService == nil || imagePath == nil || *imagePath == "" {
		return imagePath
	}

	return imageService.GetFullImageURL(imagePath)
}

func GenerateUsername(firstName string) string {
	// Generate a random string of 6 characters
	randomString := uuid.New().String()[:6]
	return strings.ToLower(firstName + randomString)
}
