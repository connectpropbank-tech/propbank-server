package handlers

import (
	"context"
	"encoding/json"
	"log"
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
		log.Printf("Error fetching users: %v", err)
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
		log.Printf("Error fetching user %s: %v", path, err)
		response := models.UserResponse{
			Success: false,
			Message: "User not found",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
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
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
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
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if userReq.UID == "" || userReq.Email == "" || userReq.Name == "" {
		http.Error(w, "Missing required fields: uid, email, name", http.StatusBadRequest)
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
		log.Printf("Error creating/updating user %s: %v", userReq.UID, err)
		response := models.UserResponse{
			Success: false,
			Message: "Failed to create/update user",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(response)
		return
	}

	log.Printf("✅ User created/updated successfully: %s (%s) - Role: %s", user.Name, user.Email, user.Role)

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
		log.Printf("User not found with phone %s: %v", phoneNumber, err)
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
