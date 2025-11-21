package models

import (
	"time"
)

// Review represents a review in the database
type Review struct {
	ID            string `firestore:"id" json:"id"`
	PropertyID    string `firestore:"propertyId" json:"propertyId"`
	PropertyTitle string `firestore:"propertyTitle" json:"propertyTitle"`
	ReviewerID    string `firestore:"reviewerId" json:"reviewerId"` // User who submitted the review
	ReviewerName  string `firestore:"reviewerName" json:"reviewerName"`
	ReviewerEmail string `firestore:"reviewerEmail" json:"reviewerEmail"`
	ReviewerPhone string `firestore:"reviewerPhone" json:"reviewerPhone"`
	ReviewerType  string `firestore:"reviewerType" json:"reviewerType"` // "tenant" or "owner"
	OwnerID       string `firestore:"ownerId" json:"ownerId"`
	OwnerName     string `firestore:"ownerName" json:"ownerName"`
	OwnerEmail    string `firestore:"ownerEmail" json:"ownerEmail"`
	OwnerPhone    string `firestore:"ownerPhone" json:"ownerPhone"`
	TenantID      string `firestore:"tenantId,omitempty" json:"tenantId,omitempty"`
	TenantName    string `firestore:"tenantName,omitempty" json:"tenantName,omitempty"`
	// Tenant part (reviewing owner)
	TenantPart TenantReviewPart `firestore:"tenantPart" json:"tenantPart"`
	// Owner part (reviewing tenant)
	OwnerPart OwnerReviewPart `firestore:"ownerPart" json:"ownerPart"`
	CreatedAt time.Time       `firestore:"createdAt" json:"createdAt"`
	UpdatedAt time.Time       `firestore:"updatedAt" json:"updatedAt"`
}

// TenantReviewPart represents the tenant's review of the owner
type TenantReviewPart struct {
	OwnerUnderstandable string `firestore:"ownerUnderstandable" json:"ownerUnderstandable"` // Rating/comment
	SoftNature          string `firestore:"softNature" json:"softNature"`
	OwnerTransparent    string `firestore:"ownerTransparent" json:"ownerTransparent"`
	ProblemSolver       string `firestore:"problemSolver" json:"problemSolver"`
	EasyOnRefundMoney   string `firestore:"easyOnRefundMoney" json:"easyOnRefundMoney"`
	OverallExperience   string `firestore:"overallExperience" json:"overallExperience"`
}

// OwnerReviewPart represents the owner's review of the tenant
type OwnerReviewPart struct {
	TenantUnderstandable string `firestore:"tenantUnderstandable" json:"tenantUnderstandable"` // Rating/comment
	SoftNature           string `firestore:"softNature" json:"softNature"`
	TenantTransparent    string `firestore:"tenantTransparent" json:"tenantTransparent"`
	ProblemSolver        string `firestore:"problemSolver" json:"problemSolver"`
	PunctualOnPayment    string `firestore:"punctualOnPayment" json:"punctualOnPayment"`
	OverallExperience    string `firestore:"overallExperience" json:"overallExperience"`
}

// CreateReviewRequest represents the request body for creating a review
type CreateReviewRequest struct {
	PropertyID   string           `json:"propertyId" validate:"required"`
	ReviewerType string           `json:"reviewerType" validate:"required,oneof=tenant owner"`
	TenantPart   TenantReviewPart `json:"tenantPart"`
	OwnerPart    OwnerReviewPart  `json:"ownerPart"`
}
