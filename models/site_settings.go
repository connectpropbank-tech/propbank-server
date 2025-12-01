package models

import "time"

// SiteSettings stores dynamic content that admin can update
type SiteSettings struct {
	ID                   string    `json:"id" firestore:"id"`
	Quote                string    `json:"quote" firestore:"quote"`
	QuoteAuthor          string    `json:"quoteAuthor,omitempty" firestore:"quoteAuthor,omitempty"`
	HeroTitle            string    `json:"heroTitle,omitempty" firestore:"heroTitle,omitempty"`
	HeroSubtitle         string    `json:"heroSubtitle,omitempty" firestore:"heroSubtitle,omitempty"`
	AnnouncementText     string    `json:"announcementText,omitempty" firestore:"announcementText,omitempty"`
	IsAnnouncementActive bool      `json:"isAnnouncementActive" firestore:"isAnnouncementActive"`
	UpdatedAt            time.Time `json:"updatedAt" firestore:"updatedAt"`
	UpdatedBy            string    `json:"updatedBy,omitempty" firestore:"updatedBy,omitempty"`
}

// SiteSettingsResponse is the response format for site settings
type SiteSettingsResponse struct {
	Quote                string `json:"quote"`
	QuoteAuthor          string `json:"quoteAuthor,omitempty"`
	HeroTitle            string `json:"heroTitle,omitempty"`
	HeroSubtitle         string `json:"heroSubtitle,omitempty"`
	AnnouncementText     string `json:"announcementText,omitempty"`
	IsAnnouncementActive bool   `json:"isAnnouncementActive"`
}
