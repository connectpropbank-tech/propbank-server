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
	service *services.AdminNotificationService
}

func NewAdminNotificationHandler(client *firestore.Client) *AdminNotificationHandler {
	return &AdminNotificationHandler{
		service: services.NewAdminNotificationService(client),
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

	notification, err := h.service.CreateNotification(ctx, req)
	if err != nil {
		log.Printf("Error creating admin notification: %v", err)
		http.Error(w, "Failed to create notification", http.StatusInternalServerError)
		return
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

	// Check if requesting only unread notifications
	unreadOnly := r.URL.Query().Get("unread") == "true"

	var notifications []models.AdminNotification
	var err error

	if unreadOnly {
		notifications, err = h.service.GetUnreadNotifications(ctx)
	} else {
		notifications, err = h.service.GetAllNotifications(ctx)
	}

	if err != nil {
		log.Printf("Error getting admin notifications: %v", err)
		http.Error(w, "Failed to get notifications", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(notifications)
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
