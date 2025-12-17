package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"shoprop-backend/models"
	"shoprop-backend/services"

	"cloud.google.com/go/firestore"
)

type DocumentHandler struct {
	service         *services.DocumentService
	r2Service       *services.R2StorageService
	propertyService *services.PropertyService
	userService     *services.UserService
}

func NewDocumentHandler(client *firestore.Client, r2Service *services.R2StorageService) *DocumentHandler {
	return &DocumentHandler{
		service:         services.NewDocumentService(client),
		r2Service:       r2Service,
		propertyService: services.NewPropertyService(client),
		userService:     services.NewUserService(client),
	}
}

// CreateDocument handles POST /documents
func (h *DocumentHandler) CreateDocument(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get user ID from header or query parameter
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	if userID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "User ID is required",
		})
		return
	}

	// Read the request body with a size limit (20MB for base64 encoded files)
	r.Body = http.MaxBytesReader(w, r.Body, 20<<20) // 20MB limit

	var req models.CreateDocumentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	// Validate required fields
	if req.PropertyID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "PropertyID is required",
		})
		return
	}
	if req.DocumentName == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "DocumentName is required",
		})
		return
	}
	if req.DocumentType == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "DocumentType is required",
		})
		return
	}
	if req.FileURL == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "FileURL is required",
		})
		return
	}

	// Get user details
	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "User not found: " + err.Error(),
		})
		return
	}

	// Get property details
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Property not found: " + err.Error(),
		})
		return
	}

	// Check if fileUrl is already a URL
	var storageURL string
	if strings.HasPrefix(req.FileURL, "https://") || strings.HasPrefix(req.FileURL, "http://") {
		// Already a URL, use it directly
		storageURL = req.FileURL
	} else {
		// Upload base64 file to R2
		// Sanitize document name for filename
		sanitizedName := strings.ReplaceAll(req.DocumentName, " ", "_")
		sanitizedName = strings.ReplaceAll(sanitizedName, "/", "_")
		sanitizedName = strings.ReplaceAll(sanitizedName, "\\", "_")

		folder := fmt.Sprintf("documents/%s", req.PropertyID)
		identifier := fmt.Sprintf("%s_%s", sanitizedName, req.DocumentType)

		uploadedURL, err := h.r2Service.UploadBase64Document(
			req.FileURL,
			folder,
			identifier,
		)
		if err != nil {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Failed to upload file to Storage: " + err.Error(),
			})
			return
		}
		storageURL = uploadedURL
	}

	// Update request with Storage URL instead of base64
	req.FileURL = storageURL

	// Create document with Storage URL
	document, err := h.service.CreateDocument(
		ctx,
		&req,
		userID,
		user.Name,
		property.Title,
	)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to create document: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Document uploaded successfully",
		"document": document,
	}); err != nil {
		// Error encoding response
	}
}

// GetDocumentsByProperty handles GET /documents/property/:propertyId
func (h *DocumentHandler) GetDocumentsByProperty(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract property ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/documents/property/")
	propertyID := strings.TrimSuffix(path, "/")

	if propertyID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Property ID is required",
		})
		return
	}

	documents, err := h.service.GetDocumentsByPropertyID(ctx, propertyID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   false,
			"message":   "Failed to fetch documents: " + err.Error(),
			"documents": []interface{}{},
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"documents": documents,
	}); err != nil {
	}
}

// GetDocumentByID handles GET /documents/:documentId
func (h *DocumentHandler) GetDocumentByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract document ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/documents/")
	documentID := strings.TrimSuffix(path, "/")

	if documentID == "" {
		http.Error(w, "Document ID is required", http.StatusBadRequest)
		return
	}

	document, err := h.service.GetDocumentByID(ctx, documentID)
	if err != nil {
		http.Error(w, "Document not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"document": document,
	})
}

// DeleteDocument handles DELETE /documents/:documentId
func (h *DocumentHandler) DeleteDocument(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract document ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/documents/")
	documentID := strings.TrimSuffix(path, "/")

	if documentID == "" {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Document ID is required",
		})
		return
	}

	// Get document first to get the Storage URL
	document, err := h.service.GetDocumentByID(ctx, documentID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Document not found: " + err.Error(),
		})
		return
	}

	// Delete from Storage if it's a Storage URL
	if document.FileURL != "" {
		if err := h.r2Service.DeleteFile(ctx, document.FileURL); err != nil {
			// Continue to delete from Firestore even if Storage deletion fails
			// Log error if needed
		} else {
		}
	}

	// Delete from Firestore
	err = h.service.DeleteDocument(ctx, documentID)
	if err != nil {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to delete document: " + err.Error(),
		})
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Document deleted successfully",
	})
}
