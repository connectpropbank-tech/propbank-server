package models

import (
	"time"
)

// ArchiveHistoryEntry tracks when a notification was archived or unarchived
type ArchiveHistoryEntry struct {
	Action    string    `firestore:"action" json:"action"`       // "archived" or "unarchived"
	Timestamp time.Time `firestore:"timestamp" json:"timestamp"` // When the action happened
}

// AdminNotification represents an admin notification in the database
type AdminNotification struct {
	ID         string `firestore:"id" json:"id"`
	Type       string `firestore:"type" json:"type"`
	Title      string `firestore:"title" json:"title"`
	Message    string `firestore:"message" json:"message"`
	PropertyID string `firestore:"propertyId" json:"propertyId"`
	OwnerID    string `firestore:"ownerId" json:"ownerId"`
	OwnerName  string `firestore:"ownerName" json:"ownerName"`
	OwnerPhone string `firestore:"ownerPhone,omitempty" json:"ownerPhone,omitempty"`
	OwnerEmail string `firestore:"ownerEmail,omitempty" json:"ownerEmail,omitempty"`
	OwnerRole  string `firestore:"ownerRole,omitempty" json:"ownerRole,omitempty"`
	// User details for property enquiry notifications
	UserID    string `firestore:"userId,omitempty" json:"userId,omitempty"`
	UserName  string `firestore:"userName,omitempty" json:"userName,omitempty"`
	UserEmail string `firestore:"userEmail,omitempty" json:"userEmail,omitempty"`
	UserPhone string `firestore:"userPhone,omitempty" json:"userPhone,omitempty"`
	// Tenant details (if property is rented)
	TenantName  string `firestore:"tenantName,omitempty" json:"tenantName,omitempty"`
	TenantPhone string `firestore:"tenantPhone,omitempty" json:"tenantPhone,omitempty"`
	TenantEmail string `firestore:"tenantEmail,omitempty" json:"tenantEmail,omitempty"`
	Buyers      string `firestore:"buyers,omitempty" json:"buyers,omitempty"`
	// Property details for property enquiry notifications
	PropertyTitle       string `firestore:"propertyTitle,omitempty" json:"propertyTitle,omitempty"`
	PropertyAddress     string `firestore:"propertyAddress,omitempty" json:"propertyAddress,omitempty"`
	PropertyListingType string `firestore:"propertyListingType,omitempty" json:"propertyListingType,omitempty"`
	// Service request specific fields
	ServiceType    string `firestore:"serviceType,omitempty" json:"serviceType,omitempty"`
	ServiceComment string `firestore:"serviceComment,omitempty" json:"serviceComment,omitempty"`
	ServiceImage   string `firestore:"serviceImage,omitempty" json:"serviceImage,omitempty"`
	// General inquiry specific fields
	InquiryType  string     `firestore:"inquiryType,omitempty" json:"inquiryType,omitempty"`
	PropertyType string     `firestore:"propertyType,omitempty" json:"propertyType,omitempty"`
	RequestVisit bool       `firestore:"requestVisit,omitempty" json:"requestVisit,omitempty"`
	VisitDate    string     `firestore:"visitDate,omitempty" json:"visitDate,omitempty"`
	VisitTime    string     `firestore:"visitTime,omitempty" json:"visitTime,omitempty"`
	Timestamp    time.Time             `firestore:"timestamp" json:"timestamp"`
	IsRead       bool                  `firestore:"isRead" json:"isRead"`
	ResolvedAt   *time.Time            `firestore:"resolvedAt,omitempty" json:"resolvedAt,omitempty"`
	Priority     string                `firestore:"priority" json:"priority"`
	AdminRemarks string                `firestore:"adminRemarks,omitempty" json:"adminRemarks,omitempty"`
	AdminImage   string                `firestore:"adminImage,omitempty" json:"adminImage,omitempty"`
	ArchiveHistory []ArchiveHistoryEntry `firestore:"archiveHistory,omitempty" json:"archiveHistory,omitempty"`
	CreatedAt    time.Time             `firestore:"createdAt" json:"createdAt"`
	UpdatedAt    time.Time             `firestore:"updatedAt" json:"updatedAt"`
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
	OwnerRole  string `json:"ownerRole,omitempty"`
	// User details for property enquiry notifications
	UserID    string `json:"userId,omitempty"`
	UserName  string `json:"userName,omitempty"`
	UserEmail string `json:"userEmail,omitempty"`
	UserPhone string `json:"userPhone,omitempty"`
	// Tenant details (if property is rented)
	TenantName  string `json:"tenantName,omitempty"`
	TenantPhone string `json:"tenantPhone,omitempty"`
	TenantEmail string `json:"tenantEmail,omitempty"`
	// Buyer details (for property enquiries/offers)
	Buyers string `json:"buyers,omitempty"` // JSON stringified array of BuyerInfo
	// Property details for property enquiry notifications
	PropertyTitle       string `json:"propertyTitle,omitempty"`
	PropertyAddress     string `json:"propertyAddress,omitempty"`
	PropertyListingType string `json:"propertyListingType,omitempty"`
	// Service request specific fields
	ServiceType    string `json:"serviceType,omitempty"`
	ServiceComment string `json:"serviceComment,omitempty"`
	ServiceImage   string `json:"serviceImage,omitempty"`
	// General inquiry specific fields
	InquiryType  string `json:"inquiryType,omitempty"`
	PropertyType string `json:"propertyType,omitempty"`
	RequestVisit bool   `json:"requestVisit,omitempty"`
	VisitDate    string `json:"visitDate,omitempty"`
	VisitTime    string `json:"visitTime,omitempty"`
	Timestamp    string `json:"timestamp" validate:"required"`
	IsRead       bool   `json:"isRead"`
	Priority     string `json:"priority"`
	AdminRemarks string `json:"adminRemarks,omitempty"`
}
