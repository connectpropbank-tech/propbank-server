package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

type ImageService struct {
	client     *storage.Client
	bucketName string
}

func NewImageService(client *storage.Client, bucketName string) *ImageService {
	return &ImageService{
		client:     client,
		bucketName: bucketName,
	}
}

// UploadPropertyImages uploads base64 images to Firebase Storage and returns download URLs
func (s *ImageService) UploadPropertyImages(ctx context.Context, images []string, propertyID string) ([]string, error) {
	if len(images) == 0 {
		return []string{}, nil
	}

	var imageURLs []string

	for i, base64Image := range images {
		// Skip empty images
		if base64Image == "" {
			continue
		}

		// Parse base64 data
		imageData, contentType, err := s.parseBase64Image(base64Image)
		if err != nil {
			fmt.Printf("❌ Failed to parse image %d: %v\n", i+1, err)
			continue
		}

		// Generate unique filename
		filename := s.generateFilename(propertyID, i+1, contentType)

		// Upload to Firebase Storage
		imageURL, err := s.uploadToStorage(ctx, imageData, filename, contentType)
		if err != nil {
			fmt.Printf("❌ Failed to upload image %d: %v\n", i+1, err)
			continue
		}

		imageURLs = append(imageURLs, imageURL)
		fmt.Printf("✅ Uploaded image %d: %s\n", i+1, imageURL)
	}

	return imageURLs, nil
}

// parseBase64Image extracts image data and content type from base64 string
func (s *ImageService) parseBase64Image(base64Image string) ([]byte, string, error) {
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
	}

	// Decode base64 data
	imageData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode base64: %v", err)
	}

	return imageData, contentType, nil
}

// generateFilename creates a unique filename for the image
func (s *ImageService) generateFilename(propertyID string, imageIndex int, contentType string) string {
	// Get file extension from content type
	extension := ".jpg"
	switch contentType {
	case "image/png":
		extension = ".png"
	case "image/gif":
		extension = ".gif"
	case "image/webp":
		extension = ".webp"
	}

	// Create unique filename with timestamp
	timestamp := time.Now().Unix()
	return fmt.Sprintf("properties/%s/image_%d_%d%s", propertyID, imageIndex, timestamp, extension)
}

// uploadToStorage uploads image data to Firebase Storage
func (s *ImageService) uploadToStorage(ctx context.Context, imageData []byte, filename, contentType string) (string, error) {
	// Get bucket reference
	bucket := s.client.Bucket(s.bucketName)

	// Create object reference
	obj := bucket.Object(filename)

	// Create writer
	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType
	writer.CacheControl = "public, max-age=86400" // Cache for 1 day

	// Upload image data
	if _, err := writer.Write(imageData); err != nil {
		writer.Close()
		return "", fmt.Errorf("failed to write image data: %v", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	// Make object publicly readable
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		fmt.Printf("⚠️  Warning: Failed to make object public: %v\n", err)
	}

	// Generate public URL
	imageURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, filename)

	return imageURL, nil
}

// DeletePropertyImages deletes images from Firebase Storage
func (s *ImageService) DeletePropertyImages(ctx context.Context, imageURLs []string) error {
	bucket := s.client.Bucket(s.bucketName)

	for _, imageURL := range imageURLs {
		// Extract filename from URL
		filename := s.extractFilenameFromURL(imageURL)
		if filename == "" {
			continue
		}

		// Delete object
		obj := bucket.Object(filename)
		if err := obj.Delete(ctx); err != nil {
			fmt.Printf("⚠️  Warning: Failed to delete image %s: %v\n", imageURL, err)
		} else {
			fmt.Printf("🗑️  Deleted image: %s\n", imageURL)
		}
	}

	return nil
}

// extractFilenameFromURL extracts the filename from a Firebase Storage URL
func (s *ImageService) extractFilenameFromURL(imageURL string) string {
	// Expected format: https://storage.googleapis.com/bucket-name/properties/propertyID/image_1_timestamp.jpg
	parts := strings.Split(imageURL, "/")
	if len(parts) < 4 {
		return ""
	}

	// Join the path parts after the bucket name
	return strings.Join(parts[4:], "/")
}
