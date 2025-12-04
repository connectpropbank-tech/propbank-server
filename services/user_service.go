package services

import (
	"context"
	"fmt"
	"shoprop-backend/models"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

// UserService handles user-related operations
type UserService struct {
	client *firestore.Client
}

// NewUserService creates a new UserService
func NewUserService(client *firestore.Client) *UserService {
	return &UserService{
		client: client,
	}
}

// GetAllUsers fetches all users from Firestore
func (us *UserService) GetAllUsers(ctx context.Context) ([]models.User, error) {
	var users []models.User

	iter := us.client.Collection("users").Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, err
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			continue
		}

		// Set the UID from document ID if not set
		if user.UID == "" {
			user.UID = doc.Ref.ID
		}

		users = append(users, user)
	}

	return users, nil
}

// GetUserByID fetches a single user by UID
func (us *UserService) GetUserByID(ctx context.Context, uid string) (*models.User, error) {
	doc, err := us.client.Collection("users").Doc(uid).Get(ctx)
	if err != nil {
		return nil, err
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, err
	}

	// Set the UID from document ID if not set
	if user.UID == "" {
		user.UID = doc.Ref.ID
	}

	return &user, nil
}

// CreateOrUpdateUser creates or updates a user in Firestore
func (us *UserService) CreateOrUpdateUser(ctx context.Context, user *models.User) error {
	_, err := us.client.Collection("users").Doc(user.UID).Set(ctx, user)
	return err
}

// GetUserByPhoneNumber fetches a user by phone number
func (us *UserService) GetUserByPhoneNumber(ctx context.Context, phoneNumber string) (*models.User, error) {
	// Normalize phone number (remove spaces, dashes, etc.)
	normalizedPhone := normalizePhoneNumber(phoneNumber)

	// Search for user by phone number
	iter := us.client.Collection("users").Where("phoneNumber", "==", normalizedPhone).Documents(ctx)
	defer iter.Stop()

	// Get first matching user
	doc, err := iter.Next()
	if err == iterator.Done {
		return nil, fmt.Errorf("user not found with phone number: %s", phoneNumber)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to search user by phone: %v", err)
	}

	var user models.User
	if err := doc.DataTo(&user); err != nil {
		return nil, fmt.Errorf("failed to parse user data: %v", err)
	}

	// Set the UID from document ID if not set
	if user.UID == "" {
		user.UID = doc.Ref.ID
	}

	return &user, nil
}

// NormalizePhoneNumber normalizes phone number for consistent searching and storage
// This is exported so it can be used by handlers to normalize phone numbers before storing
func NormalizePhoneNumber(phone string) string {
	// Remove all non-digit characters except +
	normalized := ""
	for _, char := range phone {
		if char >= '0' && char <= '9' || char == '+' {
			normalized += string(char)
		}
	}
	return normalized
}

// normalizePhoneNumber is a private wrapper for backwards compatibility
func normalizePhoneNumber(phone string) string {
	return NormalizePhoneNumber(phone)
}
