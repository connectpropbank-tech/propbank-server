package services

import (
	"context"
	"log"
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
			log.Printf("Error iterating users: %v", err)
			return nil, err
		}

		var user models.User
		if err := doc.DataTo(&user); err != nil {
			log.Printf("Error converting document to user: %v", err)
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
