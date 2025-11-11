package models

import (
	"time"
)

// AdminNotification represents an admin notification in the database
type AdminNotification struct {
	ID         string    `firestore:"id" json:"id"`
	Type       string    `firestore:"type" json:"type"`
	Title      string    `firestore:"title" json:"title"`
	Message    string    `firestore:"message" json:"message"`
	PropertyID string    `firestore:"propertyId" json:"propertyId"`
	OwnerID    string    `firestore:"ownerId" json:"ownerId"`
	OwnerName  string    `firestore:"ownerName" json:"ownerName"`
	Timestamp  time.Time `firestore:"timestamp" json:"timestamp"`
	IsRead     bool      `firestore:"isRead" json:"isRead"`
	Priority   string    `firestore:"priority" json:"priority"`
	CreatedAt  time.Time `firestore:"createdAt" json:"createdAt"`
	UpdatedAt  time.Time `firestore:"updatedAt" json:"updatedAt"`
}

// CreateAdminNotificationRequest represents the request body for creating an admin notification
type CreateAdminNotificationRequest struct {
	Type       string `json:"type" validate:"required"`
	Title      string `json:"title" validate:"required"`
	Message    string `json:"message" validate:"required"`
	PropertyID string `json:"propertyId" validate:"required"`
	OwnerID    string `json:"ownerId" validate:"required"`
	OwnerName  string `json:"ownerName" validate:"required"`
	Timestamp  string `json:"timestamp" validate:"required"`
	IsRead     bool   `json:"isRead"`
	Priority   string `json:"priority"`
}
