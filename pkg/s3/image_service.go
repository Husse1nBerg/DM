package s3

import (
	"context"
	"errors"
	"mime/multipart"
	"strings"
)

// ImageType represents different entity types for image storage
type ImageType string

const (
	UserImageType         ImageType = "users"
	MarinaImageType       ImageType = "marinas"
	OrganizationImageType ImageType = "organizations"
)

// ImageService provides methods for handling image uploads
type ImageService struct {
	s3Service *S3Service
	baseURL   string // Base URL for serving images
}

// NewImageService creates a new ImageService
func NewImageService(s3Service *S3Service, baseURL string) *ImageService {
	return &ImageService{
		s3Service: s3Service,
		baseURL:   baseURL,
	}
}

// UploadImage uploads an image to S3 and returns the path
func (i *ImageService) UploadImage(ctx context.Context, file multipart.File, header *multipart.FileHeader, imageType ImageType) (string, error) {
	// Validate file type
	if !isValidImageType(header.Filename) {
		return "", errors.New("invalid image type, only jpg, jpeg, png are allowed")
	}

	// Upload the file to S3
	key, err := i.s3Service.UploadFileToS3(ctx, file, header, string(imageType))
	if err != nil {
		return "", err
	}

	return key, nil
}

// DeleteImage deletes an image from S3
func (i *ImageService) DeleteImage(ctx context.Context, key string) error {
	if key == "" {
		return nil
	}

	// Don't try to delete URLs
	if strings.HasPrefix(key, "http://") || strings.HasPrefix(key, "https://") {
		return nil
	}

	return i.s3Service.DeleteFileFromS3(ctx, key)
}

// isValidImageType checks if the file has a valid image extension
func isValidImageType(filename string) bool {
	ext := strings.ToLower(filename[strings.LastIndex(filename, ".")+1:])
	validTypes := map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
	}
	return validTypes[ext]
}
