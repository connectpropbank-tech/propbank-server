package models

import "time"

// UserRole represents the different types of users in the system
type UserRole string

const (
	RoleAdmin      UserRole = "admin"
	RoleAgent      UserRole = "agent"
	RoleIndividual UserRole = "individual"
)

// User represents a user in the system
type User struct {
	UID         string   `json:"uid" firestore:"uid"`
	Email       string   `json:"email" firestore:"email"`
	Name        string   `json:"name" firestore:"name"`
	PhotoURL    string   `json:"photoURL" firestore:"photoURL"`
	PhoneNumber string   `json:"phoneNumber" firestore:"phoneNumber"`
	Role        UserRole `json:"role" firestore:"role"`

	// Rented Property Information (if user is a tenant)
	RentedPropertyID        string `json:"rentedPropertyId,omitempty" firestore:"rentedPropertyId,omitempty"`
	RentedPropertyOwnerID   string `json:"rentedPropertyOwnerId,omitempty" firestore:"rentedPropertyOwnerId,omitempty"`
	RentedPropertyOwnerName string `json:"rentedPropertyOwnerName,omitempty" firestore:"rentedPropertyOwnerName,omitempty"`

	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" firestore:"updatedAt"`
	IsActive  bool      `json:"isActive" firestore:"isActive"`
}

// UserResponse represents the response structure for user data
type UserResponse struct {
	Success bool   `json:"success"`
	Message string `json:"message"`
	Data    []User `json:"data,omitempty"`
	User    *User  `json:"user,omitempty"`
}
