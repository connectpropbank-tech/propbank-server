package services

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

// R2StorageService handles file uploads to Cloudflare R2
type R2StorageService struct {
	client     *s3.Client
	bucketName string
	publicURL  string
}

// NewR2StorageService creates a new R2 storage service
func NewR2StorageService() (*R2StorageService, error) {
	accountID := os.Getenv("CLOUDFLARE_ACCOUNT_ID")
	accessKeyID := os.Getenv("CLOUDFLARE_R2_ACCESS_KEY_ID")
	secretAccessKey := os.Getenv("CLOUDFLARE_R2_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("CLOUDFLARE_R2_BUCKET_NAME")
	publicURL := os.Getenv("CLOUDFLARE_R2_PUBLIC_URL")

	if accountID == "" || accessKeyID == "" || secretAccessKey == "" || bucketName == "" {
		return nil, fmt.Errorf("missing required Cloudflare R2 environment variables")
	}

	// R2 endpoint
	endpoint := fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID)

	// Create custom resolver for R2 endpoint
	r2Resolver := aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...interface{}) (aws.Endpoint, error) {
		return aws.Endpoint{
			URL: endpoint,
		}, nil
	})

	// Create AWS config for R2
	cfg, err := config.LoadDefaultConfig(context.TODO(),
		config.WithEndpointResolverWithOptions(r2Resolver),
		config.WithCredentialsProvider(credentials.NewStaticCredentialsProvider(accessKeyID, secretAccessKey, "")),
		config.WithRegion("auto"),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to load R2 config: %v", err)
	}

	// Create S3 client for R2
	client := s3.NewFromConfig(cfg)

	return &R2StorageService{
		client:     client,
		bucketName: bucketName,
		publicURL:  publicURL,
	}, nil
}

// UploadFile uploads a file to R2 and returns the public URL
func (s *R2StorageService) UploadFile(ctx context.Context, file multipart.File, header *multipart.FileHeader, folder string) (string, error) {
	// Read file content
	fileBytes, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("failed to read file: %v", err)
	}

	// Generate unique filename
	filename := s.generateFilename(header.Filename, folder)

	// Detect content type
	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = http.DetectContentType(fileBytes)
	}

	// Upload to R2
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(fileBytes),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %v", err)
	}

	// Generate public URL
	publicURL := s.getPublicURL(filename)

	return publicURL, nil
}

// UploadBase64Image uploads a base64 encoded image to R2
func (s *R2StorageService) UploadBase64Image(ctx context.Context, base64Image string, folder string, identifier string) (string, error) {
	// Parse base64 data
	imageData, contentType, err := s.parseBase64Image(base64Image)
	if err != nil {
		return "", fmt.Errorf("failed to parse base64 image: %v", err)
	}

	// Generate filename
	extension := s.getExtensionFromContentType(contentType)
	timestamp := time.Now().Unix()
	filename := fmt.Sprintf("%s/%s_%d%s", folder, identifier, timestamp, extension)

	// Upload to R2
	_, err = s.client.PutObject(ctx, &s3.PutObjectInput{
		Bucket:      aws.String(s.bucketName),
		Key:         aws.String(filename),
		Body:        bytes.NewReader(imageData),
		ContentType: aws.String(contentType),
	})
	if err != nil {
		return "", fmt.Errorf("failed to upload to R2: %v", err)
	}

	// Generate public URL
	publicURL := s.getPublicURL(filename)

	return publicURL, nil
}

// UploadPropertyImages uploads multiple base64 images for a property
func (s *R2StorageService) UploadPropertyImages(ctx context.Context, images []string, propertyID string) ([]string, error) {
	if len(images) == 0 {
		return []string{}, nil
	}

	var imageURLs []string
	folder := fmt.Sprintf("properties/%s", propertyID)

	for i, base64Image := range images {
		if base64Image == "" {
			continue
		}

		// Skip if it's already a URL (not base64)
		if strings.HasPrefix(base64Image, "http://") || strings.HasPrefix(base64Image, "https://") {
			imageURLs = append(imageURLs, base64Image)
			continue
		}

		identifier := fmt.Sprintf("image_%d", i+1)
		url, err := s.UploadBase64Image(ctx, base64Image, folder, identifier)
		if err != nil {
			continue
		}

		imageURLs = append(imageURLs, url)
	}

	return imageURLs, nil
}

// UploadServiceRequestImage uploads an image for a service request
func (s *R2StorageService) UploadServiceRequestImage(ctx context.Context, base64Image string, requestID string) (string, error) {
	if base64Image == "" {
		return "", nil
	}

	// Skip if it's already a URL
	if strings.HasPrefix(base64Image, "http://") || strings.HasPrefix(base64Image, "https://") {
		return base64Image, nil
	}

	folder := fmt.Sprintf("service-requests/%s", requestID)
	return s.UploadBase64Image(ctx, base64Image, folder, "image")
}

// UploadDocument uploads a document file to R2
func (s *R2StorageService) UploadDocument(ctx context.Context, file multipart.File, header *multipart.FileHeader, propertyID string, docType string) (string, error) {
	folder := fmt.Sprintf("documents/%s/%s", propertyID, docType)
	return s.UploadFile(ctx, file, header, folder)
}

// UploadBase64Document uploads a base64 encoded document (image or PDF) to R2
func (s *R2StorageService) UploadBase64Document(base64Data string, folder string, identifier string) (string, error) {
	if base64Data == "" {
		return "", fmt.Errorf("document data is empty")
	}

	// Skip if it's already a URL
	if strings.HasPrefix(base64Data, "http://") || strings.HasPrefix(base64Data, "https://") {
		return base64Data, nil
	}

	ctx := context.Background()
	return s.UploadBase64Image(ctx, base64Data, folder, identifier)
}

// DeleteFile deletes a file from R2
func (s *R2StorageService) DeleteFile(ctx context.Context, fileURL string) error {
	// Extract key from URL
	key := s.extractKeyFromURL(fileURL)
	if key == "" {
		return fmt.Errorf("invalid file URL")
	}

	_, err := s.client.DeleteObject(ctx, &s3.DeleteObjectInput{
		Bucket: aws.String(s.bucketName),
		Key:    aws.String(key),
	})
	if err != nil {
		return fmt.Errorf("failed to delete from R2: %v", err)
	}

	return nil
}

// Helper functions

func (s *R2StorageService) generateFilename(originalName string, folder string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()
	baseName := strings.TrimSuffix(originalName, ext)
	// Sanitize filename
	baseName = strings.ReplaceAll(baseName, " ", "_")
	return fmt.Sprintf("%s/%s_%d%s", folder, baseName, timestamp, ext)
}

func (s *R2StorageService) parseBase64Image(base64Image string) ([]byte, string, error) {
	// Expected format: data:image/jpeg;base64,/9j/4AAQSkZJRgABAQAAAQ...
	parts := strings.Split(base64Image, ",")
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid base64 format")
	}

	// Extract content type from header
	header := parts[0]
	contentType := "image/jpeg" // default
	if strings.Contains(header, "image/png") {
		contentType = "image/png"
	} else if strings.Contains(header, "image/gif") {
		contentType = "image/gif"
	} else if strings.Contains(header, "image/webp") {
		contentType = "image/webp"
	} else if strings.Contains(header, "application/pdf") {
		contentType = "application/pdf"
	}

	// Decode base64 data
	imageData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode base64: %v", err)
	}

	return imageData, contentType, nil
}

func (s *R2StorageService) getExtensionFromContentType(contentType string) string {
	switch contentType {
	case "image/png":
		return ".png"
	case "image/gif":
		return ".gif"
	case "image/webp":
		return ".webp"
	case "application/pdf":
		return ".pdf"
	default:
		return ".jpg"
	}
}

func (s *R2StorageService) getPublicURL(key string) string {
	if s.publicURL != "" {
		return fmt.Sprintf("%s/%s", strings.TrimSuffix(s.publicURL, "/"), key)
	}
	// Fallback to R2.dev URL if public URL not configured
	return fmt.Sprintf("https://%s.r2.dev/%s", s.bucketName, key)
}

func (s *R2StorageService) extractKeyFromURL(fileURL string) string {
	// Try to extract from public URL
	if s.publicURL != "" && strings.HasPrefix(fileURL, s.publicURL) {
		return strings.TrimPrefix(fileURL, s.publicURL+"/")
	}
	// Try to extract from r2.dev URL
	r2DevPrefix := fmt.Sprintf("https://%s.r2.dev/", s.bucketName)
	if strings.HasPrefix(fileURL, r2DevPrefix) {
		return strings.TrimPrefix(fileURL, r2DevPrefix)
	}
	return ""
}
