package services

import (
	"context"
	"log"
	"time"

	"shoprop-backend/models"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
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

	// Ensure all fields are set (even if empty) for property_enquiry notifications
	notification := models.AdminNotification{
		ID:                  notificationID,
		Type:                req.Type,
		Title:               req.Title,
		Message:             req.Message,
		PropertyID:          req.PropertyID,
		OwnerID:             req.OwnerID,
		OwnerName:           req.OwnerName,
		OwnerPhone:          req.OwnerPhone,
		OwnerEmail:          req.OwnerEmail,
		UserID:              req.UserID,
		UserName:            req.UserName,
		UserEmail:           req.UserEmail,
		UserPhone:           req.UserPhone,
		PropertyTitle:       req.PropertyTitle,
		PropertyAddress:     req.PropertyAddress,
		PropertyListingType: req.PropertyListingType,
		ServiceType:         req.ServiceType,
		ServiceComment:      req.ServiceComment,
		ServiceImage:        req.ServiceImage,
		Timestamp:           timestamp,
		IsRead:              req.IsRead,
		Priority:            req.Priority,
		CreatedAt:           now,
		UpdatedAt:           now,
	}

	// For property_enquiry, ensure all fields are explicitly set (even if empty)
	// This ensures they are saved to Firestore even with omitempty tags
	if req.Type == "property_enquiry" {
		// Log what we're about to save
		log.Printf("🔍 Property Enquiry - Setting fields: UserID='%s', UserName='%s', UserEmail='%s', UserPhone='%s', PropertyTitle='%s', PropertyAddress='%s', PropertyListingType='%s'",
			notification.UserID, notification.UserName, notification.UserEmail, notification.UserPhone,
			notification.PropertyTitle, notification.PropertyAddress, notification.PropertyListingType)
	}

	// Set default priority if not provided
	if notification.Priority == "" {
		notification.Priority = "medium"
	}

	// Save to Firestore
	// Set() will save all fields in the struct, even empty strings
	_, err = collection.Doc(notificationID).Set(ctx, notification)
	if err != nil {
		log.Printf("Error creating admin notification: %v", err)
		return nil, err
	}

	// Debug: Log what was saved to Firestore
	if notification.Type == "property_enquiry" {
		log.Printf("✅ Saved Property Enquiry Notification to Firestore - ID: %s, UserID: %s, UserName: %s, UserEmail: %s, UserPhone: %s, PropertyTitle: %s, PropertyAddress: %s, PropertyListingType: %s",
			notificationID, notification.UserID, notification.UserName, notification.UserEmail, notification.UserPhone, notification.PropertyTitle, notification.PropertyAddress, notification.PropertyListingType)
	} else {
		log.Printf("Admin notification created successfully with ID: %s", notificationID)
	}
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
			// Check if it's iterator.Done (no more documents) or actual error
			if err == iterator.Done {
				break
			}
			log.Printf("Error iterating notifications: %v", err)
			break
		}

		var notification models.AdminNotification
		if err := doc.DataTo(&notification); err != nil {
			log.Printf("Error parsing notification: %v", err)
			continue
		}

		// Set ID from document ID (important for proper identification)
		notification.ID = doc.Ref.ID

		// Debug: Log notification data for property_enquiry type
		if notification.Type == "property_enquiry" {
			log.Printf("🔍 Retrieved Property Enquiry Notification - ID: %s, UserID: %s, UserName: %s, UserEmail: %s, UserPhone: %s, PropertyTitle: %s, PropertyAddress: %s, PropertyListingType: %s",
				notification.ID, notification.UserID, notification.UserName, notification.UserEmail, notification.UserPhone, notification.PropertyTitle, notification.PropertyAddress, notification.PropertyListingType)
		}

		notifications = append(notifications, notification)
	}

	// Always return at least an empty slice, never nil
	if notifications == nil {
		notifications = []models.AdminNotification{}
	}

	return notifications, nil
}

// GetUnreadNotifications retrieves all unread admin notifications
func (s *AdminNotificationService) GetUnreadNotifications(ctx context.Context) ([]models.AdminNotification, error) {
	collection := s.client.Collection("admin_notifications")

	// Initialize as empty slice to avoid nil
	notifications := []models.AdminNotification{}

	// Query unread notifications (without OrderBy to avoid composite index requirement)
	// We'll sort by timestamp in memory instead
	iter := collection.Where("isRead", "==", false).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err != nil {
			// Check if it's iterator.Done (no more documents) or actual error
			if err == iterator.Done {
				log.Printf("✅ Finished iterating unread notifications, found %d", len(notifications))
				break
			}
			// Log error but return empty array instead of failing
			log.Printf("❌ Error iterating unread notifications: %v", err)
			// Return the notifications we've collected so far (might be empty, but not nil)
			return notifications, nil
		}

		var notification models.AdminNotification
		if err := doc.DataTo(&notification); err != nil {
			log.Printf("Error parsing notification: %v", err)
			continue
		}

		// Set ID from document ID (important for proper identification)
		notification.ID = doc.Ref.ID

		// Debug: Log notification data for property_enquiry type
		if notification.Type == "property_enquiry" {
			log.Printf("🔍 Retrieved Unread Property Enquiry Notification - ID: %s, UserID: %s, UserName: %s, UserEmail: %s, UserPhone: %s, PropertyTitle: %s, PropertyAddress: %s, PropertyListingType: %s",
				notification.ID, notification.UserID, notification.UserName, notification.UserEmail, notification.UserPhone, notification.PropertyTitle, notification.PropertyAddress, notification.PropertyListingType)
		}

		notifications = append(notifications, notification)
	}

	// Sort by timestamp in memory (newest first)
	// This avoids requiring a Firestore composite index
	for i := 0; i < len(notifications)-1; i++ {
		for j := i + 1; j < len(notifications); j++ {
			if notifications[i].Timestamp.Before(notifications[j].Timestamp) {
				notifications[i], notifications[j] = notifications[j], notifications[i]
			}
		}
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
