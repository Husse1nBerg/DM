package s3

import (
	"context"
	"mime/multipart"
	"strings"
)

// ESignService provides methods for handling e-signature URLs
type ESignService struct {
	s3Service *S3Service
	baseURL   string // Base URL for serving documents
}

// NewDocumentService creates a new DocumentService
func NewESignService(s3Service *S3Service, baseURL string) *ESignService {
	return &ESignService{
		s3Service: s3Service,
		baseURL:   baseURL,
	}
}

// GetFullDocumentURL takes a document path and returns a complete URL with the base URL if needed
func (d *ESignService) GetFullDocumentURL(docPath *string) *string {
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

func (d *ESignService) UploadFileToS3(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, key string) (string, error) {
	return d.s3Service.UploadFileToS3(ctx, file, fileHeader, key)
}

func (d *ESignService) DeleteFileFromS3(ctx context.Context, key string) error {
	return d.s3Service.DeleteFileFromS3(ctx, key)
}
func (d *ESignService) UpdateFile(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, key string) error {
	return d.s3Service.UpdateFile(ctx, file, fileHeader, key)
}

func (d *ESignService) DuplicateFile(ctx context.Context, key string) (string, error) {
	return d.s3Service.DuplicateFile(ctx, key)
}
