package services

import (
	"context"
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
		OwnerRole:           req.OwnerRole,
		UserID:              req.UserID,
		UserName:            req.UserName,
		UserEmail:           req.UserEmail,
		UserPhone:           req.UserPhone,
		TenantName:          req.TenantName,
		TenantPhone:         req.TenantPhone,
		TenantEmail:         req.TenantEmail,
		Buyers:              req.Buyers,
		PropertyTitle:       req.PropertyTitle,
		PropertyAddress:     req.PropertyAddress,
		PropertyListingType: req.PropertyListingType,
		ServiceType:         req.ServiceType,
		ServiceComment:      req.ServiceComment,
		ServiceImage:        req.ServiceImage,
		// General inquiry specific fields
		InquiryType:  req.InquiryType,
		PropertyType: req.PropertyType,
		RequestVisit: req.RequestVisit,
		VisitDate:    req.VisitDate,
		VisitTime:    req.VisitTime,
		Timestamp:    timestamp,
		IsRead:       req.IsRead,
		Priority:     req.Priority,
		AdminRemarks: req.AdminRemarks,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	// For property_enquiry, ensure all fields are explicitly set (even if empty)
	// This ensures they are saved to Firestore even with omitempty tags
	if req.Type == "property_enquiry" {
		// Fields are set in notification struct
	}

	// For general_inquiry, log what we're saving
	if req.Type == "general_inquiry" {
		// Fields are set in notification struct
	}

	// Set default priority if not provided
	if notification.Priority == "" {
		notification.Priority = "medium"
	}

	// Save to Firestore
	// Set() will save all fields in the struct, even empty strings
	_, err = collection.Doc(notificationID).Set(ctx, notification)
	if err != nil {
		return nil, err
	}

	// Debug: Log what was saved to Firestore
	if notification.Type == "property_enquiry" {
		// notification saved successfully
	} else {
		// notification saved successfully
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
			break
		}

		var notification models.AdminNotification
		if err := doc.DataTo(&notification); err != nil {
			continue
		}

		// Set ID from document ID (important for proper identification)
		notification.ID = doc.Ref.ID

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
				break
			}
			// Log error but return empty array instead of failing
			// Return the notifications we've collected so far (might be empty, but not nil)
			return notifications, nil
		}

		var notification models.AdminNotification
		if err := doc.DataTo(&notification); err != nil {
			continue
		}

		// Set ID from document ID (important for proper identification)
		notification.ID = doc.Ref.ID

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

// MarkAsRead marks a notification as read and sets resolvedAt timestamp
func (s *AdminNotificationService) MarkAsRead(ctx context.Context, notificationID string) error {
	collection := s.client.Collection("admin_notifications")

	now := time.Now()
	// Update the notification
	_, err := collection.Doc(notificationID).Update(ctx, []firestore.Update{
		{
			Path:  "isRead",
			Value: true,
		},
		{
			Path:  "resolvedAt",
			Value: now,
		},
		{
			Path:  "updatedAt",
			Value: now,
		},
	})

	if err != nil {
		return err
	}

	return nil
}

// ToggleArchiveStatus archives or unarchives a notification and appends to archiveHistory
func (s *AdminNotificationService) ToggleArchiveStatus(ctx context.Context, notificationID string, archive bool) error {
	collection := s.client.Collection("admin_notifications")

	now := time.Now()

	action := "unarchived"
	isRead := false
	if archive {
		action = "archived"
		isRead = true
	}

	historyEntry := models.ArchiveHistoryEntry{
		Action:    action,
		Timestamp: now,
	}

	updates := []firestore.Update{
		{Path: "isRead", Value: isRead},
		{Path: "updatedAt", Value: now},
		{Path: "archiveHistory", Value: firestore.ArrayUnion(historyEntry)},
	}

	if archive {
		updates = append(updates, firestore.Update{Path: "resolvedAt", Value: now})
	}

	_, err := collection.Doc(notificationID).Update(ctx, updates)
	return err
}

// DeleteNotification deletes a notification
func (s *AdminNotificationService) DeleteNotification(ctx context.Context, notificationID string) error {
	collection := s.client.Collection("admin_notifications")

	_, err := collection.Doc(notificationID).Delete(ctx)
	if err != nil {
		return err
	}

	return nil
}

// GetNotificationsByPropertyID retrieves all admin notifications for a specific property
func (s *AdminNotificationService) GetNotificationsByPropertyID(ctx context.Context, propertyID string) ([]models.AdminNotification, error) {
	collection := s.client.Collection("admin_notifications")

	// Initialize as empty slice to avoid nil
	notifications := []models.AdminNotification{}

	// Query notifications by propertyId
	iter := collection.Where("propertyId", "==", propertyID).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err != nil {
			if err == iterator.Done {
				break
			}
			return notifications, nil
		}

		var notification models.AdminNotification
		if err := doc.DataTo(&notification); err != nil {
			continue
		}

		// Set ID from document ID
		notification.ID = doc.Ref.ID
		notifications = append(notifications, notification)
	}

	// Sort by timestamp in memory (newest first)
	for i := 0; i < len(notifications)-1; i++ {
		for j := i + 1; j < len(notifications); j++ {
			if notifications[i].Timestamp.Before(notifications[j].Timestamp) {
				notifications[i], notifications[j] = notifications[j], notifications[i]
			}
		}
	}

	return notifications, nil
}

// UpdateNotificationRemarks updates the admin remarks and optionally an image for a notification
func (s *AdminNotificationService) UpdateNotificationRemarks(ctx context.Context, notificationID string, remarks string, adminImage string) error {
	collection := s.client.Collection("admin_notifications")

	now := time.Now()
	// Update the notification
	updates := []firestore.Update{
		{
			Path:  "adminRemarks",
			Value: remarks,
		},
		{
			Path:  "updatedAt",
			Value: now,
		},
	}

	// Only add adminImage if it's provided (not empty)
	if adminImage != "" {
		updates = append(updates, firestore.Update{
			Path:  "adminImage",
			Value: adminImage,
		})
	}

	_, err := collection.Doc(notificationID).Update(ctx, updates)

	if err != nil {
		return err
	}

	return nil
}
