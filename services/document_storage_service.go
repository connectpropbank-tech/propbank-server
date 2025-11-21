package services

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"

	"cloud.google.com/go/storage"
)

type DocumentStorageService struct {
	client     *storage.Client
	bucketName string
}

func NewDocumentStorageService(client *storage.Client, bucketName string) *DocumentStorageService {
	return &DocumentStorageService{
		client:     client,
		bucketName: bucketName,
	}
}

// UploadDocument uploads a base64 encoded document to Firebase Storage and returns the download URL
func (s *DocumentStorageService) UploadDocument(ctx context.Context, base64File string, propertyID, documentName, documentType string) (string, error) {
	if base64File == "" {
		return "", fmt.Errorf("file data is empty")
	}

	// Parse base64 data (handles both images and PDFs)
	fileData, contentType, err := s.parseBase64File(base64File)
	if err != nil {
		log.Printf("❌ Failed to parse file: %v", err)
		return "", fmt.Errorf("failed to parse file: %w", err)
	}

	// Generate unique filename
	filename := s.generateFilename(propertyID, documentName, documentType, contentType)

	// Upload to Firebase Storage
	fileURL, err := s.uploadToStorage(ctx, fileData, filename, contentType)
	if err != nil {
		log.Printf("❌ Failed to upload file: %v", err)
		return "", fmt.Errorf("failed to upload file: %w", err)
	}

	log.Printf("✅ Uploaded document to Storage: %s", fileURL)
	return fileURL, nil
}

// parseBase64File extracts file data and content type from base64 string
func (s *DocumentStorageService) parseBase64File(base64File string) ([]byte, string, error) {
	// Expected format: data:image/png;base64,/9j/4AAQSkZJRgABAQAAAQ...
	// or: data:application/pdf;base64,JVBERi0xLjQKJdP...
	parts := strings.Split(base64File, ",")
	if len(parts) != 2 {
		return nil, "", fmt.Errorf("invalid base64 format")
	}

	// Extract content type from header
	header := parts[0]
	contentType := "application/octet-stream" // default
	
	if strings.Contains(header, "image/jpeg") || strings.Contains(header, "image/jpg") {
		contentType = "image/jpeg"
	} else if strings.Contains(header, "image/png") {
		contentType = "image/png"
	} else if strings.Contains(header, "image/gif") {
		contentType = "image/gif"
	} else if strings.Contains(header, "image/webp") {
		contentType = "image/webp"
	} else if strings.Contains(header, "application/pdf") {
		contentType = "application/pdf"
	} else if strings.Contains(header, "application/msword") {
		contentType = "application/msword"
	} else if strings.Contains(header, "application/vnd.openxmlformats-officedocument.wordprocessingml.document") {
		contentType = "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	}

	// Decode base64 data
	fileData, err := base64.StdEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, "", fmt.Errorf("failed to decode base64: %v", err)
	}

	return fileData, contentType, nil
}

// generateFilename creates a unique filename for the document
func (s *DocumentStorageService) generateFilename(propertyID, documentName, documentType, contentType string) string {
	// Get file extension from content type
	extension := ".bin"
	switch contentType {
	case "image/jpeg":
		extension = ".jpg"
	case "image/png":
		extension = ".png"
	case "image/gif":
		extension = ".gif"
	case "image/webp":
		extension = ".webp"
	case "application/pdf":
		extension = ".pdf"
	case "application/msword":
		extension = ".doc"
	case "application/vnd.openxmlformats-officedocument.wordprocessingml.document":
		extension = ".docx"
	}

	// Sanitize document name for filename
	sanitizedName := strings.ReplaceAll(documentName, " ", "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, "/", "_")
	sanitizedName = strings.ReplaceAll(sanitizedName, "\\", "_")
	
	// Create unique filename with timestamp
	timestamp := time.Now().Unix()
	return fmt.Sprintf("documents/%s/%s_%s_%d%s", propertyID, sanitizedName, documentType, timestamp, extension)
}

// uploadToStorage uploads file data to Firebase Storage
func (s *DocumentStorageService) uploadToStorage(ctx context.Context, fileData []byte, filename, contentType string) (string, error) {
	// Verify bucket exists first
	bucket := s.client.Bucket(s.bucketName)
	if _, err := bucket.Attrs(ctx); err != nil {
		log.Printf("❌ Bucket '%s' does not exist or is not accessible: %v", s.bucketName, err)
		return "", fmt.Errorf("bucket '%s' does not exist. Please create it in Firebase Console or check the bucket name. Error: %v", s.bucketName, err)
	}

	// Create object reference
	obj := bucket.Object(filename)

	// Create writer
	writer := obj.NewWriter(ctx)
	writer.ContentType = contentType
	writer.CacheControl = "public, max-age=86400" // Cache for 1 day

	// Upload file data
	if _, err := writer.Write(fileData); err != nil {
		writer.Close()
		return "", fmt.Errorf("failed to write file data: %v", err)
	}

	// Close writer
	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("failed to close writer: %v", err)
	}

	// Make object publicly readable
	if err := obj.ACL().Set(ctx, storage.AllUsers, storage.RoleReader); err != nil {
		log.Printf("⚠️  Warning: Failed to make document public: %v", err)
	}

	// Generate public URL
	fileURL := fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, filename)

	return fileURL, nil
}

// DeleteDocument deletes a document from Firebase Storage
func (s *DocumentStorageService) DeleteDocument(ctx context.Context, fileURL string) error {
	// Extract filename from URL
	filename := s.extractFilenameFromURL(fileURL)
	if filename == "" {
		return fmt.Errorf("invalid file URL: %s", fileURL)
	}

	bucket := s.client.Bucket(s.bucketName)
	obj := bucket.Object(filename)
	
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("failed to delete document: %v", err)
	}

	log.Printf("🗑️  Deleted document from Storage: %s", fileURL)
	return nil
}

// extractFilenameFromURL extracts the filename from a Firebase Storage URL
func (s *DocumentStorageService) extractFilenameFromURL(fileURL string) string {
	// Expected format: https://storage.googleapis.com/bucket-name/documents/propertyID/filename.ext
	parts := strings.Split(fileURL, "/")
	if len(parts) < 4 {
		return ""
	}

	// Join the path parts after the bucket name
	return strings.Join(parts[4:], "/")
}

