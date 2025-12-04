package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"shoprop-backend/services"
	"strings"
)

// UploadHandler handles file upload requests
type UploadHandler struct {
	r2Service *services.R2StorageService
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler() (*UploadHandler, error) {
	r2Service, err := services.NewR2StorageService()
	if err != nil {
		return nil, fmt.Errorf("failed to initialize R2 service: %v", err)
	}

	return &UploadHandler{
		r2Service: r2Service,
	}, nil
}

// GetR2Service returns the R2 storage service instance
func (h *UploadHandler) GetR2Service() *services.R2StorageService {
	return h.r2Service
}

// UploadImageResponse represents the response for image upload
type UploadImageResponse struct {
	Success bool     `json:"success"`
	Message string   `json:"message"`
	URL     string   `json:"url,omitempty"`
	URLs    []string `json:"urls,omitempty"`
}

// UploadImage handles single image upload
func (h *UploadHandler) UploadImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 10MB)
	err := r.ParseMultipartForm(10 << 20)
	if err != nil {
		sendUploadResponse(w, false, "Failed to parse form: "+err.Error(), "", nil)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("image")
	if err != nil {
		sendUploadResponse(w, false, "Failed to get image file: "+err.Error(), "", nil)
		return
	}
	defer file.Close()

	// Get folder from form (optional)
	folder := r.FormValue("folder")
	if folder == "" {
		folder = "uploads"
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isValidImageType(contentType) {
		sendUploadResponse(w, false, "Invalid file type. Only images are allowed.", "", nil)
		return
	}

	// Upload to R2
	ctx := context.Background()
	url, err := h.r2Service.UploadFile(ctx, file, header, folder)
	if err != nil {
		log.Printf("❌ Failed to upload image: %v", err)
		sendUploadResponse(w, false, "Failed to upload image: "+err.Error(), "", nil)
		return
	}

	sendUploadResponse(w, true, "Image uploaded successfully", url, nil)
}

// UploadDocument handles document upload
func (h *UploadHandler) UploadDocument(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 50MB for documents)
	err := r.ParseMultipartForm(50 << 20)
	if err != nil {
		sendUploadResponse(w, false, "Failed to parse form: "+err.Error(), "", nil)
		return
	}

	// Get file from form
	file, header, err := r.FormFile("document")
	if err != nil {
		sendUploadResponse(w, false, "Failed to get document file: "+err.Error(), "", nil)
		return
	}
	defer file.Close()

	// Get property ID and document type from form
	propertyID := r.FormValue("propertyId")
	docType := r.FormValue("docType")

	if propertyID == "" {
		sendUploadResponse(w, false, "Property ID is required", "", nil)
		return
	}

	if docType == "" {
		docType = "general"
	}

	// Validate file type
	contentType := header.Header.Get("Content-Type")
	if !isValidDocumentType(contentType) {
		sendUploadResponse(w, false, "Invalid file type. Only PDF, images, and common document formats are allowed.", "", nil)
		return
	}

	// Upload to R2
	ctx := context.Background()
	url, err := h.r2Service.UploadDocument(ctx, file, header, propertyID, docType)
	if err != nil {
		log.Printf("❌ Failed to upload document: %v", err)
		sendUploadResponse(w, false, "Failed to upload document: "+err.Error(), "", nil)
		return
	}

	sendUploadResponse(w, true, "Document uploaded successfully", url, nil)
}

// UploadBase64Image handles base64 image upload
func (h *UploadHandler) UploadBase64Image(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Image      string `json:"image"`
		Folder     string `json:"folder"`
		Identifier string `json:"identifier"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		sendUploadResponse(w, false, "Invalid request body: "+err.Error(), "", nil)
		return
	}

	if reqBody.Image == "" {
		sendUploadResponse(w, false, "Image data is required", "", nil)
		return
	}

	if reqBody.Folder == "" {
		reqBody.Folder = "uploads"
	}

	if reqBody.Identifier == "" {
		reqBody.Identifier = "image"
	}

	// Upload to R2
	ctx := context.Background()
	url, err := h.r2Service.UploadBase64Image(ctx, reqBody.Image, reqBody.Folder, reqBody.Identifier)
	if err != nil {
		log.Printf("❌ Failed to upload base64 image: %v", err)
		sendUploadResponse(w, false, "Failed to upload image: "+err.Error(), "", nil)
		return
	}

	sendUploadResponse(w, true, "Image uploaded successfully", url, nil)
}

// UploadPropertyImages handles multiple property images upload
func (h *UploadHandler) UploadPropertyImages(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Images     []string `json:"images"`
		PropertyID string   `json:"propertyId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		sendUploadResponse(w, false, "Invalid request body: "+err.Error(), "", nil)
		return
	}

	if reqBody.PropertyID == "" {
		sendUploadResponse(w, false, "Property ID is required", "", nil)
		return
	}

	if len(reqBody.Images) == 0 {
		sendUploadResponse(w, false, "At least one image is required", "", nil)
		return
	}

	// Upload to R2
	ctx := context.Background()
	urls, err := h.r2Service.UploadPropertyImages(ctx, reqBody.Images, reqBody.PropertyID)
	if err != nil {
		log.Printf("❌ Failed to upload property images: %v", err)
		sendUploadResponse(w, false, "Failed to upload images: "+err.Error(), "", nil)
		return
	}

	sendUploadResponse(w, true, fmt.Sprintf("Successfully uploaded %d images", len(urls)), "", urls)
}

// UploadServiceRequestImage handles service request image upload
func (h *UploadHandler) UploadServiceRequestImage(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		Image     string `json:"image"`
		RequestID string `json:"requestId"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		sendUploadResponse(w, false, "Invalid request body: "+err.Error(), "", nil)
		return
	}

	if reqBody.Image == "" {
		sendUploadResponse(w, false, "Image data is required", "", nil)
		return
	}

	if reqBody.RequestID == "" {
		reqBody.RequestID = fmt.Sprintf("request_%d", getCurrentTimestamp())
	}

	// Upload to R2
	ctx := context.Background()
	url, err := h.r2Service.UploadServiceRequestImage(ctx, reqBody.Image, reqBody.RequestID)
	if err != nil {
		log.Printf("❌ Failed to upload service request image: %v", err)
		sendUploadResponse(w, false, "Failed to upload image: "+err.Error(), "", nil)
		return
	}

	sendUploadResponse(w, true, "Image uploaded successfully", url, nil)
}

// DeleteFile handles file deletion
func (h *UploadHandler) DeleteFile(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodDelete {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		sendUploadResponse(w, false, "Invalid request body: "+err.Error(), "", nil)
		return
	}

	if reqBody.URL == "" {
		sendUploadResponse(w, false, "File URL is required", "", nil)
		return
	}

	// Delete from R2
	ctx := context.Background()
	err := h.r2Service.DeleteFile(ctx, reqBody.URL)
	if err != nil {
		log.Printf("❌ Failed to delete file: %v", err)
		sendUploadResponse(w, false, "Failed to delete file: "+err.Error(), "", nil)
		return
	}

	sendUploadResponse(w, true, "File deleted successfully", "", nil)
}

// Helper functions

func sendUploadResponse(w http.ResponseWriter, success bool, message string, url string, urls []string) {
	response := UploadImageResponse{
		Success: success,
		Message: message,
		URL:     url,
		URLs:    urls,
	}

	w.Header().Set("Content-Type", "application/json")
	if !success {
		w.WriteHeader(http.StatusBadRequest)
	}
	json.NewEncoder(w).Encode(response)
}

func isValidImageType(contentType string) bool {
	validTypes := []string{
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"image/svg+xml",
	}

	for _, validType := range validTypes {
		if strings.HasPrefix(contentType, validType) {
			return true
		}
	}
	return false
}

func isValidDocumentType(contentType string) bool {
	validTypes := []string{
		"application/pdf",
		"image/jpeg",
		"image/png",
		"image/gif",
		"image/webp",
		"application/msword",
		"application/vnd.openxmlformats-officedocument.wordprocessingml.document",
		"application/vnd.ms-excel",
		"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
		"text/plain",
	}

	for _, validType := range validTypes {
		if strings.HasPrefix(contentType, validType) {
			return true
		}
	}
	return false
}

func getCurrentTimestamp() int64 {
	return int64(1000000000) // Use time.Now().Unix() in production
}
