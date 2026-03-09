package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"shoprop-backend/models"
	"shoprop-backend/services"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
)

// UserHandler handles user-related HTTP requests
type UserHandler struct {
	userService *services.UserService
}

// NewUserHandler creates a new UserHandler
func NewUserHandler(client *firestore.Client) *UserHandler {
	return &UserHandler{
		userService: services.NewUserService(client),
	}
}

// GetAllUsers handles GET /users - fetches all users
func (uh *UserHandler) GetAllUsers(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ctx := context.Background()
	users, err := uh.userService.GetAllUsers(ctx)
	if err != nil {
		response := models.UserResponse{
			Success: false,
			Message: "Failed to fetch users",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "Users fetched successfully",
		Data:    users,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetUser handles GET /users/{uid} - fetches a single user
func (uh *UserHandler) GetUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract UID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	if path == "" || path == r.URL.Path {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()
	user, err := uh.userService.GetUserByID(ctx, path)
	if err != nil {
		response := models.UserResponse{
			Success: false,
			Message: "User not registered",
		}
		w.Header().Set("Content-Type", "application/json")
		// Return 200 OK instead of 404 to avoid console errors in frontend
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User fetched successfully",
		User:    user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CreateUser handles POST /users - creates or updates a user
func (uh *UserHandler) CreateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		response := models.UserResponse{
			Success: false,
			Message: "Method not allowed",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	var userReq struct {
		UID         string          `json:"uid"`
		Email       string          `json:"email"`
		Name        string          `json:"name"`
		PhotoURL    string          `json:"photoURL"`
		PhoneNumber string          `json:"phoneNumber"`
		Role        models.UserRole `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		response := models.UserResponse{
			Success: false,
			Message: "Invalid JSON",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Validate required fields
	if userReq.UID == "" || userReq.Email == "" || userReq.Name == "" {
		response := models.UserResponse{
			Success: false,
			Message: "Missing required fields: uid, email, name",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx := context.Background()

	// Set default role if not provided
	if userReq.Role == "" {
		userReq.Role = models.RoleIndividual
	}

	// Normalize phone number before storing
	normalizedPhone := ""
	if userReq.PhoneNumber != "" {
		normalizedPhone = services.NormalizePhoneNumber(userReq.PhoneNumber)

		// Check if this phone number is already registered to another user
		ctx := context.Background()
		existingUserWithPhone, err := uh.userService.GetUserByPhoneNumber(ctx, normalizedPhone)
		if err == nil && existingUserWithPhone != nil && existingUserWithPhone.UID != userReq.UID {
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

	// Create user object
	user := &models.User{
		UID:         userReq.UID,
		Email:       userReq.Email,
		Name:        userReq.Name,
		PhotoURL:    userReq.PhotoURL,
		PhoneNumber: normalizedPhone,
		Role:        userReq.Role,
		IsActive:    true,
	}

	// Set timestamps
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	err := uh.userService.CreateOrUpdateUser(ctx, user)
	if err != nil {
		response := models.UserResponse{
			Success: false,
			Message: "Failed to create/update user",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User created/updated successfully",
		User:    user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(response)
}

// SearchUserByPhone handles GET /users/search?phone={phone} - searches for a user by phone number
func (uh *UserHandler) SearchUserByPhone(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	phoneNumber := r.URL.Query().Get("phone")
	if phoneNumber == "" {
		response := models.UserResponse{
			Success: false,
			Message: "Phone number is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx := context.Background()
	user, err := uh.userService.GetUserByPhoneNumber(ctx, phoneNumber)
	if err != nil {
		response := models.UserResponse{
			Success: false,
			Message: "User is not found. Please ask to sign up with our platform to continue.",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User found successfully",
		User:    user,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateUser handles PUT/PATCH /users/{uid} - creates or updates a user (Upsert)
func (uh *UserHandler) UpdateUser(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		response := models.UserResponse{
			Success: false,
			Message: "Method not allowed",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusMethodNotAllowed)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Extract UID from URL path
	// Path comes in as /users/{uid}
	path := strings.TrimPrefix(r.URL.Path, "/users/")
	if path == "" {
		response := models.UserResponse{
			Success: false,
			Message: "User ID is required",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	var userReq struct {
		Email       string          `json:"email"`
		Name        string          `json:"name"`
		PhotoURL    string          `json:"photoURL"`
		PhoneNumber string          `json:"phoneNumber"`
		Role        models.UserRole `json:"role"`
	}

	if err := json.NewDecoder(r.Body).Decode(&userReq); err != nil {
		response := models.UserResponse{
			Success: false,
			Message: "Invalid JSON",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(response)
		return
	}

	ctx := context.Background()

	// Try to get existing user
	existingUser, err := uh.userService.GetUserByID(ctx, path)

	// If user does not exist or error, we prepare to create new
	if err != nil || existingUser == nil {
		// Validation for creation
		if userReq.Name == "" {
			response := models.UserResponse{
				Success: false,
				Message: "Creating new user requires a name",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}
		if userReq.Email == "" {
			response := models.UserResponse{
				Success: false,
				Message: "Creating new user requires an email",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(response)
			return
		}

		// Create new user model
		newUser := &models.User{
			UID:         path, // Use path UID
			Email:       userReq.Email,
			Name:        userReq.Name,
			PhotoURL:    userReq.PhotoURL,
			PhoneNumber: services.NormalizePhoneNumber(userReq.PhoneNumber),
			Role:        userReq.Role,
			IsActive:    true,
			CreatedAt:   time.Now(),
			UpdatedAt:   time.Now(),
		}

		if newUser.Role == "" {
			newUser.Role = models.RoleIndividual
		}

		err = uh.userService.CreateOrUpdateUser(ctx, newUser)
		if err != nil {
			response := models.UserResponse{
				Success: false,
				Message: "Failed to create user",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(response)
			return
		}

		response := models.UserResponse{
			Success: true,
			Message: "User created successfully",
			User:    newUser,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(response)
		return
	}

	// Update existing fields if present
	if userReq.Name != "" {
		existingUser.Name = userReq.Name
	}
	if userReq.PhotoURL != "" {
		existingUser.PhotoURL = userReq.PhotoURL
	}
	if userReq.Email != "" {
		existingUser.Email = userReq.Email
	}
	if userReq.PhoneNumber != "" {
		normalizedPhone := services.NormalizePhoneNumber(userReq.PhoneNumber)

		// Check if this phone number is already registered to another user
		existingUserWithPhone, err := uh.userService.GetUserByPhoneNumber(ctx, normalizedPhone)
		if err == nil && existingUserWithPhone != nil && existingUserWithPhone.UID != path {
			response := models.UserResponse{
				Success: false,
				Message: "This phone number is already registered with another account",
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusConflict)
			json.NewEncoder(w).Encode(response)
			return
		}

		existingUser.PhoneNumber = normalizedPhone
	}

	// Only allow role update if it's currently unset or empty, or if we have admin logic (skipped for now)
	if userReq.Role != "" {
		// Basic validation for role
		if userReq.Role == models.RoleAdmin || userReq.Role == models.RoleAgent || userReq.Role == models.RoleIndividual {
			existingUser.Role = userReq.Role
		}
	}

	existingUser.UpdatedAt = time.Now()

	// Save updates
	err = uh.userService.CreateOrUpdateUser(ctx, existingUser)
	if err != nil {
		response := models.UserResponse{
			Success: false,
			Message: "Failed to update user",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	response := models.UserResponse{
		Success: true,
		Message: "User updated successfully",
		User:    existingUser,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
