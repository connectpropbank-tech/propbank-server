package models

import "time"

// Visit represents a scheduled property visit
type Visit struct {
	ID           string    `json:"id" firestore:"id"`
	UserID       string    `json:"userId" firestore:"userId"`
	Title        string    `json:"title" firestore:"title"`
	Description  string    `json:"description" firestore:"description"`
	PropertyID   string    `json:"propertyId,omitempty" firestore:"propertyId,omitempty"`
	VisitDate    time.Time `json:"visitDate" firestore:"visitDate"`
	ReminderType string    `json:"reminderType" firestore:"reminderType"` // email, sms, both
	ReminderTime int       `json:"reminderTime" firestore:"reminderTime"` // minutes before visit
	Status       string    `json:"status" firestore:"status"`             // scheduled, completed, cancelled
	IsCompleted  bool      `json:"isCompleted" firestore:"isCompleted"`
	CreatedAt    time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt    time.Time `json:"updatedAt" firestore:"updatedAt"`
	NotifiedAt   time.Time `json:"notifiedAt,omitempty" firestore:"notifiedAt,omitempty"` // when notification was sent
}

// VisitResponse represents the API response structure for visits
type VisitResponse struct {
	Success bool    `json:"success"`
	Message string  `json:"message"`
	Visit   *Visit  `json:"visit,omitempty"`
	Visits  []Visit `json:"visits,omitempty"`
}

// CreateVisitRequest represents the request body for creating a visit
type CreateVisitRequest struct {
	Title        string `json:"title"`
	Description  string `json:"description,omitempty"`
	PropertyID   string `json:"propertyId,omitempty"`
	VisitDate    string `json:"visitDate"` // ISO 8601 format
	ReminderType string `json:"reminderType"`
	ReminderTime int    `json:"reminderTime"`
}

// UpdateVisitRequest represents the request body for updating a visit
type UpdateVisitRequest struct {
	Title        string `json:"title,omitempty"`
	Description  string `json:"description,omitempty"`
	VisitDate    string `json:"visitDate,omitempty"`
	ReminderType string `json:"reminderType,omitempty"`
	ReminderTime int    `json:"reminderTime,omitempty"`
	Status       string `json:"status,omitempty"`
	IsCompleted  bool   `json:"isCompleted,omitempty"`
}
