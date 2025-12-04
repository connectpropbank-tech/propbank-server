package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"shoprop-backend/models"
	"shoprop-backend/services"

	"cloud.google.com/go/firestore"
)

type ReviewHandler struct {
	service                  *services.ReviewService
	propertyService          *services.PropertyService
	userService              *services.UserService
	adminNotificationService *services.AdminNotificationService
}

func NewReviewHandler(client *firestore.Client) *ReviewHandler {
	return &ReviewHandler{
		service:                  services.NewReviewService(client),
		propertyService:          services.NewPropertyService(client),
		userService:              services.NewUserService(client),
		adminNotificationService: services.NewAdminNotificationService(client),
	}
}

// CreateReview handles POST /reviews
func (h *ReviewHandler) CreateReview(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get user ID from header or query parameter
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusUnauthorized)
		return
	}

	var req models.CreateReviewRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate reviewer type
	if req.ReviewerType != "tenant" && req.ReviewerType != "owner" {
		http.Error(w, "Invalid reviewer type. Must be 'tenant' or 'owner'", http.StatusBadRequest)
		return
	}

	// Get reviewer (user) details
	reviewer, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get property details
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Get owner details (optional - continue even if not found)
	_, err = h.userService.GetUserByID(ctx, property.OwnerUID)
	if err != nil {
		// Continue even if owner not found
	}

	// Get tenant details if property has tenants
	var tenantID, tenantName string
	if len(property.Tenants) > 0 {
		tenantID = property.Tenants[0].ID
		tenantName = property.Tenants[0].FirstName + " " + property.Tenants[0].LastName
	}

	// Create review
	review, err := h.service.CreateReview(
		ctx,
		&req,
		userID,
		reviewer.Name,
		reviewer.Email,
		reviewer.PhoneNumber,
		property.Title,
		property.OwnerUID,
		property.OwnerName,
		property.OwnerEmail,
		property.PrimaryNo,
		tenantID,
		tenantName,
	)
	if err != nil {
		http.Error(w, "Failed to create review", http.StatusInternalServerError)
		return
	}

	// Create admin notification
	notificationReq := models.CreateAdminNotificationRequest{
		Type:  "review",
		Title: "New Review Submitted",
		Message: strings.Join([]string{
			reviewer.Name + " (" + req.ReviewerType + ") submitted a review for property: " + property.Title,
			"Property ID: " + req.PropertyID,
			"Reviewer: " + reviewer.Name + " (" + reviewer.Email + ")",
		}, "\n"),
		PropertyID:          req.PropertyID,
		OwnerID:             property.OwnerUID,
		OwnerName:           property.OwnerName,
		OwnerPhone:          property.PrimaryNo,
		OwnerEmail:          property.OwnerEmail,
		UserID:              userID,
		UserName:            reviewer.Name,
		UserEmail:           reviewer.Email,
		UserPhone:           reviewer.PhoneNumber,
		PropertyTitle:       property.Title,
		PropertyAddress:     property.Address,
		PropertyListingType: property.ListingType,
		Timestamp:           time.Now().Format(time.RFC3339),
		IsRead:              false,
		Priority:            "medium",
	}

	_, err = h.adminNotificationService.CreateNotification(ctx, notificationReq)
	if err != nil {
		// Don't fail the request if notification fails
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Review submitted successfully and admin has been notified",
		"review":  review,
	})
}

// GetReviewByID handles GET /reviews/:reviewId
func (h *ReviewHandler) GetReviewByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract review ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/reviews/")
	reviewID := strings.TrimSuffix(path, "/")

	if reviewID == "" {
		http.Error(w, "Review ID is required", http.StatusBadRequest)
		return
	}

	review, err := h.service.GetReviewByID(ctx, reviewID)
	if err != nil {
		http.Error(w, "Review not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"review":  review,
	})
}

// GetReviewsByProperty handles GET /reviews/property/:propertyId
func (h *ReviewHandler) GetReviewsByProperty(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract property ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/reviews/property/")
	propertyID := strings.TrimSuffix(path, "/")

	if propertyID == "" {
		http.Error(w, "Property ID is required", http.StatusBadRequest)
		return
	}

	reviews, err := h.service.GetReviewsByPropertyID(ctx, propertyID)
	if err != nil {
		http.Error(w, "Failed to fetch reviews", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"reviews": reviews,
	})
}
