package handlers

import (
	"context"
	"encoding/json"
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

// LoginRequest represents the login request body
type LoginRequest struct {
	UID string `json:"uid"`
}

// LoginResponse represents the login response
type LoginResponse struct {
	Success bool         `json:"success"`
	Token   string       `json:"token"`
	User    *models.User `json:"user,omitempty"`
	Message string       `json:"message"`
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

		// Check if this phone number is already registered to another user
		existingUserWithPhone, err := ah.userService.GetUserByPhoneNumber(ctx, normalizedPhone)
		if err == nil && existingUserWithPhone != nil && existingUserWithPhone.UID != req.UID {
			response := models.UserResponse{
				Success: false,
				Message: "This phone number is already registered with another account",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)
			return
		}
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
		// Update existing user timestamps
		user.CreatedAt = existingUser.CreatedAt
		user.UpdatedAt = time.Now()

		// If this user already has a complete registration (phone + role),
		// NEVER overwrite those fields. This prevents re-registration with
		// a different phone number using the same Google account.
		if existingUser.PhoneNumber != "" && string(existingUser.Role) != "" {
			user.PhoneNumber = existingUser.PhoneNumber
			user.Role = existingUser.Role
		} else if existingUser.Role != "" && req.Role == models.RoleIndividual {
			// Preserve existing role if not explicitly changed
			user.Role = existingUser.Role
		}
	} else {
		// Create new user
		user.CreatedAt = time.Now()
		user.UpdatedAt = time.Now()
	}

	// Save to Firestore
	err = ah.userService.CreateOrUpdateUser(ctx, &user)
	if err != nil {
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

}

// Login handles POST /auth/login - issues JWT token for existing user
func (ah *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response := LoginResponse{
			Success: false,
			Message: "Invalid request body",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	if req.UID == "" {
		response := LoginResponse{
			Success: false,
			Message: "UID is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx := context.Background()
	user, err := ah.userService.GetUserByID(ctx, req.UID)
	if err != nil || user == nil {
		response := LoginResponse{
			Success: false,
			Message: "User not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusUnauthorized)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Generate JWT Token
	token, err := services.GenerateToken(user.UID, string(user.Role))
	if err != nil {
		response := LoginResponse{
			Success: false,
			Message: "Failed to generate token",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := LoginResponse{
		Success: true,
		Token:   token,
		User:    user,
		Message: "Login successful",
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
