package handlers

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"shoprop-backend/models"
	"shoprop-backend/services"

	"cloud.google.com/go/firestore"
)

type AgreementHandler struct {
	propertyService          *services.PropertyService
	userService              *services.UserService
	adminNotificationService *services.AdminNotificationService
}

func NewAgreementHandler(client *firestore.Client) *AgreementHandler {
	return &AgreementHandler{
		propertyService:          services.NewPropertyService(client),
		userService:              services.NewUserService(client),
		adminNotificationService: services.NewAdminNotificationService(client),
	}
}

// RenewAgreement handles POST /agreements/renew
func (h *AgreementHandler) RenewAgreement(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusUnauthorized)
		return
	}

	var req struct {
		PropertyID string `json:"propertyId" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get property
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
		log.Printf("Error fetching property: %v", err)
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if property.OwnerUID != userID {
		http.Error(w, "Unauthorized: You can only renew your own property agreements", http.StatusForbidden)
		return
	}

	// Update property with renewed agreement status
	updateData := map[string]interface{}{
		"agreementStatus": "renewed",
		"updatedAt":       time.Now(),
	}

	updatedProperty, err := h.propertyService.UpdateProperty(ctx, req.PropertyID, updateData)
	if err != nil {
		log.Printf("Error updating property: %v", err)
		http.Error(w, "Failed to renew agreement", http.StatusInternalServerError)
		return
	}

	// Get owner details
	owner, err := h.userService.GetUserByID(ctx, property.OwnerUID)
	ownerName := property.OwnerName
	ownerEmail := property.OwnerEmail
	if err == nil && owner != nil {
		ownerName = owner.Name
		ownerEmail = owner.Email
	}

	// Create admin notification
	notificationReq := models.CreateAdminNotificationRequest{
		Type:  "agreement_renewal",
		Title: "Agreement Renewal Request",
		Message: strings.Join([]string{
			ownerName + " has requested to renew the agreement for property: " + property.Title,
			"Property ID: " + req.PropertyID,
			"Owner: " + ownerName + " (" + ownerEmail + ")",
		}, "\n"),
		PropertyID:          req.PropertyID,
		OwnerID:             property.OwnerUID,
		OwnerName:           property.OwnerName,
		OwnerPhone:          property.PrimaryNo,
		OwnerEmail:          property.OwnerEmail,
		PropertyTitle:       property.Title,
		PropertyAddress:     property.Address,
		PropertyListingType: property.ListingType,
		Timestamp:           time.Now().Format(time.RFC3339),
		IsRead:              false,
		Priority:            "high",
	}

	_, err = h.adminNotificationService.CreateNotification(ctx, notificationReq)
	if err != nil {
		log.Printf("Error creating admin notification for agreement renewal: %v", err)
		// Don't fail the request if notification fails
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Agreement renewal request submitted successfully. Admin has been notified.",
		"property": updatedProperty,
	})
}

// TerminateAgreement handles POST /agreements/terminate
// This clears tenant data, marks property as available for rent, and notifies admin
func (h *AgreementHandler) TerminateAgreement(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	log.Printf("✅ TerminateAgreement handler called: Method=%s, Path=%s, URL=%s", r.Method, r.URL.Path, r.URL.String())

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from header
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		userID = r.URL.Query().Get("userId")
	}
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusUnauthorized)
		return
	}

	var req struct {
		PropertyID string `json:"propertyId" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Printf("Error decoding request: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get property BEFORE updating (to capture tenant details for notification)
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
		log.Printf("Error fetching property: %v", err)
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if property.OwnerUID != userID {
		http.Error(w, "Unauthorized: You can only terminate your own property agreements", http.StatusForbidden)
		return
	}

	// Store tenant details before clearing (for notification)
	tenantName := property.TenantName
	tenantPhone := property.MobileNumber
	tenantPersonName := property.PersonName
	tenantUltNo := property.UltNo

	// Also check Tenants array for more details
	var tenantEmail string
	var tenantDetails []string
	if len(property.Tenants) > 0 {
		for _, tenant := range property.Tenants {
			if tenant.IsActive {
				tenantName = tenant.FirstName + " " + tenant.LastName
				tenantEmail = tenant.Email
				tenantPhone = tenant.Phone
				tenantDetails = append(tenantDetails, "- "+tenant.FirstName+" "+tenant.LastName+" ("+tenant.Email+", "+tenant.Phone+")")
			}
		}
	}

	// Update property:
	// 1. Set agreementStatus to "terminated"
	// 2. Set rentalStatus to "available" (NOT archive - keep it active for new tenants)
	// 3. Clear all tenant information
	// 4. Clear agreement dates
	updateData := map[string]interface{}{
		"agreementStatus": "terminated",
		"rentalStatus":    "available", // Mark as available for rent again
		"status":          "active",    // Keep property active (not archived)

		// Clear tenant information
		"tenantName":   "",
		"personName":   "",
		"mobileNumber": "",
		"primaryNo":    "",
		"ultNo":        "",
		"tenants":      []interface{}{}, // Clear tenants array

		// Clear agreement dates (optional - you may want to keep these for history)
		"agreementStartDate": "",
		"agreementEndDate":   "",
		"agreementPeriod":    "",

		// Clear lease-related fields
		"noticePeriod": "",
		"lockInPeriod": "",

		"updatedAt": time.Now(),
	}

	updatedProperty, err := h.propertyService.UpdateProperty(ctx, req.PropertyID, updateData)
	if err != nil {
		log.Printf("Error updating property: %v", err)
		http.Error(w, "Failed to terminate agreement", http.StatusInternalServerError)
		return
	}

	// Get owner details
	owner, err := h.userService.GetUserByID(ctx, property.OwnerUID)
	ownerName := property.OwnerName
	ownerEmail := property.OwnerEmail
	ownerPhone := property.PrimaryNo
	if err == nil && owner != nil {
		ownerName = owner.Name
		ownerEmail = owner.Email
		if owner.PhoneNumber != "" {
			ownerPhone = owner.PhoneNumber
		}
	}

	// Build comprehensive notification message with owner, tenant, and property details
	notificationLines := []string{
		"🔴 AGREEMENT TERMINATED",
		"",
		"📋 PROPERTY DETAILS:",
		"  • Property: " + property.Title,
		"  • Property ID: " + req.PropertyID,
		"  • Type: " + property.PropertyType,
		"  • Address: " + property.Location,
		"  • Monthly Rent: ₹" + property.MonthlyRent,
		"",
		"👤 OWNER DETAILS:",
		"  • Name: " + ownerName,
		"  • Email: " + ownerEmail,
		"  • Phone: " + ownerPhone,
		"",
		"🏠 TENANT DETAILS (Now Removed):",
	}

	if tenantName != "" {
		notificationLines = append(notificationLines, "  • Name: "+tenantName)
	}
	if tenantEmail != "" {
		notificationLines = append(notificationLines, "  • Email: "+tenantEmail)
	}
	if tenantPhone != "" {
		notificationLines = append(notificationLines, "  • Phone: "+tenantPhone)
	}
	if tenantPersonName != "" {
		notificationLines = append(notificationLines, "  • Contact Person: "+tenantPersonName)
	}
	if tenantUltNo != "" {
		notificationLines = append(notificationLines, "  • Alt. Phone: "+tenantUltNo)
	}
	if len(tenantDetails) > 0 {
		notificationLines = append(notificationLines, "  Additional Tenants:")
		notificationLines = append(notificationLines, tenantDetails...)
	}

	notificationLines = append(notificationLines, "", "📅 AGREEMENT INFO:")
	if property.AgreementStartDate != "" {
		notificationLines = append(notificationLines, "  • Start Date: "+property.AgreementStartDate)
	}
	if property.AgreementEndDate != "" {
		notificationLines = append(notificationLines, "  • End Date: "+property.AgreementEndDate)
	}
	if property.AgreementPeriod != "" {
		notificationLines = append(notificationLines, "  • Period: "+property.AgreementPeriod+" months")
	}
	if property.SecurityDeposit != "" {
		notificationLines = append(notificationLines, "  • Security Deposit: ₹"+property.SecurityDeposit)
	}

	notificationLines = append(notificationLines, "", "✅ Property is now marked as 'Available for Rent'")

	// Create admin notification with full details
	notificationReq := models.CreateAdminNotificationRequest{
		Type:                "agreement_termination",
		Title:               "Agreement Terminated - " + property.Title,
		Message:             strings.Join(notificationLines, "\n"),
		PropertyID:          req.PropertyID,
		OwnerID:             property.OwnerUID,
		OwnerName:           ownerName,
		OwnerPhone:          ownerPhone,
		OwnerEmail:          ownerEmail,
		UserName:            tenantName, // Store tenant name in user fields
		UserEmail:           tenantEmail,
		UserPhone:           tenantPhone,
		PropertyTitle:       property.Title,
		PropertyAddress:     property.Location,
		PropertyListingType: property.ListingType,
		Timestamp:           time.Now().Format(time.RFC3339),
		IsRead:              false,
		Priority:            "high",
	}

	_, err = h.adminNotificationService.CreateNotification(ctx, notificationReq)
	if err != nil {
		log.Printf("Error creating admin notification for agreement termination: %v", err)
		// Don't fail the request if notification fails
	}

	log.Printf("✅ Agreement terminated successfully for property %s. Tenant data cleared, property marked as available.", req.PropertyID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  "Agreement terminated successfully. Tenant information has been removed and property is now available for rent. Admin has been notified.",
		"property": updatedProperty,
	})
}
