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

type DocumentService struct {
	client *firestore.Client
}

func NewDocumentService(client *firestore.Client) *DocumentService {
	return &DocumentService{
		client: client,
	}
}

// CreateDocument creates a new document
func (s *DocumentService) CreateDocument(ctx context.Context, req *models.CreateDocumentRequest, userID, userName, propertyTitle string) (*models.Document, error) {
	// Generate ID
	id := uuid.New().String()

	now := time.Now()
	document := &models.Document{
		ID:            id,
		PropertyID:    req.PropertyID,
		PropertyTitle: propertyTitle,
		UserID:        userID,
		UserName:      userName,
		DocumentName:  req.DocumentName,
		DocumentType:  req.DocumentType,
		FileURL:       req.FileURL,
		Description:   req.Description,
		CreatedAt:     now,
		UpdatedAt:     now,
	}

	// Save to Firestore
	_, err := s.client.Collection("documents").Doc(id).Set(ctx, document)
	if err != nil {
		return nil, fmt.Errorf("failed to create document: %w", err)
	}

	return document, nil
}

// GetDocumentsByPropertyID retrieves all documents for a property
func (s *DocumentService) GetDocumentsByPropertyID(ctx context.Context, propertyID string) ([]*models.Document, error) {
	iter := s.client.Collection("documents").
		Where("propertyId", "==", propertyID).
		Documents(ctx)

	var documents []*models.Document
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate documents: %w", err)
		}

		var document models.Document
		if err := doc.DataTo(&document); err != nil {
			continue // Skip invalid documents instead of failing
		}

		documents = append(documents, &document)
	}

	// Sort by createdAt descending (newest first)
	// Do this in memory to avoid requiring a Firestore index
	for i := 0; i < len(documents)-1; i++ {
		for j := i + 1; j < len(documents); j++ {
			if documents[i].CreatedAt.Before(documents[j].CreatedAt) {
				documents[i], documents[j] = documents[j], documents[i]
			}
		}
	}

	return documents, nil
}

// GetDocumentByID retrieves a single document by ID
func (s *DocumentService) GetDocumentByID(ctx context.Context, documentID string) (*models.Document, error) {
	doc, err := s.client.Collection("documents").Doc(documentID).Get(ctx)
	if err != nil {
		if err == iterator.Done {
			return nil, errors.New("document not found")
		}
		return nil, fmt.Errorf("failed to get document: %w", err)
	}

	var document models.Document
	if err := doc.DataTo(&document); err != nil {
		return nil, fmt.Errorf("failed to parse document: %w", err)
	}

	return &document, nil
}

// DeleteDocument deletes a document
func (s *DocumentService) DeleteDocument(ctx context.Context, documentID string) error {
	_, err := s.client.Collection("documents").Doc(documentID).Delete(ctx)
	if err != nil {
		return fmt.Errorf("failed to delete document: %w", err)
	}
	return nil
}
