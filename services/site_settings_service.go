package services

import (
	"context"
	"fmt"
	"shoprop-backend/models"
	"time"

	"cloud.google.com/go/firestore"
)

type SiteSettingsService struct {
	client *firestore.Client
}

func NewSiteSettingsService(client *firestore.Client) *SiteSettingsService {
	return &SiteSettingsService{
		client: client,
	}
}

const siteSettingsDocID = "main_settings"

// GetSiteSettings retrieves the site settings from Firestore
func (s *SiteSettingsService) GetSiteSettings(ctx context.Context) (*models.SiteSettings, error) {
	doc, err := s.client.Collection("siteSettings").Doc(siteSettingsDocID).Get(ctx)
	if err != nil {
		// If document doesn't exist, return default settings
		return &models.SiteSettings{
			ID:                   siteSettingsDocID,
			Quote:                "Manage your properties and plan visits with ease",
			QuoteAuthor:          "",
			HeroTitle:            "Your Smart Hub for Property Management",
			HeroSubtitle:         "Manage, list your properties and find your dream house— all in one platform",
			AnnouncementText:     "",
			IsAnnouncementActive: false,
			BannerImages:         []string{}, // Empty array for banner images
			UpdatedAt:            time.Now(),
		}, nil
	}

	var settings models.SiteSettings
	if err := doc.DataTo(&settings); err != nil {
		return nil, fmt.Errorf("failed to parse site settings: %v", err)
	}

	settings.ID = doc.Ref.ID
	return &settings, nil
}

// UpdateSiteSettings updates the site settings in Firestore
func (s *SiteSettingsService) UpdateSiteSettings(ctx context.Context, settings models.SiteSettings) (*models.SiteSettings, error) {
	settings.ID = siteSettingsDocID
	settings.UpdatedAt = time.Now()

	_, err := s.client.Collection("siteSettings").Doc(siteSettingsDocID).Set(ctx, settings)
	if err != nil {
		return nil, fmt.Errorf("failed to update site settings: %v", err)
	}

	return &settings, nil
}

// UpdateQuote updates just the quote field
func (s *SiteSettingsService) UpdateQuote(ctx context.Context, quote string, updatedBy string) (*models.SiteSettings, error) {
	updates := []firestore.Update{
		{Path: "quote", Value: quote},
		{Path: "updatedAt", Value: time.Now()},
		{Path: "updatedBy", Value: updatedBy},
	}

	_, err := s.client.Collection("siteSettings").Doc(siteSettingsDocID).Update(ctx, updates)
	if err != nil {
		// If document doesn't exist, create it
		settings := models.SiteSettings{
			ID:        siteSettingsDocID,
			Quote:     quote,
			UpdatedAt: time.Now(),
			UpdatedBy: updatedBy,
		}
		return s.UpdateSiteSettings(ctx, settings)
	}

	return s.GetSiteSettings(ctx)
}
