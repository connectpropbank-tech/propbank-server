package models

import (
	"time"
)

// Document represents a document attached to a property
type Document struct {
	ID            string    `firestore:"id" json:"id"`
	PropertyID    string    `firestore:"propertyId" json:"propertyId"`
	PropertyTitle string    `firestore:"propertyTitle" json:"propertyTitle"`
	UserID        string    `firestore:"userId" json:"userId"`
	UserName      string    `firestore:"userName" json:"userName"`
	DocumentName  string    `firestore:"documentName" json:"documentName"`
	DocumentType  string    `firestore:"documentType" json:"documentType"` // e.g., "agreement", "receipt", "other"
	FileURL       string    `firestore:"fileUrl" json:"fileUrl"`           // Base64 or file URL
	Description   string    `firestore:"description,omitempty" json:"description,omitempty"`
	CreatedAt     time.Time `firestore:"createdAt" json:"createdAt"`
	UpdatedAt     time.Time `firestore:"updatedAt" json:"updatedAt"`
}

// CreateDocumentRequest represents the request body for creating a document
type CreateDocumentRequest struct {
	PropertyID   string `json:"propertyId" validate:"required"`
	DocumentName string `json:"documentName" validate:"required"`
	DocumentType string `json:"documentType" validate:"required"`
	FileURL      string `json:"fileUrl" validate:"required"`
	Description  string `json:"description,omitempty"`
}
