package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"shoprop-backend/models"
	"shoprop-backend/services"
	"time"

	"cloud.google.com/go/firestore"
)

// CreateUserRequest represents the request body for creating/updating a user
type CreateUserRequest struct {
	UID         string          `json:"uid"`
	Email       string          `json:"email"`
	Name        string          `json:"name"`
	PhotoURL    string          `json:"photoURL"`
	PhoneNumber string          `json:"phoneNumber"`
	Role        models.UserRole `json:"role"`
}

// AuthHandler handles authentication-related requests
type AuthHandler struct {
	userService *services.UserService
}

// NewAuthHandler creates a new AuthHandler
func NewAuthHandler(client *firestore.Client) *AuthHandler {
	return &AuthHandler{
		userService: services.NewUserService(client),
	}
}

// CreateOrUpdateUser handles POST /auth/user - creates or updates user after Firebase auth
func (ah *AuthHandler) CreateOrUpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request body: %v", err)
		response := models.UserResponse{
			Success: false,
			Message: "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate required fields
	if req.UID == "" || req.Email == "" || req.Name == "" || req.PhoneNumber == "" {
		response := models.UserResponse{
			Success: false,
			Message: "Missing required fields: uid, email, name, phoneNumber",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate role - set default if not provided
	if req.Role == "" {
		req.Role = models.RoleIndividual // Default role
	}

	// Validate role is one of the allowed values
	if req.Role != models.RoleAdmin && req.Role != models.RoleAgent && req.Role != models.RoleIndividual {
		response := models.UserResponse{
			Success: false,
			Message: "Invalid role. Must be 'admin', 'agent', or 'individual'",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx := context.Background()

	// Check if user already exists
	existingUser, err := ah.userService.GetUserByID(ctx, req.UID)
	isUpdate := err == nil && existingUser != nil

	// Normalize phone number before storing
	normalizedPhone := ""
	if req.PhoneNumber != "" {
		normalizedPhone = services.NormalizePhoneNumber(req.PhoneNumber)
	}

	// Create user model
	user := models.User{
		UID:         req.UID,
		Email:       req.Email,
		Name:        req.Name,
		PhotoURL:    req.PhotoURL,
		PhoneNumber: normalizedPhone,
		Role:        req.Role,
		IsActive:    true,
	}

	if isUpdate {
		// Update existing user (preserve role if it exists and req.Role is empty)
		user.CreatedAt = existingUser.CreatedAt
		user.UpdatedAt = time.Now()
		if existingUser.Role != "" && req.Role == models.RoleIndividual {
			user.Role = existingUser.Role // Preserve existing role if not explicitly changed
		}
		log.Printf("Updating existing user: %s with role: %s", req.UID, user.Role)
	} else {
		// Create new user
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
		log.Printf("Creating new user: %s with role: %s", req.UID, user.Role)
	}

	// Save to Firestore
	err = ah.userService.CreateOrUpdateUser(ctx, &user)
	if err != nil {
		log.Printf("Error saving user %s: %v", req.UID, err)
		response := models.UserResponse{
			Success: false,
			Message: "Failed to save user",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Success response
	message := "User created successfully"
	if isUpdate {
		message = "User updated successfully"
	}

	response := models.UserResponse{
		Success: true,
		Message: message,
		User:    &user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)

	log.Printf("✅ %s: %s (%s) - Phone: %s - Role: %s", message, user.Name, user.Email, user.PhoneNumber, user.Role)
}
