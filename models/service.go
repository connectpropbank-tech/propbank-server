package models

import "time"

// ServiceCategory represents different service categories
type ServiceCategory string

const (
	CategoryLegal ServiceCategory = "legal"
	CategoryOther ServiceCategory = "other"
)

// ServiceStatus represents the status of a service request
type ServiceStatus string

const (
	StatusPending    ServiceStatus = "pending"
	StatusInProgress ServiceStatus = "in_progress"
	StatusCompleted  ServiceStatus = "completed"
	StatusCancelled  ServiceStatus = "cancelled"
)

// Service represents a service offered by the platform
type Service struct {
	ID          string          `json:"id" firestore:"id"`
	Name        string          `json:"name" firestore:"name"`
	Description string          `json:"description" firestore:"description"`
	Category    ServiceCategory `json:"category" firestore:"category"`
	Code        string          `json:"code" firestore:"code"` // e.g., "1A", "2B"
	IsActive    bool            `json:"isActive" firestore:"isActive"`
	CreatedAt   time.Time       `json:"createdAt" firestore:"createdAt"`
	UpdatedAt   time.Time       `json:"updatedAt" firestore:"updatedAt"`
}

// ServiceRequest represents a user's request for a service
type ServiceRequest struct {
	ID          string        `json:"id" firestore:"id"`
	UserUID     string        `json:"userUID" firestore:"userUID"`
	UserName    string        `json:"userName" firestore:"userName"`
	UserEmail   string        `json:"userEmail" firestore:"userEmail"`
	UserPhone   string        `json:"userPhone" firestore:"userPhone"`
	ServiceID   string        `json:"serviceId" firestore:"serviceId"`
	ServiceName string        `json:"serviceName" firestore:"serviceName"`
	PropertyID  string        `json:"propertyId,omitempty" firestore:"propertyId,omitempty"`
	Message     string        `json:"message" firestore:"message"`
	Image       string        `json:"image,omitempty" firestore:"image,omitempty"` // Image URL for service request
	Status      ServiceStatus `json:"status" firestore:"status"`
	AdminNotes  string        `json:"adminNotes,omitempty" firestore:"adminNotes,omitempty"`
	CreatedAt   time.Time     `json:"createdAt" firestore:"createdAt"`
	UpdatedAt   time.Time     `json:"updatedAt" firestore:"updatedAt"`
}

// ServiceResponse represents the response structure for service data
type ServiceResponse struct {
	Success bool      `json:"success"`
	Message string    `json:"message"`
	Data    []Service `json:"data,omitempty"`
	Service *Service  `json:"service,omitempty"`
}

// ServiceRequestResponse represents the response structure for service request data
type ServiceRequestResponse struct {
	Success        bool             `json:"success"`
	Message        string           `json:"message"`
	Data           []ServiceRequest `json:"data,omitempty"`
	ServiceRequest *ServiceRequest  `json:"serviceRequest,omitempty"`
}
