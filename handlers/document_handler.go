package handlers

import (
	"context"
	"encoding/json"
	"log"
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

	log.Printf("📄 CreateDocument endpoint called: Method=%s, Path=%s", r.Method, r.URL.Path)

	// Get user ID from header or query parameter
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	if userID == "" {
		log.Printf("❌ User ID missing in request")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "User ID is required",
		})
		return
	}

	log.Printf("📄 Processing document upload for user: %s", userID)
	contentLength := r.Header.Get("Content-Length")
	log.Printf("📄 Request Content-Length: %s", contentLength)

	// Read the request body with a size limit (20MB for base64 encoded files)
	r.Body = http.MaxBytesReader(w, r.Body, 20<<20) // 20MB limit

	var req models.CreateDocumentRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&req); err != nil {
		log.Printf("❌ Error decoding request: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Invalid request body: " + err.Error(),
		})
		return
	}

	log.Printf("📄 Document request decoded: PropertyID=%s, Name=%s, Type=%s, FileURL length=%d",
		req.PropertyID, req.DocumentName, req.DocumentType, len(req.FileURL))

	// Validate required fields
	if req.PropertyID == "" {
		log.Printf("❌ PropertyID is required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "PropertyID is required",
		})
		return
	}
	if req.DocumentName == "" {
		log.Printf("❌ DocumentName is required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "DocumentName is required",
		})
		return
	}
	if req.DocumentType == "" {
		log.Printf("❌ DocumentType is required")
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "DocumentType is required",
		})
		return
	}
	if req.FileURL == "" {
		log.Printf("❌ FileURL is required")
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
		log.Printf("❌ Error fetching user: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "User not found: " + err.Error(),
		})
		return
	}

	log.Printf("📄 User found: %s (%s)", user.Name, user.Email)

	// Get property details
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
		log.Printf("❌ Error fetching property: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Property not found: " + err.Error(),
		})
		return
	}

	log.Printf("📄 Property found: %s", property.Title)

	// Check if fileUrl is already a URL or base64
	var storageURL string
	if strings.HasPrefix(req.FileURL, "https://") {
		// Already a URL, use it directly
		storageURL = req.FileURL
		log.Printf("📄 File is already a URL, using directly")
	} else {
		// Upload base64 file to Cloudflare R2
		log.Printf("📄 Uploading file to Cloudflare R2...")
		uploadedURL, err := h.r2Service.UploadBase64Document(
			req.FileURL,
			"documents",
			req.PropertyID+"-"+req.DocumentName,
		)
		if err != nil {
			log.Printf("❌ Error uploading file to R2: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Failed to upload file to Storage: " + err.Error(),
			})
			return
		}
		storageURL = uploadedURL
		log.Printf("✅ File uploaded to R2: %s", storageURL)
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
		log.Printf("❌ Error creating document in Firestore: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to create document: " + err.Error(),
		})
		return
	}

	log.Printf("✅ Document created successfully: ID=%s, Name=%s, PropertyID=%s, StorageURL=%s",
		document.ID, document.DocumentName, document.PropertyID, storageURL)

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Document uploaded successfully",
		"document": document,
	}); err != nil {
		log.Printf("Error encoding response: %v", err)
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

	log.Printf("📄 Fetching documents for property: %s", propertyID)
	documents, err := h.service.GetDocumentsByPropertyID(ctx, propertyID)
	if err != nil {
		log.Printf("Error fetching documents: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success":   false,
			"message":   "Failed to fetch documents: " + err.Error(),
			"documents": []interface{}{},
		})
		return
	}

	log.Printf("✅ Found %d documents for property %s", len(documents), propertyID)
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"documents": documents,
	}); err != nil {
		log.Printf("Error encoding documents response: %v", err)
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
		log.Printf("Error fetching document: %v", err)
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
		log.Printf("❌ Error fetching document for deletion: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Document not found: " + err.Error(),
		})
		return
	}

	// Delete from Storage if it's a URL (R2 or Firebase Storage)
	if strings.HasPrefix(document.FileURL, "https://") {
		log.Printf("🗑️  Deleting document from storage: %s", document.FileURL)
		if err := h.r2Service.DeleteFile(context.Background(), document.FileURL); err != nil {
			log.Printf("⚠️  Warning: Failed to delete document from storage: %v", err)
			// Continue to delete from Firestore even if storage deletion fails
		} else {
			log.Printf("✅ Document deleted from storage successfully")
		}
	}

	// Delete from Firestore
	err = h.service.DeleteDocument(ctx, documentID)
	if err != nil {
		log.Printf("❌ Error deleting document from Firestore: %v", err)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(map[string]interface{}{
			"success": false,
			"message": "Failed to delete document: " + err.Error(),
		})
		return
	}

	log.Printf("✅ Document deleted successfully: ID=%s", documentID)
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Document deleted successfully",
	})
}
