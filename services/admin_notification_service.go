package services

import (
	"context"
	"log"
	"time"

	"shoprop-backend/models"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
)

// AdminNotificationService provides methods for managing admin notifications
type AdminNotificationService struct {
	client *firestore.Client
}

// NewAdminNotificationService creates a new AdminNotificationService
func NewAdminNotificationService(client *firestore.Client) *AdminNotificationService {
	return &AdminNotificationService{
		client: client,
	}
}

// CreateNotification creates a new admin notification
func (s *AdminNotificationService) CreateNotification(ctx context.Context, req models.CreateAdminNotificationRequest) (*models.AdminNotification, error) {
	collection := s.client.Collection("admin_notifications")

	// Parse timestamp
	timestamp, err := time.Parse(time.RFC3339, req.Timestamp)
	if err != nil {
		log.Printf("Error parsing timestamp: %v", err)
		timestamp = time.Now()
	}

	// Create notification with generated ID
	notificationID := uuid.New().String()
	now := time.Now()

	notification := models.AdminNotification{
		ID:         notificationID,
		Type:       req.Type,
		Title:      req.Title,
		Message:    req.Message,
		PropertyID: req.PropertyID,
		OwnerID:    req.OwnerID,
		OwnerName:  req.OwnerName,
		Timestamp:  timestamp,
		IsRead:     req.IsRead,
		Priority:   req.Priority,
		CreatedAt:  now,
		UpdatedAt:  now,
	}

	// Set default priority if not provided
	if notification.Priority == "" {
		notification.Priority = "medium"
	}

	// Save to Firestore
	_, err = collection.Doc(notificationID).Set(ctx, notification)
	if err != nil {
		log.Printf("Error creating admin notification: %v", err)
		return nil, err
	}

	log.Printf("Admin notification created successfully with ID: %s", notificationID)
	return &notification, nil
}

// GetAllNotifications retrieves all admin notifications
func (s *AdminNotificationService) GetAllNotifications(ctx context.Context) ([]models.AdminNotification, error) {
	collection := s.client.Collection("admin_notifications")

	// Query all notifications ordered by timestamp (newest first)
	iter := collection.OrderBy("timestamp", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	var notifications []models.AdminNotification
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var notification models.AdminNotification
		if err := doc.DataTo(&notification); err != nil {
			log.Printf("Error parsing notification: %v", err)
			continue
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// GetUnreadNotifications retrieves all unread admin notifications
func (s *AdminNotificationService) GetUnreadNotifications(ctx context.Context) ([]models.AdminNotification, error) {
	collection := s.client.Collection("admin_notifications")

	// Query unread notifications ordered by timestamp (newest first)
	iter := collection.Where("isRead", "==", false).OrderBy("timestamp", firestore.Desc).Documents(ctx)
	defer iter.Stop()

	var notifications []models.AdminNotification
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var notification models.AdminNotification
		if err := doc.DataTo(&notification); err != nil {
			log.Printf("Error parsing notification: %v", err)
			continue
		}

		notifications = append(notifications, notification)
	}

	return notifications, nil
}

// MarkAsRead marks a notification as read
func (s *AdminNotificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	collection := s.client.Collection("admin_notifications")

	// Update the notification
	_, err := collection.Doc(notificationID).Update(ctx, []firestore.Update{
		{
			Path:  "isRead",
			Value: true,
		},
		{
			Path:  "updatedAt",
			Value: time.Now(),
		},
	})

	if err != nil {
		log.Printf("Error marking notification as read: %v", err)
		return err
	}

	log.Printf("Notification %s marked as read", notificationID)
	return nil
}

// DeleteNotification deletes a notification
func (s *AdminNotificationService) DeleteNotification(ctx context.Context, notificationID string) error {
	collection := s.client.Collection("admin_notifications")

	_, err := collection.Doc(notificationID).Delete(ctx)
	if err != nil {
		log.Printf("Error deleting notification: %v", err)
		return err
	}

	log.Printf("Notification %s deleted successfully", notificationID)
	return nil
}
