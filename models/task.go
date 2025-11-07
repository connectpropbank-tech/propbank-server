package models

import "time"

// TaskStatus represents the status of a task
type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusCancelled  TaskStatus = "cancelled"
	TaskStatusInProgress TaskStatus = "in_progress"
)

// TaskPriority represents the priority level of a task
type TaskPriority string

const (
	TaskPriorityLow    TaskPriority = "low"
	TaskPriorityMedium TaskPriority = "medium"
	TaskPriorityHigh   TaskPriority = "high"
	TaskPriorityUrgent TaskPriority = "urgent"
)

// VisitType represents the type of visit
type VisitType string

const (
	VisitTypePropertyShowing VisitType = "property_showing"
	VisitTypeInspection      VisitType = "inspection"
	VisitTypeMeeting         VisitType = "meeting"
	VisitTypeDocumentation   VisitType = "documentation"
	VisitTypeOther           VisitType = "other"
)

// Task represents a task/visit in the system
type Task struct {
	ID               string       `json:"id" firestore:"id"`
	UserID           string       `json:"userId" firestore:"userId"`
	PropertyID       string       `json:"propertyId,omitempty" firestore:"propertyId,omitempty"`
	Title            string       `json:"title" firestore:"title"`
	Description      string       `json:"description,omitempty" firestore:"description,omitempty"`
	VisitType        VisitType    `json:"visitType" firestore:"visitType"`
	Priority         TaskPriority `json:"priority" firestore:"priority"`
	Status           TaskStatus   `json:"status" firestore:"status"`
	ScheduledTime    time.Time    `json:"scheduledTime" firestore:"scheduledTime"`
	Duration         int          `json:"duration" firestore:"duration"` // Duration in minutes
	Location         string       `json:"location,omitempty" firestore:"location,omitempty"`
	ContactName      string       `json:"contactName,omitempty" firestore:"contactName,omitempty"`
	ContactPhone     string       `json:"contactPhone,omitempty" firestore:"contactPhone,omitempty"`
	ContactEmail     string       `json:"contactEmail,omitempty" firestore:"contactEmail,omitempty"`
	Notes            string       `json:"notes,omitempty" firestore:"notes,omitempty"`
	ReminderSent     bool         `json:"reminderSent" firestore:"reminderSent"`
	NotificationSent bool         `json:"notificationSent" firestore:"notificationSent"`
	CreatedAt        time.Time    `json:"createdAt" firestore:"createdAt"`
	UpdatedAt        time.Time    `json:"updatedAt" firestore:"updatedAt"`
}

// CreateTaskRequest represents the request body for creating a task
type CreateTaskRequest struct {
	PropertyID    string       `json:"propertyId,omitempty"`
	Title         string       `json:"title"`
	Description   string       `json:"description,omitempty"`
	VisitType     VisitType    `json:"visitType"`
	Priority      TaskPriority `json:"priority"`
	ScheduledTime string       `json:"scheduledTime"` // ISO 8601 format
	Duration      int          `json:"duration"`      // Duration in minutes
	Location      string       `json:"location,omitempty"`
	ContactName   string       `json:"contactName,omitempty"`
	ContactPhone  string       `json:"contactPhone,omitempty"`
	ContactEmail  string       `json:"contactEmail,omitempty"`
	Notes         string       `json:"notes,omitempty"`
}

// UpdateTaskRequest represents the request body for updating a task
type UpdateTaskRequest struct {
	Title         *string       `json:"title,omitempty"`
	Description   *string       `json:"description,omitempty"`
	VisitType     *VisitType    `json:"visitType,omitempty"`
	Priority      *TaskPriority `json:"priority,omitempty"`
	Status        *TaskStatus   `json:"status,omitempty"`
	ScheduledTime *string       `json:"scheduledTime,omitempty"` // ISO 8601 format
	Duration      *int          `json:"duration,omitempty"`      // Duration in minutes
	Location      *string       `json:"location,omitempty"`
	ContactName   *string       `json:"contactName,omitempty"`
	ContactPhone  *string       `json:"contactPhone,omitempty"`
	ContactEmail  *string       `json:"contactEmail,omitempty"`
	Notes         *string       `json:"notes,omitempty"`
}

// TaskResponse represents the API response for task operations
type TaskResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Task    *Task  `json:"task,omitempty"`
	Tasks   []Task `json:"tasks,omitempty"`
}

// TaskFilters represents filters for querying tasks
type TaskFilters struct {
	Status     []TaskStatus   `json:"status,omitempty"`
	Priority   []TaskPriority `json:"priority,omitempty"`
	VisitType  []VisitType    `json:"visitType,omitempty"`
	StartDate  *time.Time     `json:"startDate,omitempty"`
	EndDate    *time.Time     `json:"endDate,omitempty"`
	PropertyID string         `json:"propertyId,omitempty"`
}
