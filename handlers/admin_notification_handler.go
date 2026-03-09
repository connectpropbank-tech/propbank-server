package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"

	"shoprop-backend/models"
	"shoprop-backend/services"

	"cloud.google.com/go/firestore"
)

type AdminNotificationHandler struct {
	service         *services.AdminNotificationService
	propertyService *services.PropertyService
	emailService    *services.EmailService
}

func NewAdminNotificationHandler(client *firestore.Client, emailService *services.EmailService) *AdminNotificationHandler {
	return &AdminNotificationHandler{
		service:         services.NewAdminNotificationService(client),
		propertyService: services.NewPropertyService(client),
		emailService:    emailService,
	}
}

// CreateAdminNotification handles POST /admin/notifications
func (h *AdminNotificationHandler) CreateAdminNotification(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req models.CreateAdminNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// For property_enquiry notifications, fetch property details from database if PropertyID is provided
	if req.Type == "property_enquiry" && req.PropertyID != "" {
		property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
		if err != nil {
			// Property not found, continue with request data
		} else {

			// Override with actual property data from database (prioritize database values)
			if property.Title != "" {
				req.PropertyTitle = property.Title
			}
			if property.Address != "" {
				req.PropertyAddress = property.Address
			} else if property.Location != "" {
				// Fallback to Location if Address is empty
				req.PropertyAddress = property.Location
			}
			if property.ListingType != "" {
				req.PropertyListingType = property.ListingType
			}

			// Also update owner details if available
			if property.OwnerName != "" && req.OwnerName == "" {
				req.OwnerName = property.OwnerName
			}
			if property.OwnerEmail != "" && req.OwnerEmail == "" {
				req.OwnerEmail = property.OwnerEmail
			}
			if property.OwnerUID != "" && req.OwnerID == "" {
				req.OwnerID = property.OwnerUID
			}
		}
	}

	notification, err := h.service.CreateNotification(ctx, req)
	if err != nil {
		http.Error(w, "Failed to create notification", http.StatusInternalServerError)
		return
	}

	// Trigger email notification for general inquiries
	if req.Type == "general_inquiry" {
		adminEmail := "connectpropbank@gmail.com" // Default admin email
		// If ADMIN_EMAIL is set, use that as the recipient
		if adminEnv := os.Getenv("ADMIN_EMAIL"); adminEnv != "" {
			adminEmail = adminEnv
		}

		// Send in background so we don't block the API response
		go func() {
			if err := h.emailService.SendGeneralInquiryNotification(adminEmail, *notification); err != nil {
				fmt.Printf("Error sending general inquiry email: %v\n", err)
			}
		}()
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":      true,
		"message":      "Admin notification created successfully",
		"notification": notification,
	})
}

// GetAdminNotifications handles GET /admin/notifications
func (h *AdminNotificationHandler) GetAdminNotifications(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Always initialize as empty array to prevent nil
	notifications := []models.AdminNotification{}

	// Check if requesting only unread notifications
	unreadOnly := r.URL.Query().Get("unread") == "true"

	// Check if requesting notifications for a specific property
	propertyID := r.URL.Query().Get("propertyId")

	var err error

	if propertyID != "" {
		// Get notifications for specific property
		notifications, err = h.service.GetNotificationsByPropertyID(ctx, propertyID)
	} else if unreadOnly {
		notifications, err = h.service.GetUnreadNotifications(ctx)
	} else {
		notifications, err = h.service.GetAllNotifications(ctx)
	}

	// Double-check: ensure we have an array, never nil
	if notifications == nil {
		notifications = []models.AdminNotification{}
	}

	// Log error but still return array to prevent breaking frontend
	if err != nil {
	}

	// Count unread for logging
	unreadCount := 0
	for _, n := range notifications {
		if !n.IsRead {
			unreadCount++
		}
	}

	// Always return JSON array, never null
	// Ensure notifications is never nil before encoding
	if notifications == nil {
		notifications = []models.AdminNotification{}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	// Encode the array - this should never be null
	if encodeErr := json.NewEncoder(w).Encode(notifications); encodeErr != nil {
		// Fallback: write empty array directly
		w.Write([]byte("[]"))
	} else {
	}
}

// MarkNotificationAsRead handles PUT /admin/notifications/{id}/read
func (h *AdminNotificationHandler) MarkNotificationAsRead(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/admin/notifications/")
	notificationID := strings.TrimSuffix(path, "/read")

	if notificationID == "" {
		http.Error(w, "Notification ID is required", http.StatusBadRequest)
		return
	}

	err := h.service.MarkAsRead(ctx, notificationID)
	if err != nil {
		http.Error(w, "Failed to mark notification as read", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Notification marked as read",
		"id":      notificationID,
	})
}

// DeleteAdminNotification handles DELETE /admin/notifications/{id}
func (h *AdminNotificationHandler) DeleteAdminNotification(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract ID from URL path
	notificationID := strings.TrimPrefix(r.URL.Path, "/admin/notifications/")

	if notificationID == "" {
		http.Error(w, "Notification ID is required", http.StatusBadRequest)
		return
	}

	err := h.service.DeleteNotification(ctx, notificationID)
	if err != nil {
		http.Error(w, "Failed to delete notification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Notification deleted successfully",
		"id":      notificationID,
	})
}
