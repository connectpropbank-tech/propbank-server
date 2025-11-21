package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"shoprop-backend/models"
	"shoprop-backend/services"

	"cloud.google.com/go/firestore"
)

type AgreementHandler struct {
	propertyService          *services.PropertyService
	userService              *services.UserService
	adminNotificationService *services.AdminNotificationService
}

func NewAgreementHandler(client *firestore.Client) *AgreementHandler {
	return &AgreementHandler{
		propertyService:          services.NewPropertyService(client),
		userService:              services.NewUserService(client),
		adminNotificationService: services.NewAdminNotificationService(client),
	}
}

// RenewAgreement handles POST /agreements/renew
func (h *AgreementHandler) RenewAgreement(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusUnauthorized)
		return
	}

	var req struct {
		PropertyID string `json:"propertyId" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get property
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
		log.Printf("Error fetching property: %v", err)
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if property.OwnerUID != userID {
		http.Error(w, "Unauthorized: You can only renew your own property agreements", http.StatusForbidden)
		return
	}

	// Update property with renewed agreement status
	updateData := map[string]interface{}{
		"agreementStatus": "renewed",
		"updatedAt":       time.Now(),
	}

	updatedProperty, err := h.propertyService.UpdateProperty(ctx, req.PropertyID, updateData)
	if err != nil {
		log.Printf("Error updating property: %v", err)
		http.Error(w, "Failed to renew agreement", http.StatusInternalServerError)
		return
	}

	// Get owner details
	owner, err := h.userService.GetUserByID(ctx, property.OwnerUID)
	ownerName := property.OwnerName
	ownerEmail := property.OwnerEmail
	if err == nil && owner != nil {
		ownerName = owner.Name
		ownerEmail = owner.Email
	}

	// Create admin notification
	notificationReq := models.CreateAdminNotificationRequest{
		Type:  "agreement_renewal",
		Title: "Agreement Renewal Request",
		Message: strings.Join([]string{
			ownerName + " has requested to renew the agreement for property: " + property.Title,
			"Property ID: " + req.PropertyID,
			"Owner: " + ownerName + " (" + ownerEmail + ")",
		}, "\n"),
		PropertyID:          req.PropertyID,
		OwnerID:             property.OwnerUID,
		OwnerName:           property.OwnerName,
		OwnerPhone:          property.PrimaryNo,
		OwnerEmail:          property.OwnerEmail,
		PropertyTitle:       property.Title,
		PropertyAddress:     property.Address,
		PropertyListingType: property.ListingType,
		Timestamp:           time.Now().Format(time.RFC3339),
		IsRead:              false,
		Priority:            "high",
	}

	_, err = h.adminNotificationService.CreateNotification(ctx, notificationReq)
	if err != nil {
		log.Printf("Error creating admin notification for agreement renewal: %v", err)
		// Don't fail the request if notification fails
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Agreement renewal request submitted successfully. Admin has been notified.",
		"property": updatedProperty,
	})
}

// TerminateAgreement handles POST /agreements/terminate
func (h *AgreementHandler) TerminateAgreement(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	log.Printf("✅ TerminateAgreement handler called: Method=%s, Path=%s, URL=%s", r.Method, r.URL.Path, r.URL.String())

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusUnauthorized)
		return
	}

	var req struct {
		PropertyID string `json:"propertyId" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get property
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
		log.Printf("Error fetching property: %v", err)
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if property.OwnerUID != userID {
		http.Error(w, "Unauthorized: You can only terminate your own property agreements", http.StatusForbidden)
		return
	}

	// Update property: set agreementStatus to "terminated" and status to "inactive" (move to archive)
	updateData := map[string]interface{}{
		"agreementStatus": "terminated",
		"status":          "inactive", // Move to archive
		"updatedAt":       time.Now(),
	}

	updatedProperty, err := h.propertyService.UpdateProperty(ctx, req.PropertyID, updateData)
	if err != nil {
		log.Printf("Error updating property: %v", err)
		http.Error(w, "Failed to terminate agreement", http.StatusInternalServerError)
		return
	}

	// Get owner details
	owner, err := h.userService.GetUserByID(ctx, property.OwnerUID)
	ownerName := property.OwnerName
	ownerEmail := property.OwnerEmail
	if err == nil && owner != nil {
		ownerName = owner.Name
		ownerEmail = owner.Email
	}

	// Create admin notification
	notificationReq := models.CreateAdminNotificationRequest{
		Type:  "agreement_termination",
		Title: "Agreement Terminated",
		Message: strings.Join([]string{
			ownerName + " has terminated the agreement for property: " + property.Title,
			"Property ID: " + req.PropertyID,
			"Owner: " + ownerName + " (" + ownerEmail + ")",
			"Property has been moved to archive with status: agreement terminated",
		}, "\n"),
		PropertyID:          req.PropertyID,
		OwnerID:             property.OwnerUID,
		OwnerName:           property.OwnerName,
		OwnerPhone:          property.PrimaryNo,
		OwnerEmail:          property.OwnerEmail,
		PropertyTitle:       property.Title,
		PropertyAddress:     property.Address,
		PropertyListingType: property.ListingType,
		Timestamp:           time.Now().Format(time.RFC3339),
		IsRead:              false,
		Priority:            "high",
	}

	_, err = h.adminNotificationService.CreateNotification(ctx, notificationReq)
	if err != nil {
		log.Printf("Error creating admin notification for agreement termination: %v", err)
		// Don't fail the request if notification fails
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Agreement terminated successfully. Property moved to archive. Admin has been notified.",
		"property": updatedProperty,
	})
}
