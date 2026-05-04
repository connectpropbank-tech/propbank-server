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

type TenantHandler struct {
	client       *firestore.Client
	userService  *services.UserService
	emailService *services.EmailService
}

func NewTenantHandler(client *firestore.Client, emailService *services.EmailService) *TenantHandler {
	return &TenantHandler{
		client:       client,
		userService:  services.NewUserService(client),
		emailService: emailService,
	}
}

// CreateTenant creates new tenants and adds them to the property document
func (h *TenantHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Define request structure
	var requestData struct {
		PropertyID string          `json:"propertyId"`
		OwnerUID   string          `json:"ownerUID"`
		Tenants    []models.Tenant `json:"tenants"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if requestData.PropertyID == "" {
		http.Error(w, "Property ID is required", http.StatusBadRequest)
		return
	}

	if requestData.OwnerUID == "" {
		http.Error(w, "Owner UID is required", http.StatusBadRequest)
		return
	}

	if len(requestData.Tenants) == 0 {
		http.Error(w, "At least one tenant is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// First, get the current property to check if it exists
	propertyRef := h.client.Collection("properties").Doc(requestData.PropertyID)
	propertyDoc, err := propertyRef.Get(ctx)
	if err != nil {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Get property details for user mapping
	var property models.Property
	if err := propertyDoc.DataTo(&property); err != nil {
		http.Error(w, "Failed to read property data", http.StatusInternalServerError)
		return
	}

	// Get current tenants array from property (if any)
	var propertyData map[string]interface{}
	if err := propertyDoc.DataTo(&propertyData); err != nil {
		http.Error(w, "Failed to read property data", http.StatusInternalServerError)
		return
	}

	// Get existing tenants or initialize empty array
	var existingTenants []map[string]interface{}
	if tenants, exists := propertyData["tenants"]; exists && tenants != nil {
		if tenantsArray, ok := tenants.([]interface{}); ok {
			for _, tenant := range tenantsArray {
				if tenantMap, ok := tenant.(map[string]interface{}); ok {
					existingTenants = append(existingTenants, tenantMap)
				}
			}
		}
	}

	var newTenantInfos []map[string]interface{}

	// Process each tenant
	for _, tenant := range requestData.Tenants {
		// Validate tenant required fields
		if tenant.FirstName == "" || tenant.LastName == "" || tenant.Email == "" || tenant.Phone == "" {
			http.Error(w, fmt.Sprintf("First name, last name, email, and phone are required for tenant %s %s", tenant.FirstName, tenant.LastName), http.StatusBadRequest)
			return
		}

		// Generate ID and set timestamps
		docRef := h.client.Collection("tenants").NewDoc()
		tenant.ID = docRef.ID
		tenant.PropertyID = requestData.PropertyID
		tenant.OwnerUID = requestData.OwnerUID
		tenant.IsActive = true
		tenant.CreatedAt = time.Now()
		tenant.UpdatedAt = time.Now()

		// Save to separate tenants collection (for detailed management)
		_, err := docRef.Set(ctx, tenant)
		if err != nil {
			http.Error(w, fmt.Sprintf("Failed to create tenant %s %s", tenant.FirstName, tenant.LastName), http.StatusInternalServerError)
			return
		}

		// Convert tenant to map for property storage
		tenantInfo := map[string]interface{}{
			"id":                   tenant.ID,
			"firstName":            tenant.FirstName,
			"lastName":             tenant.LastName,
			"email":                tenant.Email,
			"phone":                tenant.Phone,
			"emergencyContact":     tenant.EmergencyContact,
			"userUID":              tenant.UserUID, // Map to platform user if found
			"isMarried":            tenant.IsMarried,
			"leaseStartDate":       tenant.LeaseStartDate,
			"leaseEndDate":         tenant.LeaseEndDate,
			"monthlyRent":          tenant.MonthlyRent,
			"securityDeposit":      tenant.SecurityDeposit,
			"paymentDueDate":       tenant.PaymentDueDate,
			"escalationPercentage": tenant.EscalationPercentage,
			"escalationAmount":     tenant.EscalationAmount,
			"previousAddress":      tenant.PreviousAddress,
			"employmentStatus":     tenant.EmploymentStatus,
			"employer":             tenant.Employer,
			"monthlyIncome":        tenant.MonthlyIncome,
			"notes":                tenant.Notes,
			"isActive":             tenant.IsActive,
			"createdAt":            tenant.CreatedAt,
			"updatedAt":            tenant.UpdatedAt,
		}

		// Add spouse info if tenant is married
		if tenant.IsMarried && tenant.Spouse != nil {
			tenantInfo["spouse"] = map[string]interface{}{
				"firstName":        tenant.Spouse.FirstName,
				"lastName":         tenant.Spouse.LastName,
				"email":            tenant.Spouse.Email,
				"phone":            tenant.Spouse.Phone,
				"employmentStatus": tenant.Spouse.EmploymentStatus,
				"employer":         tenant.Spouse.Employer,
				"notes":            tenant.Spouse.Notes,
			}
		}

		newTenantInfos = append(newTenantInfos, tenantInfo)

		// Send Welcome Email to Tenant and Notification to Owner
		emailData := services.TenantAddedEmailData{
			TenantName:      tenant.FirstName + " " + tenant.LastName,
			TenantEmail:     tenant.Email,
			TenantPhone:     tenant.Phone,
			OwnerName:       property.OwnerName,
			OwnerEmail:      property.OwnerEmail,
			OwnerPhone:      property.OwnerPhone,
			PropertyTitle:   property.Title,
			PropertyAddress: property.Address,
			PropertyType:    property.PropertyType,
			MonthlyRent:     tenant.MonthlyRent,
			AgreementStart:  tenant.LeaseStartDate,
			AgreementEnd:    tenant.LeaseEndDate,
		}

		// Send notification in background
		go func(data services.TenantAddedEmailData) {
			if err := h.emailService.SendTenantAddedNotification(data); err != nil {
				fmt.Printf("[TenantHandler] Failed to send tenant welcome notification: %v\n", err)
			}
		}(emailData)

		// If tenant has a userUID (platform user), update user with property information
		if tenant.UserUID != "" {
			err := h.updateUserRentedProperty(ctx, tenant.UserUID, requestData.PropertyID, property.OwnerUID, property.OwnerName)
			if err != nil {
				// Don't fail the tenant creation if user update fails
			} else {
			}
		}
	}

	// Update property with new tenants (append to existing)
	allTenants := append(existingTenants, newTenantInfos...)

	_, err = propertyRef.Update(ctx, []firestore.Update{
		{Path: "tenants", Value: allTenants},
		{Path: "updatedAt", Value: time.Now()},
	})
	if err != nil {
		http.Error(w, "Failed to add tenants to property", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": fmt.Sprintf("%d tenant(s) added successfully", len(newTenantInfos)),
		"tenants": newTenantInfos,
	})
}

// GetTenantsByProperty gets all tenants for a specific property
func (h *TenantHandler) GetTenantsByProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	propertyID := r.URL.Query().Get("propertyId")
	if propertyID == "" {
		http.Error(w, "Property ID is required", http.StatusBadRequest)
		return
	}

	ctx := context.Background()

	// Query tenants by property ID
	iter := h.client.Collection("tenants").Where("propertyId", "==", propertyID).Where("isActive", "==", true).Documents(ctx)

	var tenants []models.Tenant
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var tenant models.Tenant
		if err := doc.DataTo(&tenant); err != nil {
			continue
		}

		tenants = append(tenants, tenant)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tenants": tenants,
	})
}

// GetTenant gets a specific tenant by ID
func (h *TenantHandler) GetTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != "GET" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract tenant ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "Tenant ID is required", http.StatusBadRequest)
		return
	}
	tenantID := pathParts[1]

	ctx := context.Background()
	doc, err := h.client.Collection("tenants").Doc(tenantID).Get(ctx)
	if err != nil {
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	var tenant models.Tenant
	if err := doc.DataTo(&tenant); err != nil {
		http.Error(w, "Failed to parse tenant data", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"tenant":  tenant,
	})
}

// UpdateTenant updates an existing tenant
func (h *TenantHandler) UpdateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != "PUT" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract tenant ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "Tenant ID is required", http.StatusBadRequest)
		return
	}
	tenantID := pathParts[1]

	var updates map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updates); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Add update timestamp
	updates["updatedAt"] = time.Now()

	ctx := context.Background()
	_, err := h.client.Collection("tenants").Doc(tenantID).Update(ctx, []firestore.Update{
		{Path: "updatedAt", Value: updates["updatedAt"]},
	})

	// Apply other updates
	for key, value := range updates {
		if key != "updatedAt" && key != "id" && key != "createdAt" {
			_, err = h.client.Collection("tenants").Doc(tenantID).Update(ctx, []firestore.Update{
				{Path: key, Value: value},
			})
			if err != nil {
			}
		}
	}

	if err != nil {
		http.Error(w, "Failed to update tenant", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tenant updated successfully",
	})
}

// DeleteTenant soft deletes a tenant (sets isActive to false)
func (h *TenantHandler) DeleteTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != "DELETE" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract tenant ID from URL path
	pathParts := strings.Split(strings.Trim(r.URL.Path, "/"), "/")
	if len(pathParts) < 2 {
		http.Error(w, "Tenant ID is required", http.StatusBadRequest)
		return
	}
	tenantID := pathParts[1]

	ctx := context.Background()
	_, err := h.client.Collection("tenants").Doc(tenantID).Update(ctx, []firestore.Update{
		{Path: "isActive", Value: false},
		{Path: "updatedAt", Value: time.Now()},
	})

	if err != nil {
		http.Error(w, "Failed to delete tenant", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tenant deleted successfully",
	})
}

// updateUserRentedProperty updates a user's record with rented property information
func (h *TenantHandler) updateUserRentedProperty(ctx context.Context, userUID, propertyID, ownerUID, ownerName string) error {
	// Get existing user
	user, err := h.userService.GetUserByID(ctx, userUID)
	if err != nil {
		return fmt.Errorf("failed to get user: %v", err)
	}

	// Update user with rented property information
	user.RentedPropertyID = propertyID
	user.RentedPropertyOwnerID = ownerUID
	user.RentedPropertyOwnerName = ownerName
	user.UpdatedAt = time.Now()

	// Save updated user
	err = h.userService.CreateOrUpdateUser(ctx, user)
	if err != nil {
		return fmt.Errorf("failed to update user: %v", err)
	}

	return nil
}
