package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
)

type TenantHandler struct {
	client *firestore.Client
}

type Tenant struct {
	ID         string `json:"id" firestore:"id"`
	PropertyID string `json:"propertyId" firestore:"propertyId"`
	OwnerUID   string `json:"ownerUID" firestore:"ownerUID"`

	// Personal Information
	FirstName        string `json:"firstName" firestore:"firstName"`
	LastName         string `json:"lastName" firestore:"lastName"`
	Email            string `json:"email" firestore:"email"`
	Phone            string `json:"phone" firestore:"phone"`
	EmergencyContact string `json:"emergencyContact" firestore:"emergencyContact"`

	// Lease Information
	LeaseStartDate  string `json:"leaseStartDate" firestore:"leaseStartDate"`
	LeaseEndDate    string `json:"leaseEndDate" firestore:"leaseEndDate"`
	MonthlyRent     string `json:"monthlyRent" firestore:"monthlyRent"`
	SecurityDeposit string `json:"securityDeposit" firestore:"securityDeposit"`

	// Address Information
	PreviousAddress  string `json:"previousAddress" firestore:"previousAddress"`
	EmploymentStatus string `json:"employmentStatus" firestore:"employmentStatus"`
	Employer         string `json:"employer" firestore:"employer"`
	MonthlyIncome    string `json:"monthlyIncome" firestore:"monthlyIncome"`

	// Additional Notes
	Notes string `json:"notes" firestore:"notes"`

	// System Info
	IsActive  bool      `json:"isActive" firestore:"isActive"`
	CreatedAt time.Time `json:"createdAt" firestore:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt" firestore:"updatedAt"`
}

func NewTenantHandler(client *firestore.Client) *TenantHandler {
	return &TenantHandler{client: client}
}

// CreateTenant creates new tenants and adds them to the property document
func (h *TenantHandler) CreateTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Define request structure
	var requestData struct {
		PropertyID string   `json:"propertyId"`
		OwnerUID   string   `json:"ownerUID"`
		Tenants    []Tenant `json:"tenants"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestData); err != nil {
		log.Printf("❌ Error decoding tenant data: %v", err)
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
		log.Printf("❌ Property %s not found: %v", requestData.PropertyID, err)
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Get current tenants array from property (if any)
	var propertyData map[string]interface{}
	if err := propertyDoc.DataTo(&propertyData); err != nil {
		log.Printf("❌ Error reading property data: %v", err)
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
			log.Printf("❌ Failed to create tenant %s %s: %v", tenant.FirstName, tenant.LastName, err)
			http.Error(w, fmt.Sprintf("Failed to create tenant %s %s", tenant.FirstName, tenant.LastName), http.StatusInternalServerError)
			return
		}

		// Convert tenant to map for property storage
		tenantInfo := map[string]interface{}{
			"id":               tenant.ID,
			"firstName":        tenant.FirstName,
			"lastName":         tenant.LastName,
			"email":            tenant.Email,
			"phone":            tenant.Phone,
			"emergencyContact": tenant.EmergencyContact,
			"leaseStartDate":   tenant.LeaseStartDate,
			"leaseEndDate":     tenant.LeaseEndDate,
			"monthlyRent":      tenant.MonthlyRent,
			"securityDeposit":  tenant.SecurityDeposit,
			"previousAddress":  tenant.PreviousAddress,
			"employmentStatus": tenant.EmploymentStatus,
			"employer":         tenant.Employer,
			"monthlyIncome":    tenant.MonthlyIncome,
			"notes":            tenant.Notes,
			"isActive":         tenant.IsActive,
			"createdAt":        tenant.CreatedAt,
			"updatedAt":        tenant.UpdatedAt,
		}

		newTenantInfos = append(newTenantInfos, tenantInfo)
		log.Printf("✅ Tenant created successfully: %s %s for property %s", tenant.FirstName, tenant.LastName, tenant.PropertyID)
	}

	// Update property with new tenants (append to existing)
	allTenants := append(existingTenants, newTenantInfos...)

	_, err = propertyRef.Update(ctx, []firestore.Update{
		{Path: "tenants", Value: allTenants},
		{Path: "updatedAt", Value: time.Now()},
	})
	if err != nil {
		log.Printf("❌ Failed to update property with tenant info: %v", err)
		http.Error(w, "Failed to add tenants to property", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Property %s updated with %d new tenant(s)", requestData.PropertyID, len(newTenantInfos))

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

	var tenants []Tenant
	for {
		doc, err := iter.Next()
		if err != nil {
			break
		}

		var tenant Tenant
		if err := doc.DataTo(&tenant); err != nil {
			log.Printf("❌ Error converting tenant document: %v", err)
			continue
		}

		tenants = append(tenants, tenant)
	}

	log.Printf("📍 Getting tenants for property %s, found %d tenants", propertyID, len(tenants))

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
		log.Printf("❌ Failed to get tenant %s: %v", tenantID, err)
		http.Error(w, "Tenant not found", http.StatusNotFound)
		return
	}

	var tenant Tenant
	if err := doc.DataTo(&tenant); err != nil {
		log.Printf("❌ Error converting tenant document: %v", err)
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
		log.Printf("❌ Error decoding update data: %v", err)
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
				log.Printf("❌ Failed to update tenant field %s: %v", key, err)
			}
		}
	}

	if err != nil {
		log.Printf("❌ Failed to update tenant %s: %v", tenantID, err)
		http.Error(w, "Failed to update tenant", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Tenant %s updated successfully", tenantID)

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
		log.Printf("❌ Failed to delete tenant %s: %v", tenantID, err)
		http.Error(w, "Failed to delete tenant", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Tenant %s deleted successfully", tenantID)

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Tenant deleted successfully",
	})
}
