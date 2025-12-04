package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"shoprop-backend/models"
	"shoprop-backend/services"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

// ServiceHandler handles service-related HTTP requests
type ServiceHandler struct {
	serviceService           *services.ServiceService
	r2Service                *services.R2StorageService
	adminNotificationService *services.AdminNotificationService
}

// NewServiceHandler creates a new ServiceHandler
func NewServiceHandler(client *firestore.Client, r2Service *services.R2StorageService) *ServiceHandler {
	return &ServiceHandler{
		serviceService:           services.NewServiceService(client),
		r2Service:                r2Service,
		adminNotificationService: services.NewAdminNotificationService(client),
	}
}

// GetAllServices handles GET /services - fetches all active services
func (sh *ServiceHandler) GetAllServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	services, err := sh.serviceService.GetAllServices(ctx)
	if err != nil {
		log.Printf("Error fetching services: %v", err)
		response := models.ServiceResponse{
			Success: false,
			Message: "Failed to fetch services",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.ServiceResponse{
		Success: true,
		Message: "Services fetched successfully",
		Data:    services,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateServiceRequest handles POST /service-requests - creates a new service request
func (sh *ServiceHandler) CreateServiceRequest(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody struct {
		UserUID    string `json:"userUID"`
		ServiceID  string `json:"serviceId"`
		PropertyID string `json:"propertyId,omitempty"`
		Message    string `json:"message"`
		Image      string `json:"image,omitempty"` // Base64 image string
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Printf("Error decoding request body: %v", err)
		response := models.ServiceRequestResponse{
			Success: false,
			Message: "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate required fields
	if reqBody.UserUID == "" || reqBody.ServiceID == "" {
		response := models.ServiceRequestResponse{
			Success: false,
			Message: "Missing required fields: userUID, serviceId",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx := context.Background()

	// Get user information
	user, err := sh.serviceService.GetUserByID(ctx, reqBody.UserUID)
	if err != nil {
		log.Printf("Error fetching user %s: %v", reqBody.UserUID, err)
		response := models.ServiceRequestResponse{
			Success: false,
			Message: "User not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get service information
	service, err := sh.serviceService.GetServiceByID(ctx, reqBody.ServiceID)
	if err != nil {
		log.Printf("Error fetching service %s: %v", reqBody.ServiceID, err)
		response := models.ServiceRequestResponse{
			Success: false,
			Message: "Service not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Handle image upload if provided
	var imageURL string
	if reqBody.Image != "" {
		requestID := uuid.New().String()
		uploadedURL, err := sh.r2Service.UploadServiceRequestImage(ctx, reqBody.Image, requestID)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to upload service request image: %v", err)
			// Continue without image rather than failing the entire request
		} else {
			imageURL = uploadedURL
			log.Printf("✅ Service request image uploaded: %s", imageURL)
		}
	}

	// Create service request
	serviceRequest := models.ServiceRequest{
		ID:          uuid.New().String(),
		UserUID:     reqBody.UserUID,
		UserName:    user.Name,
		UserEmail:   user.Email,
		UserPhone:   user.PhoneNumber,
		ServiceID:   reqBody.ServiceID,
		ServiceName: service.Name,
		PropertyID:  reqBody.PropertyID,
		Message:     reqBody.Message,
		Image:       imageURL, // Store uploaded image URL
		Status:      models.StatusPending,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	err = sh.serviceService.CreateServiceRequest(ctx, &serviceRequest)
	if err != nil {
		log.Printf("Error creating service request: %v", err)
		response := models.ServiceRequestResponse{
			Success: false,
			Message: "Failed to create service request",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create admin notification for the service request
	sh.createServiceRequestNotification(ctx, &serviceRequest)

	response := models.ServiceRequestResponse{
		Success:        true,
		Message:        "Service request created successfully. One of our members will reach out to you soon!",
		ServiceRequest: &serviceRequest,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	log.Printf("✅ Service request created: %s by %s for service %s", serviceRequest.ID, user.Name, service.Name)
}

// GetServiceRequests handles GET /service-requests - fetches service requests (admin only or user's own)
func (sh *ServiceHandler) GetServiceRequests(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	userUID := r.URL.Query().Get("userUID")
	isAdmin := r.URL.Query().Get("admin") == "true"

	ctx := context.Background()
	var serviceRequests []models.ServiceRequest
	var err error

	if isAdmin {
		// Admin can see all requests
		serviceRequests, err = sh.serviceService.GetAllServiceRequests(ctx)
	} else if userUID != "" {
		// User can see their own requests
		serviceRequests, err = sh.serviceService.GetServiceRequestsByUser(ctx, userUID)
	} else {
		response := models.ServiceRequestResponse{
			Success: false,
			Message: "Missing userUID parameter or admin access",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if err != nil {
		log.Printf("Error fetching service requests: %v", err)
		response := models.ServiceRequestResponse{
			Success: false,
			Message: "Failed to fetch service requests",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.ServiceRequestResponse{
		Success: true,
		Message: "Service requests fetched successfully",
		Data:    serviceRequests,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateServiceRequestStatus handles PUT /service-requests/{id} - updates service request status (admin only)
func (sh *ServiceHandler) UpdateServiceRequestStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/service-requests/")
	if path == "" || path == r.URL.Path {
		http.Error(w, "Service request ID is required", http.StatusBadRequest)
		return
	}

	var reqBody struct {
		Status     models.ServiceStatus `json:"status"`
		AdminNotes string               `json:"adminNotes,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		response := models.ServiceRequestResponse{
			Success: false,
			Message: "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx := context.Background()
	err := sh.serviceService.UpdateServiceRequestStatus(ctx, path, reqBody.Status, reqBody.AdminNotes)
	if err != nil {
		log.Printf("Error updating service request %s: %v", path, err)
		response := models.ServiceRequestResponse{
			Success: false,
			Message: "Failed to update service request",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.ServiceRequestResponse{
		Success: true,
		Message: "Service request updated successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("✅ Service request %s updated to status: %s", path, reqBody.Status)
}

// PopulateServices handles POST /admin/populate-services - adds all services to database
func (sh *ServiceHandler) PopulateServices(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Define all services based on the attachment
	services := []models.Service{
		// Legal Services
		{
			ID:          uuid.New().String(),
			Name:        "Registration Of Rent Agreement",
			Description: "Legal consultation and registration of rent agreement",
			Category:    models.CategoryLegal,
			Code:        "1A",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Sale Agreement Procedure",
			Description: "Legal consultation and sale agreement procedure",
			Category:    models.CategoryLegal,
			Code:        "1B",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Any Other Legal Work",
			Description: "Legal consultation for any other legal work",
			Category:    models.CategoryLegal,
			Code:        "1C",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "CIDCO/MNNC Work",
			Description: "Legal consultation for CIDCO/MNNC work",
			Category:    models.CategoryLegal,
			Code:        "1D",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		// Other Related Services
		{
			ID:          uuid.New().String(),
			Name:        "Keys Management",
			Description: "Professional keys management service",
			Category:    models.CategoryOther,
			Code:        "2A",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Police Verification",
			Description: "Police verification service for tenants",
			Category:    models.CategoryOther,
			Code:        "2B",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Society Procedure",
			Description: "Assistance with society procedure and documentation",
			Category:    models.CategoryOther,
			Code:        "2C",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "All Maintenance Work To Shift (exp. Borne By Owner)",
			Description: "Complete maintenance work for property shifting, expenses borne by owner",
			Category:    models.CategoryOther,
			Code:        "2D",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Pre-Rental Service - Painting, Plumbing, Carpenter, Electrical",
			Description: "Complete pre-rental services including painting, plumbing, carpentry, and electrical work",
			Category:    models.CategoryOther,
			Code:        "2E",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Flat Verification Before Handover (In Case of Purchase) - Cost Rs 25000/-",
			Description: "Professional flat verification service before handover in case of purchase, cost Rs 25000/-",
			Category:    models.CategoryOther,
			Code:        "2F",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
		{
			ID:          uuid.New().String(),
			Name:        "Client Owner Meeting Arrangement with LOI",
			Description: "Arrange client owner meeting with Letter of Intent",
			Category:    models.CategoryOther,
			Code:        "2G",
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		},
	}

	ctx := context.Background()

	// Add services to Firestore
	for _, service := range services {
		err := sh.serviceService.CreateService(ctx, &service)
		if err != nil {
			log.Printf("Error creating service %s: %v", service.Name, err)
			response := models.ServiceResponse{
				Success: false,
				Message: "Failed to create services",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}
	}

	response := models.ServiceResponse{
		Success: true,
		Message: fmt.Sprintf("Successfully created %d services", len(services)),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

	log.Printf("✅ Successfully created %d services in Firestore", len(services))
}

// createServiceRequestNotification creates an admin notification for a new service request
func (sh *ServiceHandler) createServiceRequestNotification(ctx context.Context, serviceRequest *models.ServiceRequest) {
	// Create notification request
	notificationReq := models.CreateAdminNotificationRequest{
		Type:           "service_request",
		Title:          "New Service Request",
		Message:        fmt.Sprintf("Service request from %s for %s", serviceRequest.UserName, serviceRequest.ServiceName),
		PropertyID:     serviceRequest.PropertyID, // May be empty for non-property-specific services
		UserID:         serviceRequest.UserUID,
		UserName:       serviceRequest.UserName,
		UserEmail:      serviceRequest.UserEmail,
		UserPhone:      serviceRequest.UserPhone,
		ServiceType:    serviceRequest.ServiceName,
		ServiceComment: serviceRequest.Message,
		ServiceImage:   serviceRequest.Image, // Include uploaded image URL
		Timestamp:      time.Now().Format(time.RFC3339),
		IsRead:         false,
		Priority:       "medium",
	}

	// Create notification
	_, err := sh.adminNotificationService.CreateNotification(ctx, notificationReq)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to create admin notification for service request: %v", err)
	} else {
		log.Printf("✅ Admin notification created for service request %s", serviceRequest.ID)
	}
}
