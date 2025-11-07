package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"shoprop-backend/models"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type VisitService struct {
	client *firestore.Client
}

// NewVisitService creates a new visit service
func NewVisitService(client *firestore.Client) *VisitService {
	return &VisitService{
		client: client,
	}
}

// GenerateID generates a new document ID for visit in format: 4 capitals + 4 numbers (e.g., VISI1234)
func (vs *VisitService) GenerateID() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const numbers = "0123456789"

	// Create a new random source
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Generate unique ID (retry if duplicate exists)
	for attempts := 0; attempts < 10; attempts++ {
		// Generate 4 random capital letters
		var result strings.Builder
		for i := 0; i < 4; i++ {
			result.WriteByte(letters[rng.Intn(len(letters))])
		}

		// Generate 4 random numbers
		for i := 0; i < 4; i++ {
			result.WriteByte(numbers[rng.Intn(len(numbers))])
		}

		generatedID := result.String()

		// Check if this ID already exists
		_, err := vs.client.Collection("visits").Doc(generatedID).Get(context.Background())
		if err != nil {
			// Document doesn't exist, we can use this ID
			return generatedID
		}
		// Document exists, try again
	}

	// Fallback to timestamp-based ID if all attempts fail
	return fmt.Sprintf("VISIT%d", time.Now().UnixNano()%100000000)
}

// CreateVisit creates a new visit in Firestore
func (vs *VisitService) CreateVisit(ctx context.Context, visit *models.Visit) error {
	// Generate custom visit ID (8 characters)
	visit.ID = vs.GenerateID()
	visit.CreatedAt = time.Now()
	visit.UpdatedAt = time.Now()
	visit.Status = "scheduled"
	visit.IsCompleted = false

	_, err := vs.client.Collection("visits").Doc(visit.ID).Set(ctx, visit)
	if err != nil {
		log.Printf("❌ Error creating visit: %v", err)
		return fmt.Errorf("failed to create visit: %v", err)
	}

	log.Printf("✅ Visit created successfully: %s", visit.ID)
	return nil
}

// GetVisitByID retrieves a visit by its ID
func (vs *VisitService) GetVisitByID(ctx context.Context, visitID string) (*models.Visit, error) {
	doc, err := vs.client.Collection("visits").Doc(visitID).Get(ctx)
	if err != nil {
		log.Printf("❌ Error getting visit %s: %v", visitID, err)
		return nil, fmt.Errorf("visit not found: %v", err)
	}

	var visit models.Visit
	if err := doc.DataTo(&visit); err != nil {
		log.Printf("❌ Error unmarshaling visit %s: %v", visitID, err)
		return nil, fmt.Errorf("failed to unmarshal visit: %v", err)
	}

	return &visit, nil
}

// GetVisitsByUserID retrieves all visits for a specific user
func (vs *VisitService) GetVisitsByUserID(ctx context.Context, userID string) ([]models.Visit, error) {
	iter := vs.client.Collection("visits").Where("userId", "==", userID).OrderBy("visitDate", firestore.Asc).Documents(ctx)
	defer iter.Stop()

	var visits []models.Visit
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("❌ Error iterating visits for user %s: %v", userID, err)
			return nil, fmt.Errorf("failed to get visits: %v", err)
		}

		var visit models.Visit
		if err := doc.DataTo(&visit); err != nil {
			log.Printf("❌ Error unmarshaling visit: %v", err)
			continue
		}
		visits = append(visits, visit)
	}

	log.Printf("✅ Retrieved %d visits for user %s", len(visits), userID)
	return visits, nil
}

// GetActiveVisitsByUserID retrieves active (upcoming) visits for a user
func (vs *VisitService) GetActiveVisitsByUserID(ctx context.Context, userID string) ([]models.Visit, error) {
	// Get all visits for the user (simple query to avoid index issues)
	iter := vs.client.Collection("visits").
		Where("userId", "==", userID).
		Documents(ctx)
	defer iter.Stop()

	var visits []models.Visit
	now := time.Now()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("❌ Error iterating active visits for user %s: %v", userID, err)
			return nil, fmt.Errorf("failed to get active visits: %v", err)
		}

		var visit models.Visit
		if err := doc.DataTo(&visit); err != nil {
			log.Printf("❌ Error unmarshaling visit: %v", err)
			continue
		}

		// Only include visits that are in the future and not manually marked as completed
		if visit.VisitDate.After(now) && !visit.IsCompleted {
			visits = append(visits, visit)
		} else if visit.VisitDate.Before(now) && !visit.IsCompleted {
			// Auto-complete expired visits
			vs.UpdateVisit(ctx, visit.ID, &models.UpdateVisitRequest{
				IsCompleted: true,
				Status:      "completed",
			})
		}
	}

	// Sort by visit date ascending
	for i := 0; i < len(visits)-1; i++ {
		for j := i + 1; j < len(visits); j++ {
			if visits[i].VisitDate.After(visits[j].VisitDate) {
				visits[i], visits[j] = visits[j], visits[i]
			}
		}
	}

	return visits, nil
}

// GetCompletedVisitsByUserID retrieves completed visits for a user
func (vs *VisitService) GetCompletedVisitsByUserID(ctx context.Context, userID string) ([]models.Visit, error) {
	// Get all visits for the user (simple query to avoid index issues)
	iter := vs.client.Collection("visits").
		Where("userId", "==", userID).
		Documents(ctx)
	defer iter.Stop()

	var visits []models.Visit
	now := time.Now()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("❌ Error iterating completed visits for user %s: %v", userID, err)
			return nil, fmt.Errorf("failed to get completed visits: %v", err)
		}

		var visit models.Visit
		if err := doc.DataTo(&visit); err != nil {
			log.Printf("❌ Error unmarshaling visit: %v", err)
			continue
		}

		// Include visits that are manually completed OR have expired (past visit time)
		if visit.IsCompleted || visit.VisitDate.Before(now) {
			// Auto-complete expired visits if not already completed
			if visit.VisitDate.Before(now) && !visit.IsCompleted {
				vs.UpdateVisit(ctx, visit.ID, &models.UpdateVisitRequest{
					IsCompleted: true,
					Status:      "completed",
				})
				visit.IsCompleted = true
				visit.Status = "completed"
			}
			visits = append(visits, visit)
		}
	}

	// Sort by visit date descending (most recent first)
	for i := 0; i < len(visits)-1; i++ {
		for j := i + 1; j < len(visits); j++ {
			if visits[i].VisitDate.Before(visits[j].VisitDate) {
				visits[i], visits[j] = visits[j], visits[i]
			}
		}
	}

	return visits, nil
}

// UpdateVisit updates an existing visit
func (vs *VisitService) UpdateVisit(ctx context.Context, visitID string, updateReq *models.UpdateVisitRequest) error {
	docRef := vs.client.Collection("visits").Doc(visitID)

	// Create update map with only non-empty fields
	updates := make(map[string]interface{})
	updates["updatedAt"] = time.Now()

	if updateReq.Title != "" {
		updates["title"] = updateReq.Title
	}
	if updateReq.Description != "" {
		updates["description"] = updateReq.Description
	}
	if updateReq.VisitDate != "" {
		if visitDate, err := time.Parse(time.RFC3339, updateReq.VisitDate); err == nil {
			updates["visitDate"] = visitDate
		}
	}
	if updateReq.ReminderType != "" {
		updates["reminderType"] = updateReq.ReminderType
	}
	if updateReq.ReminderTime > 0 {
		updates["reminderTime"] = updateReq.ReminderTime
	}
	if updateReq.Status != "" {
		updates["status"] = updateReq.Status
	}
	// Handle boolean field explicitly
	updates["isCompleted"] = updateReq.IsCompleted

	_, err := docRef.Set(ctx, updates, firestore.MergeAll)
	if err != nil {
		log.Printf("❌ Error updating visit %s: %v", visitID, err)
		return fmt.Errorf("failed to update visit: %v", err)
	}

	log.Printf("✅ Visit updated successfully: %s", visitID)
	return nil
}

// DeleteVisit deletes a visit from Firestore
func (vs *VisitService) DeleteVisit(ctx context.Context, visitID string) error {
	_, err := vs.client.Collection("visits").Doc(visitID).Delete(ctx)
	if err != nil {
		log.Printf("❌ Error deleting visit %s: %v", visitID, err)
		return fmt.Errorf("failed to delete visit: %v", err)
	}

	log.Printf("✅ Visit deleted successfully: %s", visitID)
	return nil
}

// GetPendingReminders gets visits that need reminders sent
func (vs *VisitService) GetPendingReminders() ([]models.Visit, error) {
	ctx := context.Background()

	// Get visits that are scheduled, not completed, and haven't been notified yet
	iter := vs.client.Collection("visits").
		Where("status", "==", "scheduled").
		Where("isCompleted", "==", false).
		Where("notifiedAt", "==", time.Time{}).
		Documents(ctx)
	defer iter.Stop()

	var visits []models.Visit
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("❌ Error getting pending reminders: %v", err)
			return nil, err
		}

		var visit models.Visit
		if err := doc.DataTo(&visit); err != nil {
			log.Printf("❌ Error unmarshaling visit: %v", err)
			continue
		}

		// Check if reminder time has passed
		reminderTime := visit.VisitDate.Add(-time.Duration(visit.ReminderTime) * time.Minute)
		if time.Now().After(reminderTime) {
			visits = append(visits, visit)
		}
	}

	return visits, nil
}

// MarkAsNotified marks a visit as notified
func (vs *VisitService) MarkAsNotified(ctx context.Context, visitID string) error {
	_, err := vs.client.Collection("visits").Doc(visitID).Update(ctx, []firestore.Update{
		{
			Path:  "notifiedAt",
			Value: time.Now(),
		},
		{
			Path:  "updatedAt",
			Value: time.Now(),
		},
	})
	if err != nil {
		log.Printf("❌ Error marking visit as notified %s: %v", visitID, err)
		return fmt.Errorf("failed to mark visit as notified: %v", err)
	}

	return nil
}
