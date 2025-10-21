package utils

import (
	"fmt"
	"strconv"
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

// StringToInt16Ptr converts a *string to a *int16, returns error if not a valid integer
func StringToInt16Ptr(s string) (*int16, error) {
	v, err := strconv.ParseInt(s, 10, 16)
	if err != nil {
		return nil, err
	}
	res := int16(v)
	return &res, nil
}

// DetectFormFields is a placeholder function for detecting form fields in PDF pages
func DetectFormFields(pages []struct {
	ImageBase64 string
	PageNumber  int
}) []struct {
	PageNumber int
	PageSize   struct {
		WidthPx  int
		HeightPx int
	}
	Fields []struct {
		ID         string
		Type       string
		Label      string
		Required   bool
		Confidence float64
		BBox       struct {
			X int
			Y int
			W int
			H int
		}
		BBoxNorm struct {
			X float64
			Y float64
			W float64
			H float64
		}
	}
} {
	// Placeholder logic for form field detection
	// This should be replaced with actual detection logic
	var detectedFields []struct {
		PageNumber int
		PageSize   struct {
			WidthPx  int
			HeightPx int
		}
		Fields []struct {
			ID         string
			Type       string
			Label      string
			Required   bool
			Confidence float64
			BBox       struct {
				X int
				Y int
				W int
				H int
			}
			BBoxNorm struct {
				X float64
				Y float64
				W float64
				H float64
			}
		}
	}

	for _, page := range pages {
		detectedFields = append(detectedFields, struct {
			PageNumber int
			PageSize   struct {
				WidthPx  int
				HeightPx int
			}
			Fields []struct {
				ID         string
				Type       string
				Label      string
				Required   bool
				Confidence float64
				BBox       struct {
					X int
					Y int
					W int
					H int
				}
				BBoxNorm struct {
					X float64
					Y float64
					W float64
					H float64
				}
			}
		}{
			PageNumber: page.PageNumber,
			PageSize: struct {
				WidthPx  int
				HeightPx int
			}{
				WidthPx:  2550, // Example width
				HeightPx: 3300, // Example height
			},
			Fields: []struct {
				ID         string
				Type       string
				Label      string
				Required   bool
				Confidence float64
				BBox       struct {
					X int
					Y int
					W int
					H int
				}
				BBoxNorm struct {
					X float64
					Y float64
					W float64
					H float64
				}
			}{
				{
					ID:         "fld_001",
					Type:       "text",
					Label:      "Customer Name",
					Required:   true,
					Confidence: 0.97,
					BBox: struct {
						X int
						Y int
						W int
						H int
					}{
						X: 210,
						Y: 480,
						W: 820,
						H: 60,
					},
					BBoxNorm: struct {
						X float64
						Y float64
						W float64
						H float64
					}{
						X: 0.082,
						Y: 0.145,
						W: 0.322,
						H: 0.018,
					},
				},
			},
		})
	}

	return detectedFields
}

// GenerateRequestID generates a unique request ID
func GenerateRequestID() string {
	return uuid.New().String()
}

// NumericToString converts a pgtype.Numeric to a string representation
func NumericToString(n pgtype.Numeric) string {
	if !n.Valid {
		return "0.00"
	}
	// Use the Int value and Exp to calculate the decimal
	var val float64
	if n.Int != nil {
		val = float64(n.Int.Int64()) * float64(n.Exp)
	}
	return fmt.Sprintf("%.2f", val)
}

// FormatCustomerName transforms customer name from "lastname, firstname" to "firstname lastname"
// If the name doesn't contain a comma, it returns the name as-is
func FormatCustomerName(name string) string {
	if name == "" {
		return name
	}

	// Trim whitespace
	name = strings.TrimSpace(name)

	// Check if name contains a comma
	if !strings.Contains(name, ",") {
		// Name is already in correct format or is just a single name
		return name
	}

	// Split by comma
	parts := strings.SplitN(name, ",", 2)
	if len(parts) != 2 {
		return name
	}

	// Extract lastname and firstname, trim whitespace
	lastname := strings.TrimSpace(parts[0])
	firstname := strings.TrimSpace(parts[1])

	// Return in "firstname lastname" format
	return fmt.Sprintf("%s %s", firstname, lastname)
}
