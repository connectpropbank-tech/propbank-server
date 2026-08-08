package handlers

import (
	"context"
	"encoding/json"
	"fmt"
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
	emailService             *services.EmailService
}

func NewAgreementHandler(client *firestore.Client) *AgreementHandler {
	return &AgreementHandler{
		propertyService:          services.NewPropertyService(client),
		userService:              services.NewUserService(client),
		adminNotificationService: services.NewAdminNotificationService(client),
		emailService:             services.NewEmailService(),
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
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get property
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
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
		PropertyID      string `json:"propertyId" validate:"required"`
		NoticePeriod    string `json:"noticePeriod"`    // "1 Month", "2 Months", "3 Months", "Immediate"
		NoticeStartDate string `json:"noticeStartDate"` // Added
		TerminationDate string `json:"terminationDate"` // Calculated date
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get property BEFORE updating (to capture tenant details for notification)
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Verify ownership
	if property.OwnerUID != userID {
		http.Error(w, "Unauthorized: You can only terminate your own property agreements", http.StatusForbidden)
		return
	}

	// Capture tenant details...
	tenantName := property.TenantName
	tenantPhone := property.MobileNumber
	tenantEmail := property.TenantEmail

	// Also check Tenants array for active tenant to get email if missing from top level
	if len(property.Tenants) > 0 {
		for _, tenant := range property.Tenants {
			if tenant.IsActive {
				tenantName = tenant.FirstName + " " + tenant.LastName
				if tenantEmail == "" {
					tenantEmail = tenant.Email
				}
				tenantPhone = tenant.Phone
			}
		}
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

	// CHECK: Is this a Notice Period or Immediate Termination?
	isNotice := req.NoticePeriod != ""

	var updateData map[string]interface{}
	var successMessage string

	if isNotice {
		// --- NOTICE PERIOD FLOW ---
		// Do NOT delete tenant data. Just mark status and set date.
		updateData = map[string]interface{}{
			"agreementStatus":            "notice_served",
			"anticipatedTerminationDate": req.TerminationDate,
			"updatedAt":                  time.Now(),
		}
		successMessage = "Termination notice served successfully. Tenant has been notified."
	} else {
		// --- IMMEDIATE TERMINATION FLOW ---
		// Existing logic: Clear all data
		updateData = map[string]interface{}{
			"agreementStatus":            "terminated",
			"rentalStatus":               "available",
			"status":                     "active",
			"anticipatedTerminationDate": "",

			// Clear tenant information
			"tenantName":   "",
			"personName":   "",
			"mobileNumber": "",
			"primaryNo":    "",
			"ultNo":        "",
			"tenants":      []interface{}{},

			// Clear agreement dates
			"agreementStartDate": "",
			"agreementEndDate":   "",
			"agreementPeriod":    "",

			// Clear lease-related fields
			"noticePeriod": "",
			"lockInPeriod": "",

			"updatedAt": time.Now(),
		}
		successMessage = "Agreement terminated successfully. Tenant information has been removed."
	}

	updatedProperty, err := h.propertyService.UpdateProperty(ctx, req.PropertyID, updateData)
	if err != nil {
		http.Error(w, "Failed to update property", http.StatusInternalServerError)
		return
	}

	// --- NOTIFICATIONS ---

	if isNotice {
		// Send NOTICE email
		go func() {
			emailData := services.AgreementTerminationEmailData{
				TenantName:      tenantName,
				TenantEmail:     tenantEmail,
				OwnerName:       ownerName,
				OwnerEmail:      ownerEmail,
				PropertyTitle:   property.Title,
				PropertyAddress: property.Location,
				PropertyType:    property.PropertyType,
				UnitNumber:      property.UnitNumber,
				Floor:           property.Floor,
				Configuration:   property.Configuration,
				CarpetArea:      property.CarpetArea,
				MonthlyRent:     property.MonthlyRent,
				TerminationDate: req.TerminationDate,
				RaisedBy:        "Owner", // The notice is served by owner
			}
			if err := h.emailService.SendAgreementNoticeNotification(emailData); err != nil {
				// Log error
			}
		}()
	} else {
		// Send email notification to both owner and tenant
		go func() {
			emailData := services.AgreementTerminationEmailData{
				TenantName:         tenantName,
				TenantEmail:        tenantEmail,
				TenantPhone:        tenantPhone,
				OwnerName:          ownerName,
				OwnerEmail:         ownerEmail,
				OwnerPhone:         ownerPhone,
				PropertyTitle:      property.Title,
				PropertyAddress:    property.Location,
				PropertyType:       property.PropertyType,
				UnitNumber:         property.UnitNumber,
				Floor:              property.Floor,
				Configuration:      property.Configuration,
				CarpetArea:         property.CarpetArea,
				TerminationDate:    services.FormatTimeIST(time.Now()),
				Reason:             "Agreement terminated by property owner",
				AgreementStartDate: property.AgreementStartDate,
				AgreementEndDate:   property.AgreementEndDate,
				AgreementPeriod:    property.AgreementPeriod,
				MonthlyRent:        property.MonthlyRent,
				RaisedBy:           "Owner",
			}

			if err := h.emailService.SendAgreementTerminationNotification(emailData); err != nil {
			}
		}()
	}

	// Create admin notification when owner terminates or serves notice
	notificationType := "agreement_termination"
	var notificationTitle string
	var notificationMessage string

	if isNotice {
		notificationTitle = "Notice Served by Owner"
		notificationMessage = fmt.Sprintf("Owner %s has served a %s termination notice for property %s. Notice Starts From: %s. Anticipated Termination Date: %s. Tenant: %s.", ownerName, req.NoticePeriod, property.Title, req.NoticeStartDate, req.TerminationDate, tenantName)
	} else {
		notificationTitle = "Agreement Terminated Immediately"
		notificationMessage = fmt.Sprintf("Owner %s has immediately terminated the agreement for property %s. Tenant: %s.", ownerName, property.Title, tenantName)
	}

	notificationReq := models.CreateAdminNotificationRequest{
		Type:                notificationType,
		Title:               notificationTitle,
		Message:             notificationMessage,
		PropertyID:          req.PropertyID,
		OwnerID:             property.OwnerUID,
		OwnerName:           ownerName,
		OwnerEmail:          ownerEmail,
		OwnerPhone:          ownerPhone,
		OwnerRole:           property.OwnerRole,
		UserName:            tenantName,
		UserEmail:           tenantEmail,
		UserPhone:           tenantPhone,
		TenantName:          tenantName,
		TenantPhone:         tenantPhone,
		TenantEmail:         tenantEmail,
		PropertyTitle:       property.Title,
		PropertyAddress:     property.Location,
		PropertyListingType: property.ListingType,
		Timestamp:           time.Now().Format(time.RFC3339),
		IsRead:              false,
		Priority:            "high",
	}

	_, _ = h.adminNotificationService.CreateNotification(ctx, notificationReq)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"message":  successMessage,
		"property": updatedProperty,
	})
}

// RequestTermination handles POST /agreements/request-termination
// Allows a tenant to request termination of their agreement
func (h *AgreementHandler) RequestTermination(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from header (Tenant's ID)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusUnauthorized)
		return
	}

	var req struct {
		PropertyID      string `json:"propertyId" validate:"required"`
		NoticePeriod    string `json:"noticePeriod"`    // Added noticePeriod
		NoticeStartDate string `json:"noticeStartDate"` // Added
		TerminationDate string `json:"terminationDate"` // Calculated date
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get property
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Verify Tenant
	isTenant := false
	var tenantName string = "Tenant"
	var tenantEmail string = ""
	var tenantPhone string = ""

	// Check TenantEmail on property struct (now that we added it)
	user, err := h.userService.GetUserByID(ctx, userID)
	if err == nil && user != nil {
		if property.TenantEmail != "" && property.TenantEmail == user.Email {
			isTenant = true
			tenantName = property.TenantName
			tenantEmail = property.TenantEmail
			tenantPhone = property.MobileNumber
		} else if property.MobileNumber != "" && property.MobileNumber == user.PhoneNumber {
			isTenant = true
			tenantName = property.TenantName
			if property.TenantEmail != "" {
				tenantEmail = property.TenantEmail
			} else {
				tenantEmail = user.Email // Fallback
			}
			tenantPhone = property.MobileNumber
		} else {
			// Check tenants array
			for _, t := range property.Tenants {
				if t.IsActive && ((t.Email != "" && t.Email == user.Email) || (t.Phone != "" && t.Phone == user.PhoneNumber)) {
					isTenant = true
					tenantName = t.FirstName + " " + t.LastName
					tenantEmail = t.Email
					tenantPhone = t.Phone
					break
				}
			}
		}
	}

	if tenantPhone == "" && user != nil {
		tenantPhone = user.PhoneNumber
	}

	ownerPhone := property.OwnerPhone
	if ownerPhone == "" {
		owner, err := h.userService.GetUserByID(ctx, property.OwnerUID)
		if err == nil && owner != nil {
			ownerPhone = owner.PhoneNumber
		}
	}

	if !isTenant {
		http.Error(w, "Unauthorized: You are not recognized as a tenant of this property", http.StatusForbidden)
		return
	}

	// Send email notification to owner
	go func() {
		if err := h.emailService.SendTerminationRequestNotification(
			property.OwnerEmail,
			property.OwnerName,
			tenantName,
			tenantEmail, // Added tenant email
			property.Title,
			property.ID,
			req.NoticePeriod, // Added notice period
		); err != nil {
			// Log error
		}
	}()

	// Create admin notification
	notificationReq := models.CreateAdminNotificationRequest{
		Type:                "agreement_termination", // Changed from termination_request
		Title:               "Termination Requested by Tenant",
		Message:             fmt.Sprintf("Tenant %s has requested termination for property %s.\nNotice Period: %s\nNotice Starts From: %s\nAnticipated Termination Date: %s", tenantName, property.Title, req.NoticePeriod, req.NoticeStartDate, req.TerminationDate),
		PropertyID:          req.PropertyID,
		OwnerID:             property.OwnerUID,
		OwnerName:           property.OwnerName,
		OwnerEmail:          property.OwnerEmail,
		OwnerPhone:          ownerPhone,
		UserName:            tenantName,
		UserEmail:           tenantEmail,
		UserPhone:           tenantPhone,
		TenantName:          tenantName,
		TenantEmail:         tenantEmail,
		TenantPhone:         tenantPhone,
		PropertyTitle:       property.Title,
		PropertyAddress:     property.Location,
		PropertyListingType: property.ListingType,
		Timestamp:           time.Now().Format(time.RFC3339),
		IsRead:              false,
		Priority:            "medium",
	}
	h.adminNotificationService.CreateNotification(ctx, notificationReq)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Termination request sent to owner successfully.",
	})
}

// RequestRenewal handles POST /agreements/request-renewal
// Allows a tenant to request renewal of their agreement
func (h *AgreementHandler) RequestRenewal(w http.ResponseWriter, r *http.Request) {
	ctx := context.Background()

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get user ID from header (Tenant's ID)
	userID := r.Header.Get("X-User-ID")
	if userID == "" {
		http.Error(w, "User ID is required", http.StatusUnauthorized)
		return
	}

	var req struct {
		PropertyID string `json:"propertyId" validate:"required"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Get property
	property, err := h.propertyService.GetPropertyByID(ctx, req.PropertyID)
	if err != nil {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Verify Tenant
	isTenant := false
	var tenantName string = "Tenant"
	var tenantEmail string = ""
	var tenantPhone string = ""

	user, err := h.userService.GetUserByID(ctx, userID)
	if err == nil && user != nil {
		if property.TenantEmail != "" && property.TenantEmail == user.Email {
			isTenant = true
			tenantName = property.TenantName
			tenantEmail = property.TenantEmail
			tenantPhone = property.MobileNumber
		} else if property.MobileNumber != "" && property.MobileNumber == user.PhoneNumber {
			isTenant = true
			tenantName = property.TenantName
			if property.TenantEmail != "" {
				tenantEmail = property.TenantEmail
			} else {
				tenantEmail = user.Email // Fallback
			}
			tenantPhone = property.MobileNumber
		} else {
			// Check tenants array
			for _, t := range property.Tenants {
				if t.IsActive && ((t.Email != "" && t.Email == user.Email) || (t.Phone != "" && t.Phone == user.PhoneNumber)) {
					isTenant = true
					tenantName = t.FirstName + " " + t.LastName
					tenantEmail = t.Email
					tenantPhone = t.Phone
					break
				}
			}
		}
	}

	if tenantPhone == "" && user != nil {
		tenantPhone = user.PhoneNumber
	}

	ownerPhone := property.OwnerPhone
	if ownerPhone == "" {
		owner, err := h.userService.GetUserByID(ctx, property.OwnerUID)
		if err == nil && owner != nil {
			ownerPhone = owner.PhoneNumber
		}
	}

	if !isTenant {
		http.Error(w, "Unauthorized: You are not recognized as a tenant of this property", http.StatusForbidden)
		return
	}

	// Create admin notification
	notificationReq := models.CreateAdminNotificationRequest{
		Type:                "agreement_renewal",
		Title:               "Renewal Requested by Tenant",
		Message:             fmt.Sprintf("Tenant %s (%s) has requested renewal for property %s.", tenantName, tenantEmail, property.Title),
		PropertyID:          req.PropertyID,
		OwnerID:             property.OwnerUID,
		OwnerName:           property.OwnerName,
		OwnerEmail:          property.OwnerEmail,
		OwnerPhone:          ownerPhone,
		UserName:            tenantName,
		UserEmail:           tenantEmail,
		UserPhone:           tenantPhone,
		TenantName:          tenantName,
		TenantEmail:         tenantEmail,
		TenantPhone:         tenantPhone,
		PropertyTitle:       property.Title,
		PropertyAddress:     property.Location,
		PropertyListingType: property.ListingType,
		Timestamp:           time.Now().Format(time.RFC3339),
		IsRead:              false,
		Priority:            "medium",
	}
	h.adminNotificationService.CreateNotification(ctx, notificationReq)

	// Send email notification to owner and tenant
	if err := h.emailService.SendRenewalRequestNotification(
		property.OwnerEmail,
		property.OwnerName,
		tenantName,
		tenantEmail,
		property.Title,
		req.PropertyID,
	); err != nil {
		// Log error but continue since this is non-critical
		fmt.Printf("Failed to send renewal request email: %v\n", err)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Renewal request sent to owner successfully.",
	})
}
