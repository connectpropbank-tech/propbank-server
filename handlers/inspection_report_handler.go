package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"shoprop-backend/models"
	"shoprop-backend/services"

	"cloud.google.com/go/firestore"
)

type InspectionReportHandler struct {
	service                  *services.InspectionReportService
	propertyService          *services.PropertyService
	userService              *services.UserService
	adminNotificationService *services.AdminNotificationService
}

func NewInspectionReportHandler(client *firestore.Client) *InspectionReportHandler {
	return &InspectionReportHandler{
		service:                  services.NewInspectionReportService(client),
		propertyService:          services.NewPropertyService(client),
		userService:              services.NewUserService(client),
		adminNotificationService: services.NewAdminNotificationService(client),
	}
}

// CreateInspectionReport handles POST /inspection-reports
func (h *InspectionReportHandler) CreateInspectionReport(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Get user ID from header or query parameter
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusUnauthorized)
		return
	}

	var req models.CreateInspectionReportRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Validate report type
	if req.ReportType != "on_possession" && req.ReportType != "on_handover" {
		http.Error(w, "Invalid report type. Must be 'on_possession' or 'on_handover'", http.StatusBadRequest)
		return
	}

	// Get user details
	user, err := h.userService.GetUserByID(ctx, userID)
	if err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	// Get property details
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Create inspection report
	report, err := h.service.CreateInspectionReport(
		ctx,
		&req,
		userID,
		user.Name,
		user.Email,
		user.PhoneNumber,
		property.Title,
	)
	if err != nil {
		http.Error(w, "Failed to create inspection report", http.StatusInternalServerError)
		return
	}

	// Get tenant details if any
	var tenantName, tenantEmail, tenantPhone string
	for _, tenant := range property.Tenants {
		if tenant.IsActive {
			tenantName = fmt.Sprintf("%s %s", tenant.FirstName, tenant.LastName)
			tenantEmail = tenant.Email
			tenantPhone = tenant.Phone
			break
		}
	}

	// Get buyers if any
	var buyersString string
	if len(property.Buyers) > 0 {
		buyersJSON, _ := json.Marshal(property.Buyers)
		buyersString = string(buyersJSON)
	}

	// Create admin notification
	notificationReq := models.CreateAdminNotificationRequest{
		Type:  "inspection_report",
		Title: "New Inspection Report Submitted",
		Message: strings.Join([]string{
			user.Name + " submitted an inspection report (" + req.ReportType + ") for property: " + property.Title,
			"Property ID: " + req.PropertyID,
			"Report Type: " + req.ReportType,
		}, "\n"),
		PropertyID:          req.PropertyID,
		OwnerID:             property.OwnerUID,
		OwnerName:           property.OwnerName,
		OwnerPhone:          property.PrimaryNo,
		OwnerEmail:          property.OwnerEmail,
		UserID:              userID,
		UserName:            user.Name,
		UserEmail:           user.Email,
		UserPhone:           user.PhoneNumber,
		PropertyTitle:       property.Title,
		PropertyAddress:     property.Address,
		PropertyListingType: property.ListingType,
		TenantName:          tenantName,
		TenantEmail:         tenantEmail,
		TenantPhone:         tenantPhone,
		Buyers:              buyersString,
		Timestamp:           time.Now().Format(time.RFC3339),
		IsRead:              false,
		Priority:            "medium",
	}

	_, err = h.adminNotificationService.CreateNotification(ctx, notificationReq)
	if err != nil {
		// Don't fail the request if notification fails
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Inspection report created successfully",
		"report":  report,
	})
}

// GetInspectionReportsByProperty handles GET /inspection-reports/property/:propertyId
func (h *InspectionReportHandler) GetInspectionReportsByProperty(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract property ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/inspection-reports/property/")
	propertyID := strings.TrimSuffix(path, "/")

	if propertyID == "" {
		http.Error(w, "Property ID is required", http.StatusBadRequest)
		return
	}

	reports, err := h.service.GetInspectionReportsByPropertyID(ctx, propertyID)
	if err != nil {
		http.Error(w, "Failed to fetch inspection reports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"reports": reports,
	})
}

// GetInspectionReportsByUser handles GET /inspection-reports/user/:userId
func (h *InspectionReportHandler) GetInspectionReportsByUser(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract user ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/inspection-reports/user/")
	userID := strings.TrimSuffix(path, "/")

	if userID == "" {
		http.Error(w, "User ID is required", http.StatusBadRequest)
		return
	}

	reports, err := h.service.GetInspectionReportsByUserID(ctx, userID)
	if err != nil {
		http.Error(w, "Failed to fetch inspection reports", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"reports": reports,
	})
}

// GetInspectionReportByID handles GET /inspection-reports/:reportId
func (h *InspectionReportHandler) GetInspectionReportByID(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	// Extract report ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/inspection-reports/")
	reportID := strings.TrimSuffix(path, "/")

	if reportID == "" {
		http.Error(w, "Report ID is required", http.StatusBadRequest)
		return
	}

	report, err := h.service.GetInspectionReportByID(ctx, reportID)
	if err != nil {
		http.Error(w, "Inspection report not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"report":  report,
	})
}
