package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"shoprop-backend/models"
	"shoprop-backend/services"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
)

type VisitHandler struct {
	visitService *services.VisitService
	userService  *services.UserService
	emailService *services.EmailService
}

// NewVisitHandler creates a new visit handler
func NewVisitHandler(client *firestore.Client) *VisitHandler {
	return &VisitHandler{
		visitService: services.NewVisitService(client),
		userService:  services.NewUserService(client),
		emailService: services.NewEmailService(),
	}
}

// CreateVisit handles POST /visits - creates a new visit
func (vh *VisitHandler) CreateVisit(w http.ResponseWriter, r *http.Request) {
	var req models.CreateVisitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := models.VisitResponse{
			Success: false,
			Message: "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate required fields
	if req.Title == "" || req.VisitDate == "" || req.ReminderType == "" {
		response := models.VisitResponse{
			Success: false,
			Message: "Missing required fields: title, visitDate, reminderType",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Parse visit date
	visitDate, err := time.Parse(time.RFC3339, req.VisitDate)
	if err != nil {
		response := models.VisitResponse{
			Success: false,
			Message: "Invalid date format. Use ISO 8601 format",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get user ID from request header or context (you'll need to implement auth middleware)
	userID := r.Header.Get("X-User-ID") // Temporary - should come from JWT token
	if userID == "" {
		response := models.VisitResponse{
			Success: false,
			Message: "User authentication required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Create visit model
	visit := models.Visit{
		UserID:       userID,
		Title:        req.Title,
		Description:  req.Description,
		PropertyID:   req.PropertyID,
		VisitDate:    visitDate,
		ReminderType: req.ReminderType,
		ReminderTime: req.ReminderTime,
		Status:       "scheduled",
		IsCompleted:  false,
	}

	// Save visit
	ctx := context.Background()
	if err := vh.visitService.CreateVisit(ctx, &visit); err != nil {
		response := models.VisitResponse{
			Success: false,
			Message: "Failed to create visit",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Send confirmation email
	user, err := vh.userService.GetUserByID(ctx, userID)
	if err != nil {
	} else {
		go func() {
			if emailErr := vh.emailService.SendVisitConfirmation(user, &visit); emailErr != nil {
			}
		}()
	}

	// Success response
	response := models.VisitResponse{
		Success: true,
		Message: "Visit scheduled successfully",
		Visit:   &visit,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)

}

// GetVisit handles GET /visits/{id} - gets a specific visit
func (vh *VisitHandler) GetVisit(w http.ResponseWriter, r *http.Request) {
	visitID := strings.TrimPrefix(r.URL.Path, "/visits/")
	if visitID == "" {
		http.Error(w, "Visit ID is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	visit, err := vh.visitService.GetVisitByID(ctx, visitID)
	if err != nil {
		response := models.VisitResponse{
			Success: false,
			Message: "Visit not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.VisitResponse{
		Success: true,
		Message: "Visit retrieved successfully",
		Visit:   visit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// GetVisitsByUser handles GET /visits?userId={id} - gets all visits for a user
func (vh *VisitHandler) GetVisitsByUser(w http.ResponseWriter, r *http.Request) {
	userID := r.URL.Query().Get("userId")
	if userID == "" {
		userID = r.Header.Get("X-User-ID") // Fallback to header
	}

	if userID == "" {
		response := models.VisitResponse{
			Success: false,
			Message: "User ID is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx := context.Background()
	status := r.URL.Query().Get("status") // active, completed, all

	var visits []models.Visit
	var err error

	switch status {
	case "active":
		visits, err = vh.visitService.GetActiveVisitsByUserID(ctx, userID)
	case "completed":
		visits, err = vh.visitService.GetCompletedVisitsByUserID(ctx, userID)
	default:
		visits, err = vh.visitService.GetVisitsByUserID(ctx, userID)
	}

	if err != nil {
		response := models.VisitResponse{
			Success: false,
			Message: "Failed to retrieve visits",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.VisitResponse{
		Success: true,
		Message: "Visits retrieved successfully",
		Visits:  visits,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// UpdateVisit handles PUT /visits/{id} - updates a visit
func (vh *VisitHandler) UpdateVisit(w http.ResponseWriter, r *http.Request) {
	visitID := strings.TrimPrefix(r.URL.Path, "/visits/")
	if visitID == "" {
		http.Error(w, "Visit ID is required", http.StatusBadRequest)
		return
	}

	var req models.UpdateVisitRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := models.VisitResponse{
			Success: false,
			Message: "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx := context.Background()
	if err := vh.visitService.UpdateVisit(ctx, visitID, &req); err != nil {
		response := models.VisitResponse{
			Success: false,
			Message: "Failed to update visit",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Get updated visit
	visit, err := vh.visitService.GetVisitByID(ctx, visitID)
	if err != nil {
	}

	response := models.VisitResponse{
		Success: true,
		Message: "Visit updated successfully",
		Visit:   visit,
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}

// DeleteVisit handles DELETE /visits/{id} - deletes a visit
func (vh *VisitHandler) DeleteVisit(w http.ResponseWriter, r *http.Request) {
	visitID := strings.TrimPrefix(r.URL.Path, "/visits/")
	if visitID == "" {
		http.Error(w, "Visit ID is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	if err := vh.visitService.DeleteVisit(ctx, visitID); err != nil {
		response := models.VisitResponse{
			Success: false,
			Message: "Failed to delete visit",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.VisitResponse{
		Success: true,
		Message: "Visit deleted successfully",
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)

}
