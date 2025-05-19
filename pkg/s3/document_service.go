package s3

import (
	"strings"
)

// DocumentService provides methods for handling document URLs
type DocumentService struct {
	s3Service *S3Service
	baseURL   string // Base URL for serving documents
}

// NewDocumentService creates a new DocumentService
func NewDocumentService(s3Service *S3Service, baseURL string) *DocumentService {
	return &DocumentService{
		s3Service: s3Service,
		baseURL:   baseURL,
	}
}

// GetFullDocumentURL takes a document path and returns a complete URL with the base URL if needed
func (d *DocumentService) GetFullDocumentURL(docPath *string) *string {
	if docPath == nil || *docPath == "" {
		return nil
	}

	// If the document path is already a full URL, return it as is
	if strings.HasPrefix(*docPath, "http://") || strings.HasPrefix(*docPath, "https://") {
		return docPath
	}

	// Concatenate the base URL with the document path
	fullURL := d.baseURL
	if !strings.HasSuffix(fullURL, "/") && !strings.HasPrefix(*docPath, "/") {
		fullURL += "/"
	}
	fullURL += *docPath

	return &fullURL
}
