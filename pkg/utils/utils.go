package utils

import (
	"fmt"
	"strings"
	"time"

	"github.com/dockworks/dm-web-backend/pkg/s3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
)

// StorageUsageGB is a custom type to handle GB values with fixed decimal places
type StorageUsageGB float64

// MarshalJSON implements json.Marshaler interface
func (s *StorageUsageGB) MarshalJSON() ([]byte, error) {
	if s == nil {
		return []byte("null"), nil
	}
	// Always format with 2 decimal places
	return []byte(fmt.Sprintf("%.2f", *s)), nil
}

var (
	imageService    *s3.ImageService
	documentService *s3.DocumentService
	esignService    *s3.ESignService
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

// SetESignService sets the e-signature service for use in response formatting
func SetESignService(service *s3.ESignService) {
	esignService = service
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

// GetFullESignURL converts an e-signature path to a full URL

func GetFullESignURL(esignPath *string) *string {
	if esignService == nil || esignPath == nil || *esignPath == "" {
		return esignPath
	}

	return esignService.GetFullDocumentURL(esignPath)
}

func GenerateUsername(firstName string) string {
	// Generate a random string of 6 characters
	randomString := uuid.New().String()[:6]
	return strings.ToLower(firstName + randomString)
}

func LowerCase(s string) string {
	return strings.ToLower(s)
}

// IntToInt32Ptr converts *int to *int32
func IntToInt32Ptr(i *int) *int32 {
	if i == nil {
		return nil
	}
	v := int32(*i)
	return &v
}

// Int64OrZero returns the value of a *int64 or 0 if nil
func Int64OrZero(ptr *int64) int64 {
	if ptr == nil {
		return 0
	}
	return *ptr
}

// Int16OrZero returns the value of a *int16 or 0 if nil
func Int16OrZero(ptr *int16) int16 {
	if ptr == nil {
		return 0
	}
	return *ptr
}
