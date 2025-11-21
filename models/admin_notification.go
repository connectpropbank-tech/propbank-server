package models

import (
	"time"
)

// AdminNotification represents an admin notification in the database
type AdminNotification struct {
	ID          string    `firestore:"id" json:"id"`
	Type        string    `firestore:"type" json:"type"`
	Title       string    `firestore:"title" json:"title"`
	Message     string    `firestore:"message" json:"message"`
	PropertyID  string    `firestore:"propertyId" json:"propertyId"`
	OwnerID     string    `firestore:"ownerId" json:"ownerId"`
	OwnerName   string    `firestore:"ownerName" json:"ownerName"`
	OwnerPhone  string    `firestore:"ownerPhone,omitempty" json:"ownerPhone,omitempty"`
	OwnerEmail  string    `firestore:"ownerEmail,omitempty" json:"ownerEmail,omitempty"`
	// User details for property enquiry notifications
	UserID      string    `firestore:"userId,omitempty" json:"userId,omitempty"`
	UserName    string    `firestore:"userName,omitempty" json:"userName,omitempty"`
	UserEmail   string    `firestore:"userEmail,omitempty" json:"userEmail,omitempty"`
	UserPhone   string    `firestore:"userPhone,omitempty" json:"userPhone,omitempty"`
	// Property details for property enquiry notifications
	PropertyTitle string  `firestore:"propertyTitle,omitempty" json:"propertyTitle,omitempty"`
	PropertyAddress string `firestore:"propertyAddress,omitempty" json:"propertyAddress,omitempty"`
	PropertyListingType string `firestore:"propertyListingType,omitempty" json:"propertyListingType,omitempty"`
	// Service request specific fields
	ServiceType    string `firestore:"serviceType,omitempty" json:"serviceType,omitempty"`
	ServiceComment string `firestore:"serviceComment,omitempty" json:"serviceComment,omitempty"`
	ServiceImage   string `firestore:"serviceImage,omitempty" json:"serviceImage,omitempty"`
	Timestamp   time.Time `firestore:"timestamp" json:"timestamp"`
	IsRead      bool      `firestore:"isRead" json:"isRead"`
	Priority    string    `firestore:"priority" json:"priority"`
	CreatedAt   time.Time `firestore:"createdAt" json:"createdAt"`
	UpdatedAt   time.Time `firestore:"updatedAt" json:"updatedAt"`
}

// CreateAdminNotificationRequest represents the request body for creating an admin notification
type CreateAdminNotificationRequest struct {
	Type       string `json:"type" validate:"required"`
	Title      string `json:"title" validate:"required"`
	Message    string `json:"message" validate:"required"`
	PropertyID string `json:"propertyId"` // Optional for service requests
	OwnerID    string `json:"ownerId"`    // Optional for service requests
	OwnerName  string `json:"ownerName"`  // Optional for service requests
	OwnerPhone string `json:"ownerPhone,omitempty"`
	OwnerEmail string `json:"ownerEmail,omitempty"`
	// User details for property enquiry notifications
	UserID      string `json:"userId,omitempty"`
	UserName    string `json:"userName,omitempty"`
	UserEmail   string `json:"userEmail,omitempty"`
	UserPhone   string `json:"userPhone,omitempty"`
	// Property details for property enquiry notifications
	PropertyTitle string `json:"propertyTitle,omitempty"`
	PropertyAddress string `json:"propertyAddress,omitempty"`
	PropertyListingType string `json:"propertyListingType,omitempty"`
	// Service request specific fields
	ServiceType    string `json:"serviceType,omitempty"`
	ServiceComment string `json:"serviceComment,omitempty"`
	ServiceImage   string `json:"serviceImage,omitempty"`
	Timestamp  string `json:"timestamp" validate:"required"`
	IsRead     bool   `json:"isRead"`
	Priority   string `json:"priority"`
}
