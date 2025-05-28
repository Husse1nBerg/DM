package utils

import (
	"strings"
	"time"

	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

var (
	imageService    *s3.ImageService
	documentService *s3.DocumentService
)

// Now returns the current time as a pgtype.Timestamp UTC.
func PgTimeNow() pgtype.Timestamp {
	return pgtype.Timestamp{Time: time.Now().UTC(), Valid: true}
}

func PgTimeNowAdd(duration time.Duration) pgtype.Timestamp {
	return pgtype.Timestamp{Time: time.Now().UTC().Add(duration), Valid: true}
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

// SetDocumentService sets the document service for use in response formatting
func SetDocumentService(service *s3.DocumentService) {
	documentService = service
}

// GetFullImageURL converts an image path to a full URL using the image service
func GetFullImageURL(imagePath *string) *string {
	if imageService == nil || imagePath == nil || *imagePath == "" {
		return imagePath
	}

	return imageService.GetFullImageURL(imagePath)
}

// GetFullDocumentURL converts a document path to a full URL
func GetFullDocumentURL(docPath *string) *string {
	if documentService == nil || docPath == nil || *docPath == "" {
		return docPath
	}

	return documentService.GetFullDocumentURL(docPath)
}

func GenerateUsername(firstName string) string {
	// Generate a random string of 6 characters
	randomString := uuid.New().String()[:6]
	return strings.ToLower(firstName + randomString)
}

func LowerCase(s string) string {
	return strings.ToLower(s)
}
