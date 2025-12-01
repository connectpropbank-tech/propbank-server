package handlers

import (
	"encoding/json"
	"log"
	"net/http"
	"shoprop-backend/models"
	"shoprop-backend/services"

	"cloud.google.com/go/firestore"
)

type SiteSettingsHandler struct {
	siteSettingsService *services.SiteSettingsService
}

func NewSiteSettingsHandler(client *firestore.Client) *SiteSettingsHandler {
	return &SiteSettingsHandler{
		siteSettingsService: services.NewSiteSettingsService(client),
	}
}

// GetSiteSettings returns the current site settings
func (h *SiteSettingsHandler) GetSiteSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	settings, err := h.siteSettingsService.GetSiteSettings(r.Context())
	if err != nil {
		log.Printf("❌ Failed to get site settings: %v", err)
		http.Error(w, "Failed to get site settings", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"settings": settings,
	})
}

// UpdateSiteSettings updates the site settings (admin only)
func (h *SiteSettingsHandler) UpdateSiteSettings(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var settings models.SiteSettings
	if err := json.NewDecoder(r.Body).Decode(&settings); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get admin user ID from header (for audit trail)
	adminUID := r.Header.Get("X-User-ID")
	settings.UpdatedBy = adminUID

	updatedSettings, err := h.siteSettingsService.UpdateSiteSettings(r.Context(), settings)
	if err != nil {
		log.Printf("❌ Failed to update site settings: %v", err)
		http.Error(w, "Failed to update site settings", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Site settings updated by admin: %s", adminUID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Site settings updated successfully",
		"settings": updatedSettings,
	})
}

// UpdateQuote updates just the quote field (admin only)
func (h *SiteSettingsHandler) UpdateQuote(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req struct {
		Quote string `json:"quote"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	if req.Quote == "" {
		http.Error(w, "Quote is required", http.StatusBadRequest)
		return
	}

	// Get admin user ID from header
	adminUID := r.Header.Get("X-User-ID")

	updatedSettings, err := h.siteSettingsService.UpdateQuote(r.Context(), req.Quote, adminUID)
	if err != nil {
		log.Printf("❌ Failed to update quote: %v", err)
		http.Error(w, "Failed to update quote", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Quote updated by admin: %s", adminUID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Quote updated successfully",
		"settings": updatedSettings,
	})
}
