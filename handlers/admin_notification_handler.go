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

type AdminNotificationHandler struct {
	service         *services.AdminNotificationService
	propertyService *services.PropertyService
}

func NewAdminNotificationHandler(client *firestore.Client) *AdminNotificationHandler {
	return &AdminNotificationHandler{
		service:         services.NewAdminNotificationService(client),
		propertyService: services.NewPropertyService(client),
	}
}

// CreateAdminNotification handles POST /admin/notifications
func (h *AdminNotificationHandler) CreateAdminNotification(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	var req models.CreateAdminNotificationRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Debug: Log the incoming request to verify all fields are present
	log.Printf("📥 Creating notification - Type: %s, UserID: %s, UserName: %s, UserEmail: %s, UserPhone: %s, PropertyTitle: %s, PropertyAddress: %s, PropertyListingType: %s",
		req.Type, req.UserID, req.UserName, req.UserEmail, req.UserPhone, req.PropertyTitle, req.PropertyAddress, req.PropertyListingType)

	// For property_enquiry notifications, fetch property details from database if PropertyID is provided
	if req.Type == "property_enquiry" && req.PropertyID != "" {
		log.Printf("🔍 Fetching property details for PropertyID: %s", req.PropertyID)
		property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to fetch property details: %v. Using provided values.", err)
		} else {
			log.Printf("✅ Fetched property: Title=%s, Address=%s, ListingType=%s", property.Title, property.Address, property.ListingType)
			
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
		log.Printf("Error creating admin notification: %v", err)
		http.Error(w, "Failed to create notification", http.StatusInternalServerError)
		return
	}

	// Debug: Log the created notification to verify all fields were saved
	log.Printf("✅ Notification created - ID: %s, UserID: %s, UserName: %s, UserEmail: %s, UserPhone: %s, PropertyTitle: %s, PropertyAddress: %s, PropertyListingType: %s",
		notification.ID, notification.UserID, notification.UserName, notification.UserEmail, notification.UserPhone, notification.PropertyTitle, notification.PropertyAddress, notification.PropertyListingType)

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

	var err error

	if unreadOnly {
		notifications, err = h.service.GetUnreadNotifications(ctx)
		log.Printf("🔍 GetUnreadNotifications returned %d notifications, error: %v", len(notifications), err)
	} else {
		notifications, err = h.service.GetAllNotifications(ctx)
		log.Printf("🔍 GetAllNotifications returned %d notifications, error: %v", len(notifications), err)
	}

	// Double-check: ensure we have an array, never nil
	if notifications == nil {
		log.Printf("⚠️  Notifications is nil, initializing as empty array")
		notifications = []models.AdminNotification{}
	}

	// Log error but still return array to prevent breaking frontend
	if err != nil {
		log.Printf("⚠️  Warning: Error getting admin notifications (returning %d notifications): %v", len(notifications), err)
	}

	// Count unread for logging
	unreadCount := 0
	for _, n := range notifications {
		if !n.IsRead {
			unreadCount++
		}
	}
	log.Printf("📤 Returning %d notifications to client (isRead=false: %d)", len(notifications), unreadCount)

	// Always return JSON array, never null
	// Ensure notifications is never nil before encoding
	if notifications == nil {
		notifications = []models.AdminNotification{}
		log.Printf("⚠️  Final check: notifications was nil, set to empty array")
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	// Encode the array - this should never be null
	if encodeErr := json.NewEncoder(w).Encode(notifications); encodeErr != nil {
		log.Printf("❌ Error encoding notifications: %v", encodeErr)
		// Fallback: write empty array directly
		w.Write([]byte("[]"))
	} else {
		log.Printf("✅ Successfully encoded %d notifications", len(notifications))
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
		log.Printf("Error marking notification as read: %v", err)
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
		log.Printf("Error deleting notification: %v", err)
		http.Error(w, "Failed to delete notification", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Notification deleted successfully",
		"id":      notificationID,
	})
}
