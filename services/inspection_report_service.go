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

type InspectionReportService struct {
	client *firestore.Client
}

func NewInspectionReportService(client *firestore.Client) *InspectionReportService {
	return &InspectionReportService{
		client: client,
	}
}

// CreateInspectionReport creates a new inspection report
func (s *InspectionReportService) CreateInspectionReport(ctx context.Context, req *models.CreateInspectionReportRequest, userID, userName, userEmail, userPhone, propertyTitle string) (*models.InspectionReport, error) {
	// Generate ID
	id := uuid.New().String()

	now := time.Now()
	report := &models.InspectionReport{
		ID:                        id,
		PropertyID:                req.PropertyID,
		PropertyTitle:             propertyTitle,
		UserID:                    userID,
		UserName:                  userName,
		UserEmail:                 userEmail,
		UserPhone:                 userPhone,
		ReportType:                req.ReportType,
		Report:                    req.Report,
		PossessionLetterFile:      req.PossessionLetterFile,
		HandoverLetterFile:        req.HandoverLetterFile,
		KeysDetails:               req.KeysDetails,
		ElectricityBillMeterImage: req.ElectricityBillMeterImage,
		ElectricityBillReceipt:    req.ElectricityBillReceipt,
		ApartmentConditionImage:   req.ApartmentConditionImage,
		MGLBillMeterImage:         req.MGLBillMeterImage,
		MGLBillReceipt:            req.MGLBillReceipt,
		InternetImage:             req.InternetImage,
		InternetReceipt:           req.InternetReceipt,
		OtherDetails:              req.OtherDetails,
		CreatedAt:                 now,
		UpdatedAt:                 now,
	}

	// Save to Firestore
	_, err := s.client.Collection("inspection_reports").Doc(id).Set(ctx, report)
	if err != nil {
		return nil, fmt.Errorf("failed to create inspection report: %w", err)
	}

	return report, nil
}

// GetInspectionReportsByPropertyID retrieves all inspection reports for a property
func (s *InspectionReportService) GetInspectionReportsByPropertyID(ctx context.Context, propertyID string) ([]*models.InspectionReport, error) {
	iter := s.client.Collection("inspection_reports").
		Where("propertyId", "==", propertyID).
		OrderBy("createdAt", firestore.Desc).
		Documents(ctx)

	var reports []*models.InspectionReport
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate inspection reports: %w", err)
		}

		var report models.InspectionReport
		if err := doc.DataTo(&report); err != nil {
			return nil, fmt.Errorf("failed to parse inspection report: %w", err)
		}

		reports = append(reports, &report)
	}

	return reports, nil
}

// GetInspectionReportsByUserID retrieves all inspection reports for a user
func (s *InspectionReportService) GetInspectionReportsByUserID(ctx context.Context, userID string) ([]*models.InspectionReport, error) {
	iter := s.client.Collection("inspection_reports").
		Where("userId", "==", userID).
		OrderBy("createdAt", firestore.Desc).
		Documents(ctx)

	var reports []*models.InspectionReport
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate inspection reports: %w", err)
		}

		var report models.InspectionReport
		if err := doc.DataTo(&report); err != nil {
			return nil, fmt.Errorf("failed to parse inspection report: %w", err)
		}

		reports = append(reports, &report)
	}

	return reports, nil
}

// GetInspectionReportByID retrieves a single inspection report by ID
func (s *InspectionReportService) GetInspectionReportByID(ctx context.Context, reportID string) (*models.InspectionReport, error) {
	doc, err := s.client.Collection("inspection_reports").Doc(reportID).Get(ctx)
	if err != nil {
		if err == iterator.Done {
			return nil, errors.New("inspection report not found")
		}
		return nil, fmt.Errorf("failed to get inspection report: %w", err)
	}

	var report models.InspectionReport
	if err := doc.DataTo(&report); err != nil {
		return nil, fmt.Errorf("failed to parse inspection report: %w", err)
	}

	return &report, nil
}
