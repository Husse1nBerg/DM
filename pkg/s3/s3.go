package s3

import (
	"bytes"
	"context"
	"fmt"
	"mime/multipart"
	"path/filepath"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/google/uuid"
)

// S3Config holds the configuration for S3 storage
type S3Config struct {
	Region          string
	AccessKeyID     string
	SecretAccessKey string
	Bucket          string
}

// S3Service provides methods for interacting with S3
type S3Service struct {
	client *s3.Client
	bucket string
}

// NewS3Service creates a new S3Service instance
func NewS3Service(cfg S3Config) (*S3Service, error) {
	var optFns []func(*config.LoadOptions) error

	// Set AWS credentials
	optFns = append(optFns, config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(
		cfg.AccessKeyID,
		cfg.SecretAccessKey,
		"",
	)))

	// Set region
	optFns = append(optFns, config.WithRegion(cfg.Region))

	// Load the AWS SDK configuration
	awsCfg, err := config.LoadDefaultConfig(context.TODO(), optFns...)
	if err != nil {
		return nil, fmt.Errorf("failed to load AWS config: %w", err)
	}

	// Create a new S3 client
	client := s3.NewFromConfig(awsCfg)

	return &S3Service{
		client: client,
		bucket: cfg.Bucket,
	}, nil
}

// UploadFileToS3 uploads a file to S3 and returns the URL
func (s *S3Service) UploadFileToS3(ctx context.Context, file multipart.File, fileHeader *multipart.FileHeader, entityType string) (string, error) {
	// Generate a unique file name to avoid collisions
	ext := filepath.Ext(fileHeader.Filename)
	uuid := uuid.New().String()
	key := fmt.Sprintf("%s/%s%s", entityType, uuid, ext)

	// Read file content
	fileBytes := make([]byte, fileHeader.Size)
	_, err := file.Read(fileBytes)
	if err != nil {
		return "", fmt.Errorf("failed to read file content: %w", err)
	}

	// Reset file pointer to beginning
	file.Seek(0, 0)

	// Create the upload input parameters
	uploadInput := &s3.PutObjectInput{
		Bucket:      aws.String(s.bucket),
		Key:         aws.String(key),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(fileHeader.Header.Get("Content-Type")),
	}

	// Upload the file
	_, err = s.client.PutObject(ctx, uploadInput)
	if err != nil {
		return "", fmt.Errorf("failed to upload file to S3: %w", err)
	}

	// Return the path to the file
	return key, nil
}

// DeleteFileFromS3 deletes a file from S3
func (s *S3Service) DeleteFileFromS3(ctx context.Context, key string) error {
	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucket),
		Key:    aws.String(key),
	})

	if err != nil {
		return fmt.Errorf("failed to delete file from S3: %w", err)
	}

	return nil
}
