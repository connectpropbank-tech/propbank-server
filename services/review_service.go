package services

import (
	"context"
	"errors"
	"fmt"
	"shoprop-backend/models"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/google/uuid"
	"google.golang.org/api/iterator"
)

type ReviewService struct {
	client *firestore.Client
}

func NewReviewService(client *firestore.Client) *ReviewService {
	return &ReviewService{
		client: client,
	}
}

// CreateReview creates a new review
func (s *ReviewService) CreateReview(ctx context.Context, req *models.CreateReviewRequest, reviewerID, reviewerName, reviewerEmail, reviewerPhone, propertyTitle, ownerID, ownerName, ownerEmail, ownerPhone, tenantID, tenantName string) (*models.Review, error) {
	// Generate ID
	id := uuid.New().String()

	now := time.Now()
	review := &models.Review{
		ID:            id,
		PropertyID:    req.PropertyID,
		PropertyTitle: propertyTitle,
		ReviewerID:    reviewerID,
		ReviewerName:  reviewerName,
		ReviewerEmail: reviewerEmail,
		ReviewerPhone: reviewerPhone,
		ReviewerType:  req.ReviewerType,
		OwnerID:       ownerID,
		OwnerName:     ownerName,
		OwnerEmail:    ownerEmail,
		OwnerPhone:    ownerPhone,
		TenantID:      tenantID,
		TenantName:    tenantName,
		TenantPart:    req.TenantPart,
		OwnerPart:     req.OwnerPart,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Save to Firestore
	_, err := s.client.Collection("reviews").Doc(id).Set(ctx, review)
	if err != nil {
		return nil, fmt.Errorf("failed to create review: %w", err)
	}

	return review, nil
}

// GetReviewByID retrieves a single review by ID
func (s *ReviewService) GetReviewByID(ctx context.Context, reviewID string) (*models.Review, error) {
	doc, err := s.client.Collection("reviews").Doc(reviewID).Get(ctx)
	if err != nil {
		if err == iterator.Done {
			return nil, errors.New("review not found")
		}
		return nil, fmt.Errorf("failed to get review: %w", err)
	}

	var review models.Review
	if err := doc.DataTo(&review); err != nil {
		return nil, fmt.Errorf("failed to parse review: %w", err)
	}

	return &review, nil
}

// GetReviewsByPropertyID retrieves all reviews for a property
func (s *ReviewService) GetReviewsByPropertyID(ctx context.Context, propertyID string) ([]*models.Review, error) {
	iter := s.client.Collection("reviews").
		Where("propertyId", "==", propertyID).
		OrderBy("createdAt", firestore.Desc).
		Documents(ctx)

	var reviews []*models.Review
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate reviews: %w", err)
		}

		var review models.Review
		if err := doc.DataTo(&review); err != nil {
			return nil, fmt.Errorf("failed to parse review: %w", err)
		}

		reviews = append(reviews, &review)
	}

	return reviews, nil
}
