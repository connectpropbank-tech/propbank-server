package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"shoprop-backend/config"
	"shoprop-backend/models"
	"shoprop-backend/services"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
)

type PropertyHandler struct {
	propertyService          *services.PropertyService
	userService              *services.UserService
	imageService             *services.ImageService
	adminNotificationService *services.AdminNotificationService
	siteSettingsService      *services.SiteSettingsService
	emailService             *services.EmailService
}

func NewPropertyHandler(client *firestore.Client, emailService *services.EmailService) *PropertyHandler {
	// Initialize image service with Firebase Storage
	storageClient := config.GetStorageClient()
	bucketName := config.GetStorageBucket()

	return &PropertyHandler{
		propertyService:          services.NewPropertyService(client),
		userService:              services.NewUserService(client),
		imageService:             services.NewImageService(storageClient, bucketName),
		adminNotificationService: services.NewAdminNotificationService(client),
		siteSettingsService:      services.NewSiteSettingsService(client),
		emailService:             emailService,
	}
}

func (h *PropertyHandler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Log the raw request body for debugging
	bodyBytes, _ := io.ReadAll(r.Body)

	// Reset body for parsing
	r.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))

	var req models.CreatePropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		fmt.Printf("Error decoding JSON: %v\n", err)
		http.Error(w, fmt.Sprintf("Invalid JSON: %v", err), http.StatusBadRequest)
		return
	}

	if len(req.Images) > 0 {
	}

	// Validate required fields
	if req.PropertyTitle == "" || req.PropertyType == "" || req.Location == "" || req.OwnerUID == "" {
		http.Error(w, "Missing required fields: propertyTitle, propertyType, location, ownerUID", http.StatusBadRequest)
		return
	}

	// Get owner information
	owner, err := h.userService.GetUserByID(r.Context(), req.OwnerUID)
	if err != nil {
		http.Error(w, "Owner not found", http.StatusBadRequest)
		return
	}

	// Generate unique property ID for image storage
	propertyID := h.propertyService.GenerateID()

	// Upload images to Firebase Storage (or keep R2 URLs as-is)
	imageURLs := []string{}
	if len(req.Images) > 0 {
		uploadedURLs, err := h.imageService.UploadPropertyImages(r.Context(), req.Images, propertyID)
		if err != nil {
			// continue without images
		}
		imageURLs = uploadedURLs
	}

	// Create property object with comprehensive fields
	property := models.Property{
		ID: propertyID, // Set the generated ID
		// Basic Property Details
		Title:         req.PropertyTitle,
		PropertyType:  req.PropertyType,
		Configuration: req.Configuration,
		ListingType:   req.ListingType,
		IsSold:        req.IsSold,

		// Unit Details
		UnitNumber: req.UnitNumber,
		Floor:      req.Floor,
		Location:   req.Location,
		Address:    req.Location, // Using location as address for compatibility

		// Area Details
		CarpetArea:      req.CarpetArea,
		ConstructedArea: req.ConstructedArea,

		// Tenant Information
		// Tenant Information
		TenantName:   fmt.Sprintf("%s %s", req.TenantFirstName, req.TenantLastName), // Combine for legacy field
		TenantEmail:  req.TenantEmail,
		PersonName:   req.PersonName,
		MobileNumber: req.MobileNumber,
		PrimaryNo:    req.PrimaryNo,
		UltNo:        req.UltNo,

		// Pricing Details
		MonthlyRent:  req.MonthlyRent,
		SellingPrice: req.SellingPrice,

		// Monthly Rent Details
		MonthlyRent1stYear: req.MonthlyRent1stYear,
		MonthlyRent2ndYear: req.MonthlyRent2ndYear,
		MonthlyRent3rdYear: req.MonthlyRent3rdYear,
		MonthlyRent4thYear: req.MonthlyRent4thYear,
		RentFromDate1:      req.RentFromDate1,
		RentToDate1:        req.RentToDate1,
		RentFromDate2:      req.RentFromDate2,
		RentToDate2:        req.RentToDate2,

		// Payment Details
		PaymentDueDate:       req.PaymentDueDate,
		EscalationPercentage: req.EscalationPercentage,
		EscalationAmount:     req.EscalationAmount,

		// Security & Agreement
		SecurityDeposit:    req.SecurityDeposit,
		AgreementPeriod:    req.AgreementPeriod,
		AgreementStartDate: req.AgreementStartDate,
		AgreementEndDate:   req.AgreementEndDate,

		// Notice & Lock-in
		NoticePeriod: req.NoticePeriod,
		LockInPeriod: req.LockInPeriod,

		// Unit Condition & Maintenance
		UnitCondition:         req.UnitCondition,
		MaintenanceToBePaidBy: req.MaintenanceToBePaidBy,
		ProjectCondition:      req.ProjectCondition,
		InternalImages:        req.InternalImages,
		PossessionDate:        req.PossessionDate,
		RentalStatus:          req.RentalStatus,
		FurnishedChecklist:    req.FurnishedChecklist,

		// Images & Comments
		Images:           imageURLs,
		SpecificComments: req.SpecificComments,

		// Owner Info
		OwnerUID:   req.OwnerUID,
		OwnerName:  owner.Name,
		OwnerEmail: owner.Email,
		OwnerPhone: owner.PhoneNumber,
		OwnerRole:  string(owner.Role),

		// Status - set default to "active" if not provided
		Status: "active", // Default status for new properties

		// Pre-leased Details
		IsPreLeased:           req.IsPreLeased,
		PreLeasedType:         req.PreLeasedType,
		AgreementTerm:         req.AgreementTerm,
		LockInPeriodPreLeased: req.LockInPeriodPreLeased,
		RentalIncome:          req.RentalIncome,
		Escalation:            req.Escalation,
		TenantDetails:         req.TenantDetails,
		Purpose:               req.Purpose,
		SpecificRequirement:   req.SpecificRequirement,
	}

	// Logic to populate Tenants slice if property is rented
	if req.RentalStatus == "rented" {
		// Fallback for Tenant Name if split fields are empty
		firstName := req.TenantFirstName
		lastName := req.TenantLastName

		if firstName == "" && lastName == "" && req.TenantName != "" {
			parts := strings.Split(req.TenantName, " ")
			if len(parts) > 0 {
				firstName = parts[0]
			}
			if len(parts) > 1 {
				lastName = strings.Join(parts[1:], " ")
			}
		}

		tenant := models.TenantInfo{
			FirstName:        firstName,
			LastName:         lastName,
			Email:            req.TenantEmail,
			Phone:            req.MobileNumber,
			PaymentDueDate:   req.PaymentDueDate,
			MonthlyRent:      req.MonthlyRent,
			LeaseStartDate:   req.AgreementStartDate,
			LeaseEndDate:     req.AgreementEndDate,
			NoticePeriod:     req.NoticePeriod,
			EmergencyContact: req.TenantEmergencyContact,
			PreviousAddress:  req.TenantPreviousAddress,
			EmploymentStatus: req.TenantEmploymentStatus,
			Employer:         req.TenantEmployer,
			MonthlyIncome:    req.TenantMonthlyIncome,
			IsMarried:        req.TenantIsMarried,
			Spouse:           req.TenantSpouse,
			RentSchedule:     req.RentSchedule, // Also map rent schedule if provided
			IsActive:         true,
			CreatedAt:        time.Now().Format(time.RFC3339),
			UpdatedAt:        time.Now().Format(time.RFC3339),
		}

		// Try to link with existing user by email
		if req.TenantEmail != "" {
			user, err := h.userService.GetUserByEmail(r.Context(), req.TenantEmail)
			if err == nil && user != nil {
				tenant.UserUID = user.UID
			}
		}

		property.Tenants = []models.TenantInfo{tenant}

		// Note: We need to verify if we should call updateUsersWithRentedProperty here.
		// CreateProperty inserts the property.
		// If we found a user, we should update them?
		// Yes, but we need the Property ID which is generated above.
		// And we need to do it after property creation or here if we pass the property ID?
		// Actually updateUsersWithRentedProperty takes propertyID.
		// We will do it AFTER successful creation below.
	}

	// Logic to populate Buyers slice if property is sold
	if req.IsSold {
		buyer := models.BuyerInfo{
			ID:        fmt.Sprintf("buyer_%d", time.Now().UnixNano()),
			FirstName: req.BuyerFirstName,
			LastName:  req.BuyerLastName,
			Phone:     req.BuyerPhone,
			IsActive:  true,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		property.Buyers = []models.BuyerInfo{buyer}
	}

	// Create property in database
	createdProperty, err := h.propertyService.CreateProperty(r.Context(), property)
	if err != nil {
		http.Error(w, "Failed to create property", http.StatusInternalServerError)
		return
	}

	// Create admin notification when property is added
	buyersJSON := ""
	if len(createdProperty.Buyers) > 0 {
		if b, err := json.Marshal(createdProperty.Buyers); err == nil {
			buyersJSON = string(b)
		}
	}

	tenantName := ""
	tenantEmail := ""
	tenantPhone := ""
	if len(createdProperty.Tenants) > 0 {
		tenant := createdProperty.Tenants[0]
		tenantName = fmt.Sprintf("%s %s", tenant.FirstName, tenant.LastName)
		tenantEmail = tenant.Email
		tenantPhone = tenant.Phone
	}

	ownerRole := ""
	if owner != nil {
		ownerRole = string(owner.Role)
	}

	notificationReq := models.CreateAdminNotificationRequest{
		Type:                "property_added",
		Title:               "New Property Added",
		Message:             fmt.Sprintf("New property \"%s\" has been added by owner: %s", createdProperty.Title, createdProperty.OwnerName),
		PropertyID:          createdProperty.ID,
		OwnerID:             createdProperty.OwnerUID,
		OwnerName:           createdProperty.OwnerName,
		OwnerEmail:          createdProperty.OwnerEmail,
		OwnerPhone:          createdProperty.OwnerPhone,
		OwnerRole:           ownerRole,
		TenantName:          tenantName,
		TenantEmail:         tenantEmail,
		TenantPhone:         tenantPhone,
		Buyers:              buyersJSON,
		PropertyTitle:       createdProperty.Title,
		PropertyAddress:     createdProperty.Location,
		PropertyListingType: createdProperty.ListingType,
		Timestamp:           time.Now().Format(time.RFC3339),
		IsRead:              false,
		Priority:            "medium",
	}

	_, _ = h.adminNotificationService.CreateNotification(r.Context(), notificationReq)

	// Update tenant users with the rented property ID if they were linked
	if len(property.Tenants) > 0 {
		// Convert Tenants to interface{} slice for sendTenantAddedEmails
		var tenantsRaw []interface{}
		for _, t := range property.Tenants {
			// Convert TenantInfo to map[string]interface{}
			tData := map[string]interface{}{
				"firstName":      t.FirstName,
				"lastName":       t.LastName,
				"email":          t.Email,
				"phone":          t.Phone,
				"monthlyRent":    t.MonthlyRent,
				"leaseStartDate": t.LeaseStartDate,
				"leaseEndDate":   t.LeaseEndDate,
				"isActive":       t.IsActive,
			}
			tenantsRaw = append(tenantsRaw, tData)
		}

		// Send email notifications
		// Since this is a new property, all tenants are considered "new"
		h.sendTenantAddedEmails(r.Context(), createdProperty, []models.TenantInfo{}, tenantsRaw)

		for _, tenant := range property.Tenants {
			if tenant.UserUID != "" {
				// Update user with rented property information
				user, err := h.userService.GetUserByID(r.Context(), tenant.UserUID)
				if err == nil {
					user.RentedPropertyID = createdProperty.ID
					user.RentedPropertyOwnerID = createdProperty.OwnerUID
					user.RentedPropertyOwnerName = createdProperty.OwnerName
					user.UpdatedAt = time.Now()

					// Save updated user
					// We ignore error here to not block response, but could log it
					_ = h.userService.CreateOrUpdateUser(r.Context(), user)
				}
			}
		}
	}

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Property created successfully",
		"property": models.PropertyResponse{
			ID:                    createdProperty.ID,
			Title:                 createdProperty.Title,
			Description:           createdProperty.Description,
			Price:                 createdProperty.Price,
			Address:               createdProperty.Address,
			City:                  createdProperty.City,
			State:                 createdProperty.State,
			ZipCode:               createdProperty.ZipCode,
			PropertyType:          createdProperty.PropertyType,
			ListingType:           createdProperty.ListingType,
			Configuration:         createdProperty.Configuration,
			UnitNumber:            createdProperty.UnitNumber,
			Floor:                 createdProperty.Floor,
			Location:              createdProperty.Location,
			CarpetArea:            createdProperty.CarpetArea,
			ConstructedArea:       createdProperty.ConstructedArea,
			SquareFeet:            createdProperty.SquareFeet,
			TenantName:            createdProperty.TenantName,
			PersonName:            createdProperty.PersonName,
			MobileNumber:          createdProperty.MobileNumber,
			PrimaryNo:             createdProperty.PrimaryNo,
			UltNo:                 createdProperty.UltNo,
			MonthlyRent:           createdProperty.MonthlyRent,
			SellingPrice:          createdProperty.SellingPrice,
			MonthlyRent1stYear:    createdProperty.MonthlyRent1stYear,
			MonthlyRent2ndYear:    createdProperty.MonthlyRent2ndYear,
			MonthlyRent3rdYear:    createdProperty.MonthlyRent3rdYear,
			MonthlyRent4thYear:    createdProperty.MonthlyRent4thYear,
			RentFromDate1:         createdProperty.RentFromDate1,
			RentToDate1:           createdProperty.RentToDate1,
			RentFromDate2:         createdProperty.RentFromDate2,
			RentToDate2:           createdProperty.RentToDate2,
			PaymentDueDate:        createdProperty.PaymentDueDate,
			EscalationPercentage:  createdProperty.EscalationPercentage,
			EscalationAmount:      createdProperty.EscalationAmount,
			SecurityDeposit:       createdProperty.SecurityDeposit,
			AgreementPeriod:       createdProperty.AgreementPeriod,
			AgreementStartDate:    createdProperty.AgreementStartDate,
			AgreementEndDate:      createdProperty.AgreementEndDate,
			NoticePeriod:          createdProperty.NoticePeriod,
			LockInPeriod:          createdProperty.LockInPeriod,
			UnitCondition:         createdProperty.UnitCondition,
			MaintenanceToBePaidBy: createdProperty.MaintenanceToBePaidBy,
			ProjectCondition:      createdProperty.ProjectCondition,
			PossessionDate:        createdProperty.PossessionDate,
			RentalStatus:          createdProperty.RentalStatus,
			FurnishedChecklist:    createdProperty.FurnishedChecklist,
			Images:                createdProperty.Images,
			SpecificComments:      createdProperty.SpecificComments,
			Tenants:               createdProperty.Tenants,
			Buyers:                createdProperty.Buyers,
			OwnerUID:              createdProperty.OwnerUID,
			OwnerName:             createdProperty.OwnerName,
			OwnerEmail:            createdProperty.OwnerEmail,
			WantToSell:            createdProperty.WantToSell,
			Status:                createdProperty.Status,
			IsActive:              createdProperty.IsActive,
			IsSold:                createdProperty.IsSold,
			CreatedAt:             createdProperty.CreatedAt,
			UpdatedAt:             createdProperty.UpdatedAt,
			Bedrooms:              createdProperty.Bedrooms,
			Bathrooms:             createdProperty.Bathrooms,
		},
	})
}

func (h *PropertyHandler) GetAllProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check if filtering by listing type
	listingType := r.URL.Query().Get("listingType")
	includeUnavailable := r.URL.Query().Get("all") == "true"

	var properties []models.Property
	var err error

	if listingType != "" {
		properties, err = h.propertyService.GetPropertiesByListingType(r.Context(), listingType, includeUnavailable)
	} else {
		properties, err = h.propertyService.GetAllProperties(r.Context(), includeUnavailable)
	}

	if err != nil {
		http.Error(w, "Failed to get properties", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var propertyResponses []models.PropertyResponse
	for _, property := range properties {
		propertyResponses = append(propertyResponses, h.convertToPropertyResponse(property))
	}

	response := map[string]interface{}{
		"success":    true,
		"properties": propertyResponses,
		"count":      len(propertyResponses),
	}

	// Get site settings for dynamic quote
	siteSettings, err := h.siteSettingsService.GetSiteSettings(r.Context())
	if err == nil && siteSettings != nil {
		response["quote"] = siteSettings.Quote
		response["quoteAuthor"] = siteSettings.QuoteAuthor
		response["heroTitle"] = siteSettings.HeroTitle
		response["heroSubtitle"] = siteSettings.HeroSubtitle
		response["announcementText"] = siteSettings.AnnouncementText
		response["isAnnouncementActive"] = siteSettings.IsAnnouncementActive
		// Always include bannerImages, even if empty
		if siteSettings.BannerImages != nil {
			response["bannerImages"] = siteSettings.BannerImages
		} else {
			response["bannerImages"] = []string{}
		}
	}

	if listingType != "" {
		response["listingType"] = listingType
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

func (h *PropertyHandler) GetProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract property ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/properties/")
	if path == "" {
		http.Error(w, "Property ID is required", http.StatusBadRequest)
		return
	}

	property, err := h.propertyService.GetPropertyByID(r.Context(), path)
	if err != nil {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"property": h.convertToPropertyResponse(*property),
	})
}

func (h *PropertyHandler) GetPropertiesByOwner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ownerUID := r.URL.Query().Get("ownerUID")
	if ownerUID == "" {
		http.Error(w, "ownerUID parameter is required", http.StatusBadRequest)
		return
	}

	// Get ALL properties for this owner (regardless of isActive status)
	ownedProperties, err := h.propertyService.GetPropertiesByOwner(r.Context(), ownerUID)
	if err != nil {
		http.Error(w, "Failed to get properties", http.StatusInternalServerError)
		return
	}

	// Also get properties where the user is a tenant
	tenantProperties, err := h.propertyService.GetPropertiesByTenantUID(r.Context(), ownerUID)
	if err != nil {
		tenantProperties = []models.Property{} // Continue with empty tenant properties if error
	}

	// Create a map to track property IDs to avoid duplicates (in case a user is both owner and tenant)
	propertyMap := make(map[string]*models.PropertyResponse)

	// Convert owned properties to response format with "owner" role
	for _, property := range ownedProperties {
		resp := h.convertToPropertyResponse(property)
		resp.UserRole = "owner"
		propertyMap[property.ID] = &resp
	}

	// Convert tenant properties to response format with "tenant" role
	// If a property already exists (user is both owner and tenant), keep the "owner" role
	for _, property := range tenantProperties {
		if _, exists := propertyMap[property.ID]; !exists {
			// Only add if not already in map (user is tenant only, not owner)
			propertyMap[property.ID] = &models.PropertyResponse{
				ID:                    property.ID,
				Title:                 property.Title,
				Description:           property.Description,
				Price:                 property.Price,
				Address:               property.Address,
				City:                  property.City,
				State:                 property.State,
				ZipCode:               property.ZipCode,
				PropertyType:          property.PropertyType,
				ListingType:           property.ListingType,
				Configuration:         property.Configuration,
				UnitNumber:            property.UnitNumber,
				Floor:                 property.Floor,
				Location:              property.Location,
				CarpetArea:            property.CarpetArea,
				ConstructedArea:       property.ConstructedArea,
				SquareFeet:            property.SquareFeet,
				TenantName:            property.TenantName,
				PersonName:            property.PersonName,
				MobileNumber:          property.MobileNumber,
				PrimaryNo:             property.PrimaryNo,
				UltNo:                 property.UltNo,
				MonthlyRent:           property.MonthlyRent,
				SellingPrice:          property.SellingPrice,
				MonthlyRent1stYear:    property.MonthlyRent1stYear,
				MonthlyRent2ndYear:    property.MonthlyRent2ndYear,
				MonthlyRent3rdYear:    property.MonthlyRent3rdYear,
				MonthlyRent4thYear:    property.MonthlyRent4thYear,
				RentFromDate1:         property.RentFromDate1,
				RentToDate1:           property.RentToDate1,
				RentFromDate2:         property.RentFromDate2,
				RentToDate2:           property.RentToDate2,
				PaymentDueDate:        property.PaymentDueDate,
				EscalationPercentage:  property.EscalationPercentage,
				EscalationAmount:      property.EscalationAmount,
				SecurityDeposit:       property.SecurityDeposit,
				AgreementPeriod:       property.AgreementPeriod,
				AgreementStartDate:    property.AgreementStartDate,
				AgreementEndDate:      property.AgreementEndDate,
				NoticePeriod:          property.NoticePeriod,
				LockInPeriod:          property.LockInPeriod,
				UnitCondition:         property.UnitCondition,
				MaintenanceToBePaidBy: property.MaintenanceToBePaidBy,
				ProjectCondition:      property.ProjectCondition,
				InternalImages:        property.InternalImages,
				PossessionDate:        property.PossessionDate,
				RentalStatus:          property.RentalStatus,
				FurnishedChecklist:    property.FurnishedChecklist,
				Images:                property.Images,
				SpecificComments:      property.SpecificComments,
				Tenants:               property.Tenants,
				Buyers:                property.Buyers,
				OwnerUID:              property.OwnerUID,
				OwnerName:             property.OwnerName,
				OwnerEmail:            property.OwnerEmail,
				WantToSell:            property.WantToSell,
				Status:                property.Status,
				IsActive:              property.IsActive,
				CreatedAt:             property.CreatedAt,
				UpdatedAt:             property.UpdatedAt,
				Bedrooms:              property.Bedrooms,
				Bathrooms:             property.Bathrooms,
				UserRole:              "tenant", // Mark as tenant
			}
		}
	}

	// Convert map to slice
	var propertyResponses []models.PropertyResponse
	for _, propertyResp := range propertyMap {
		propertyResponses = append(propertyResponses, *propertyResp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"properties": propertyResponses,
		"count":      len(propertyResponses),
		"ownerUID":   ownerUID,
	})

}

// GetPropertiesByTenant gets all properties where the user is a tenant
func (h *PropertyHandler) GetPropertiesByTenant(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Check for tenantUID query parameter first (User preferred method)
	// Check for tenantUID query parameter first (User preferred method)
	tenantUID := r.URL.Query().Get("tenantUID")
	userEmail := r.URL.Query().Get("userEmail")

	if tenantUID != "" {
		// Fetch by UID
		propertiesUID, err := h.propertyService.GetPropertiesByTenantUID(r.Context(), tenantUID)
		if err != nil {
			http.Error(w, "Failed to get properties by UID", http.StatusInternalServerError)
			return
		}

		// If email is also provided, fetch by Email and merge
		if userEmail != "" {
			propertiesEmail, err := h.propertyService.GetPropertiesByTenantEmail(r.Context(), userEmail)
			if err != nil {
				// Log error but continue with UID results
			} else {
				// Merge results
				propertyMap := make(map[string]models.Property)
				for _, p := range propertiesUID {
					propertyMap[p.ID] = p
				}
				for _, p := range propertiesEmail {
					propertyMap[p.ID] = p
				}

				// Convert back to slice
				var mergedProperties []models.Property
				for _, p := range propertyMap {
					mergedProperties = append(mergedProperties, p)
				}
				h.respondWithProperties(w, mergedProperties)
				return
			}
		}

		h.respondWithProperties(w, propertiesUID)
		return
	}

	// Fallback: Get user ID from header (My previous implementation)
	userID := r.Header.Get("X-User-ID")
	if userID != "" {
		// Get User to find their email (or just use UID if we want to switch entirely)
		// Since we have GetPropertiesByTenantUID, we can just use that!
		properties, err := h.propertyService.GetPropertiesByTenantUID(r.Context(), userID)
		if err != nil {
			http.Error(w, "Failed to get properties", http.StatusInternalServerError)
			return
		}
		h.respondWithProperties(w, properties)
		return
	}

	// Legacy Fallback: Check for userEmail query parameter
	if userEmail != "" {
		properties, err := h.propertyService.GetPropertiesByTenantEmail(r.Context(), userEmail)
		if err != nil {
			http.Error(w, "Failed to get properties", http.StatusInternalServerError)
			return
		}
		h.respondWithProperties(w, properties)
		return
	}

	http.Error(w, "tenantUID, X-User-ID header, or userEmail parameter is required", http.StatusBadRequest)
}

func (h *PropertyHandler) respondWithProperties(w http.ResponseWriter, properties []models.Property) {
	// Convert to response format
	var propertyResponses []models.PropertyResponse
	for _, property := range properties {
		resp := h.convertToPropertyResponse(property)
		resp.UserRole = "tenant"
		propertyResponses = append(propertyResponses, resp)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"properties": propertyResponses,
		"count":      len(propertyResponses),
	})
}

func (h *PropertyHandler) GetArchivedPropertiesByOwner(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	ownerUID := r.URL.Query().Get("ownerUID")
	if ownerUID == "" {
		http.Error(w, "ownerUID parameter is required", http.StatusBadRequest)
		return
	}

	properties, err := h.propertyService.GetArchivedPropertiesByOwner(r.Context(), ownerUID)
	if err != nil {
		http.Error(w, "Failed to get archived properties", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var propertyResponses []models.PropertyResponse
	for _, property := range properties {
		propertyResponses = append(propertyResponses, h.convertToPropertyResponse(property))
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"properties": propertyResponses,
		"count":      len(propertyResponses),
		"ownerUID":   ownerUID,
	})

}

func (h *PropertyHandler) UpdateProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPut && r.Method != http.MethodPatch {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Extract property ID from URL path
	path := strings.TrimPrefix(r.URL.Path, "/properties/")
	if path == "" {
		http.Error(w, "Property ID is required", http.StatusBadRequest)
		return
	}

	// Accept any JSON data for flexible updates
	var updateRequest map[string]interface{}
	if err := json.NewDecoder(r.Body).Decode(&updateRequest); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Get existing property first
	existingProperty, err := h.propertyService.GetPropertyByID(r.Context(), path)
	if err != nil {
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	// Check if ownerUID is provided for verification (optional)
	if ownerUID, hasOwner := updateRequest["ownerUID"].(string); hasOwner {
		// Verify ownership if ownerUID is provided
		if existingProperty.OwnerUID != ownerUID {
			http.Error(w, "Unauthorized: You can only update your own properties", http.StatusForbidden)
			return
		}
	}

	// Handle image updates - keep existing images if no new images provided
	imageURLs := existingProperty.Images
	if images, hasImages := updateRequest["images"].([]interface{}); hasImages && len(images) > 0 {
		// Convert interface{} slice to string slice
		var imageStrings []string
		for _, img := range images {
			if imgStr, ok := img.(string); ok {
				imageStrings = append(imageStrings, imgStr)
			}
		}

		if len(imageStrings) > 0 {
			uploadedURLs, err := h.imageService.UploadPropertyImages(r.Context(), imageStrings, path)
			if err != nil {
			} else {
				imageURLs = uploadedURLs
			}
		}
	}

	// Create update data from the request, only including provided fields
	updateData := make(map[string]interface{})

	// Copy all provided fields from updateRequest except system fields
	for key, value := range updateRequest {
		switch key {
		case "id", "createdAt": // Skip system fields that shouldn't be updated
			continue
		case "images":
			updateData[key] = imageURLs // Use processed image URLs
		case "status":
			// Handle status field: "active" or "inactive"
			// Keep status and isActive independent - do NOT sync
			if statusStr, ok := value.(string); ok {
				if statusStr == "active" || statusStr == "inactive" {
					updateData[key] = statusStr
				} else {
				}
			} else {
				updateData[key] = value
			}
		case "isActive", "isRented", "isSold":
			// isActive=false must ONLY be set via the explicit delete/archive endpoint,
			// never via a general property edit. Skip false values here.
			if boolVal, ok := value.(bool); ok {
				if boolVal {
					updateData[key] = true // only allow explicit true
				}
				// Skip if false — protects against accidental property disappearance
			} else if strVal, ok := value.(string); ok {
				if strVal == "true" {
					updateData[key] = true
				}
				// Skip "false" strings too
			}
		case "tenants":
			if tenantsArray, ok := value.([]interface{}); ok {
				var validTenants []interface{}
				for _, t := range tenantsArray {
					if tMap, ok := t.(map[string]interface{}); ok {
						email, _ := tMap["email"].(string)
						first, _ := tMap["firstName"].(string)
						last, _ := tMap["lastName"].(string)
						phone, _ := tMap["phone"].(string)
						
						if strings.TrimSpace(email) != "" || strings.TrimSpace(first) != "" || strings.TrimSpace(last) != "" || strings.TrimSpace(phone) != "" {
							validTenants = append(validTenants, t)
						}
					} else {
						// If it's not a map, keep it just in case
						validTenants = append(validTenants, t)
					}
				}
				updateData[key] = validTenants
			} else {
				updateData[key] = value
			}
		default:
			updateData[key] = value
		}
	}

	// Always update the timestamp
	updateData["updatedAt"] = time.Now()

	// Check if wantToSell is being toggled ON or OFF before updating (we need existing property for notification)
	var wantToSellBeingSetToTrue bool = false
	var wantToSellBeingSetToFalse bool = false
	if wantToSell, hasWantToSell := updateData["wantToSell"]; hasWantToSell {
		if wantToSellBool, ok := wantToSell.(bool); ok {
			if wantToSellBool && !existingProperty.WantToSell {
				// Changing from false to true
				wantToSellBeingSetToTrue = true
			} else if !wantToSellBool && existingProperty.WantToSell {
				// Changing from true to false (cancelled)
				wantToSellBeingSetToFalse = true
			}
		}
	}

	// Update property in database
	updatedProperty, err := h.propertyService.UpdateProperty(r.Context(), path, updateData)
	if err != nil {
		http.Error(w, "Failed to update property", http.StatusInternalServerError)
		return
	}

	// Check if rent was increased
	if newRentStr, ok := updateData["monthlyRent"].(string); ok && existingProperty.MonthlyRent != "" && newRentStr != "" && newRentStr != existingProperty.MonthlyRent {
		oldRentVal, err1 := parsePrice(existingProperty.MonthlyRent)
		newRentVal, err2 := parsePrice(newRentStr)
		
		if err1 == nil && err2 == nil && newRentVal > oldRentVal {
			tenantName := updatedProperty.TenantName
			tenantEmail := updatedProperty.TenantEmail
			
			// Try to get from active tenants array if not at root
			if tenantEmail == "" && len(updatedProperty.Tenants) > 0 {
				for _, t := range updatedProperty.Tenants {
					if t.IsActive && t.Email != "" {
						if tenantName == "" {
							tenantName = t.FirstName + " " + t.LastName
						}
						tenantEmail = t.Email
						break
					}
				}
			}
			
			if tenantEmail != "" {
				rentUpdateData := services.RentUpdateEmailData{
					TenantName:      tenantName,
					TenantEmail:     tenantEmail,
					PropertyTitle:   updatedProperty.Title,
					PropertyAddress: updatedProperty.Location,
					OldRent:         existingProperty.MonthlyRent,
					NewRent:         newRentStr,
				}
				
				go func() {
					_ = h.emailService.SendRentUpdateEmail(rentUpdateData)
				}()
			}
		}
	}

	// If tenants were updated, map property information to users with userUID and send email notifications
	if tenants, hasTenants := updateData["tenants"]; hasTenants {
		if tenantsArray, ok := tenants.([]interface{}); ok {
			// Map property information to users with userUID
			h.updateUsersWithRentedProperty(r.Context(), path, updatedProperty.OwnerUID, updatedProperty.OwnerName, tenantsArray)

			// Send email notifications for newly added tenants
			h.sendTenantAddedEmails(r.Context(), updatedProperty, existingProperty.Tenants, tenantsArray)

			// Send email notifications for Payment Due Date updates
			h.sendPaymentDueDateUpdateEmails(r.Context(), updatedProperty, existingProperty.Tenants, tenantsArray)
		}
	}

	// Handle "Want to Sell" toggle - create notification if toggled ON
	if wantToSellBeingSetToTrue {
		h.createWantToSellNotification(r.Context(), updatedProperty)
	}

	// Handle "Want to Sell" toggle OFF - create notification if cancelled
	if wantToSellBeingSetToFalse {
		h.createWantToSellCancelledNotification(r.Context(), updatedProperty)
	}

	// Log the isActive status after update

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Property updated successfully",
		"property": h.convertToPropertyResponse(*updatedProperty),
	})
}

// SearchProperties handles property search requests
func (h *PropertyHandler) SearchProperties(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Get query parameters
	query := r.URL.Query().Get("q")
	listingType := r.URL.Query().Get("listingType")
	projectCondition := r.URL.Query().Get("projectCondition")
	includeUnavailable := r.URL.Query().Get("all") == "true"

	// Search properties
	properties, err := h.propertyService.SearchProperties(r.Context(), query, listingType, projectCondition, includeUnavailable)
	if err != nil {
		http.Error(w, "Failed to search properties", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var propertyResponses []models.PropertyResponse
	for _, property := range properties {
		propertyResponses = append(propertyResponses, h.convertToPropertyResponse(property))
	}

	response := map[string]interface{}{
		"success":    true,
		"properties": propertyResponses,
		"count":      len(propertyResponses),
		"query":      query,
	}

	if listingType != "" {
		response["listingType"] = listingType
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}

// Helper function to parse price strings
func parsePrice(priceStr string) (float64, error) {
	// Remove common currency symbols and spaces
	cleanPrice := strings.ReplaceAll(priceStr, "₹", "")
	cleanPrice = strings.ReplaceAll(cleanPrice, ",", "")
	cleanPrice = strings.TrimSpace(cleanPrice)

	if cleanPrice == "" {
		return 0, fmt.Errorf("empty price")
	}

	var price float64
	if _, err := fmt.Sscanf(cleanPrice, "%f", &price); err != nil {
		return 0, err
	}

	return price, nil
}

// convertToPropertyResponse converts a Property model to a PropertyResponse
func (h *PropertyHandler) convertToPropertyResponse(property models.Property) models.PropertyResponse {
	displayPrice := property.Price
	if displayPrice == 0 {
		if property.ListingType == "rent" && property.MonthlyRent != "" {
			if parsedRent, err := parsePrice(property.MonthlyRent); err == nil {
				displayPrice = parsedRent
			}
		} else if property.ListingType == "sell" && property.SellingPrice != "" {
			if parsedPrice, err := parsePrice(property.SellingPrice); err == nil {
				displayPrice = parsedPrice
			}
		}
	}

	return models.PropertyResponse{
		ID:                    property.ID,
		Title:                 property.Title,
		Description:           property.Description,
		Price:                 displayPrice,
		Address:               property.Address,
		City:                  property.City,
		State:                 property.State,
		ZipCode:               property.ZipCode,
		PropertyType:          property.PropertyType,
		ListingType:           property.ListingType,
		Configuration:         property.Configuration,
		UnitNumber:            property.UnitNumber,
		Floor:                 property.Floor,
		Location:              property.Location,
		CarpetArea:            property.CarpetArea,
		ConstructedArea:       property.ConstructedArea,
		SquareFeet:            property.SquareFeet,
		TenantName:            property.TenantName,
		PersonName:            property.PersonName,
		MobileNumber:          property.MobileNumber,
		PrimaryNo:             property.PrimaryNo,
		UltNo:                 property.UltNo,
		MonthlyRent:           property.MonthlyRent,
		SellingPrice:          property.SellingPrice,
		MonthlyRent1stYear:    property.MonthlyRent1stYear,
		MonthlyRent2ndYear:    property.MonthlyRent2ndYear,
		MonthlyRent3rdYear:    property.MonthlyRent3rdYear,
		MonthlyRent4thYear:    property.MonthlyRent4thYear,
		RentFromDate1:         property.RentFromDate1,
		RentToDate1:           property.RentToDate1,
		RentFromDate2:         property.RentFromDate2,
		RentToDate2:           property.RentToDate2,
		PaymentDueDate:        property.PaymentDueDate,
		EscalationPercentage:  property.EscalationPercentage,
		EscalationAmount:      property.EscalationAmount,
		SecurityDeposit:       property.SecurityDeposit,
		AgreementPeriod:       property.AgreementPeriod,
		AgreementStartDate:    property.AgreementStartDate,
		AgreementEndDate:      property.AgreementEndDate,
		NoticePeriod:          property.NoticePeriod,
		LockInPeriod:          property.LockInPeriod,
		UnitCondition:         property.UnitCondition,
		MaintenanceToBePaidBy: property.MaintenanceToBePaidBy,
		ProjectCondition:      property.ProjectCondition,
		PossessionDate:        property.PossessionDate,
		RentalStatus:          property.RentalStatus,
		FurnishedChecklist:    property.FurnishedChecklist,
		Images:                property.Images,
		SpecificComments:      property.SpecificComments,
		Tenants:               property.Tenants,
		Buyers:                property.Buyers,
		OwnerUID:              property.OwnerUID,
		OwnerName:             property.OwnerName,
		OwnerEmail:            property.OwnerEmail,
		OwnerPhone:            property.OwnerPhone,
		OwnerRole:             property.OwnerRole,
		WantToSell:            property.WantToSell,
		Status:                property.Status,
		IsActive:              property.IsActive,
		IsSold:                property.IsSold,
		CreatedAt:             property.CreatedAt,
		UpdatedAt:             property.UpdatedAt,
		Bedrooms:              property.Bedrooms,
		Bathrooms:             property.Bathrooms,
		RentSchedule:          property.RentSchedule,

		// Pre-leased Details
		IsPreLeased:           property.IsPreLeased,
		PreLeasedType:         property.PreLeasedType,
		AgreementTerm:         property.AgreementTerm,
		LockInPeriodPreLeased: property.LockInPeriodPreLeased,
		RentalIncome:          property.RentalIncome,
		Escalation:            property.Escalation,
		TenantDetails:         property.TenantDetails,
		Purpose:               property.Purpose,
		SpecificRequirement:   property.SpecificRequirement,
	}
}

// updateUsersWithRentedProperty updates users' records with rented property information when tenants are added
func (h *PropertyHandler) updateUsersWithRentedProperty(ctx context.Context, propertyID, ownerUID, ownerName string, tenants []interface{}) {
	for _, tenantInterface := range tenants {
		tenantMap, ok := tenantInterface.(map[string]interface{})
		if !ok {
			continue
		}

		// Check if tenant has a userUID
		userUID, hasUserUID := tenantMap["userUID"].(string)
		if !hasUserUID || userUID == "" {
			continue
		}

		// Update user with rented property information
		user, err := h.userService.GetUserByID(ctx, userUID)
		if err != nil {
			continue
		}

		// Update user with rented property information
		user.RentedPropertyID = propertyID
		user.RentedPropertyOwnerID = ownerUID
		user.RentedPropertyOwnerName = ownerName
		user.UpdatedAt = time.Now()

		// Save updated user
		err = h.userService.CreateOrUpdateUser(ctx, user)
		if err != nil {
		} else {
		}
	}
}

// createWantToSellNotification creates an admin notification when owner toggles "Want to Sell" ON
func (h *PropertyHandler) createWantToSellNotification(ctx context.Context, property *models.Property) {
	// Get owner details to include phone number
	owner, err := h.userService.GetUserByID(ctx, property.OwnerUID)
	if err != nil {
		// Continue without phone number
	}

	ownerPhone := ""
	ownerEmail := property.OwnerEmail
	if owner != nil {
		ownerPhone = owner.PhoneNumber
		if ownerEmail == "" {
			ownerEmail = owner.Email
		}
	}

	// Get tenant info if available
	tenantName := ""
	tenantEmail := ""
	tenantPhone := ""
	if len(property.Tenants) > 0 {
		tenant := property.Tenants[0]
		tenantName = fmt.Sprintf("%s %s", tenant.FirstName, tenant.LastName)
		tenantEmail = tenant.Email
		tenantPhone = tenant.Phone
	}

	// Get buyers info if available
	buyersJSON := ""
	if len(property.Buyers) > 0 {
		if b, err := json.Marshal(property.Buyers); err == nil {
			buyersJSON = string(b)
		}
	}

	ownerRole := ""
	if owner != nil {
		ownerRole = string(owner.Role)
	}

	// Create notification request
	notificationReq := models.CreateAdminNotificationRequest{
		Type:       "want_to_sell",
		Title:      "Property Owner Wants to Sell",
		Message:    fmt.Sprintf("Property owner %s wants to sell property: %s", property.OwnerName, property.Title),
		PropertyID: property.ID,
		OwnerID:    property.OwnerUID,
		OwnerName:  property.OwnerName,
		OwnerPhone: ownerPhone,
		OwnerEmail: ownerEmail,
		OwnerRole:  ownerRole,
		// Tenant info
		TenantName:  tenantName,
		TenantEmail: tenantEmail,
		TenantPhone: tenantPhone,
		// Buyers info
		Buyers:    buyersJSON,
		Timestamp: time.Now().Format(time.RFC3339),
		IsRead:    false,
		Priority:  "high",
	}

	// Create notification
	_, err = h.adminNotificationService.CreateNotification(ctx, notificationReq)
	if err != nil {
	} else {
	}
}

// createWantToSellCancelledNotification creates an admin notification when owner cancels/toggles OFF "Want to Sell"
func (h *PropertyHandler) createWantToSellCancelledNotification(ctx context.Context, property *models.Property) {
	// Get owner details to include phone number
	owner, err := h.userService.GetUserByID(ctx, property.OwnerUID)
	if err != nil {
		// Continue without phone number
	}

	ownerPhone := ""
	ownerEmail := property.OwnerEmail
	if owner != nil {
		ownerPhone = owner.PhoneNumber
		if ownerEmail == "" {
			ownerEmail = owner.Email
		}
	}

	// Get tenant info if available
	tenantName := ""
	tenantEmail := ""
	tenantPhone := ""
	if len(property.Tenants) > 0 {
		tenant := property.Tenants[0]
		tenantName = fmt.Sprintf("%s %s", tenant.FirstName, tenant.LastName)
		tenantEmail = tenant.Email
		tenantPhone = tenant.Phone
	}

	// Get buyers info if available
	buyersJSON := ""
	if len(property.Buyers) > 0 {
		if b, err := json.Marshal(property.Buyers); err == nil {
			buyersJSON = string(b)
		}
	}

	ownerRole := ""
	if owner != nil {
		ownerRole = string(owner.Role)
	}

	// Create notification request
	notificationReq := models.CreateAdminNotificationRequest{
		Type:       "want_to_sell_cancelled",
		Title:      "Property Owner Cancelled Sell Request",
		Message:    fmt.Sprintf("Property owner %s has cancelled the request to sell property: %s", property.OwnerName, property.Title),
		PropertyID: property.ID,
		OwnerID:    property.OwnerUID,
		OwnerName:  property.OwnerName,
		OwnerPhone: ownerPhone,
		OwnerEmail: ownerEmail,
		OwnerRole:  ownerRole,
		// Tenant info
		TenantName:  tenantName,
		TenantEmail: tenantEmail,
		TenantPhone: tenantPhone,
		// Buyers info
		Buyers:    buyersJSON,
		Timestamp: time.Now().Format(time.RFC3339),
		IsRead:    false,
		Priority:  "medium",
	}

	// Create notification
	_, err = h.adminNotificationService.CreateNotification(ctx, notificationReq)
	if err != nil {
	} else {
	}
}

// sendTenantAddedEmails sends email notifications to both owner and newly added tenants
func (h *PropertyHandler) sendTenantAddedEmails(ctx context.Context, property *models.Property, existingTenants []models.TenantInfo, newTenantsRaw []interface{}) {
	// Get owner details
	owner, err := h.userService.GetUserByID(ctx, property.OwnerUID)
	if err != nil {
	}

	ownerName := property.OwnerName
	ownerEmail := property.OwnerEmail
	ownerPhone := ""
	if owner != nil {
		ownerName = owner.Name
		if owner.Email != "" {
			ownerEmail = owner.Email
		}
		ownerPhone = owner.PhoneNumber
	}

	// Create a map of existing tenant emails to check for new tenants
	existingTenantEmails := make(map[string]bool)
	for _, tenant := range existingTenants {
		if tenant.Email != "" {
			existingTenantEmails[strings.ToLower(strings.TrimSpace(tenant.Email))] = true
		}
	}

	// Process new tenants
	for _, tenantInterface := range newTenantsRaw {
		tenantMap, ok := tenantInterface.(map[string]interface{})
		if !ok {
			fmt.Printf("[PropertyHandler] WARN: tenant entry is not a map\n")
			continue
		}

		tenantEmail, _ := tenantMap["email"].(string)
		tenantEmailClean := strings.ToLower(strings.TrimSpace(tenantEmail))
		tenantFirstName, _ := tenantMap["firstName"].(string)
		tenantLastName, _ := tenantMap["lastName"].(string)
		tenantPhone, _ := tenantMap["phone"].(string)
		isActive := false
		if val, ok := tenantMap["isActive"].(bool); ok {
			isActive = val
		} else if val, ok := tenantMap["isActive"].(string); ok {
			isActive = (val == "true")
		}

		// Extract lease dates and rent from tenant map if present
		leaseStart, _ := tenantMap["leaseStartDate"].(string)
		leaseEnd, _ := tenantMap["leaseEndDate"].(string)
		tenantRent, _ := tenantMap["monthlyRent"].(string)

		// Skip if tenant email already existed (not a new tenant) or if inactive
		if tenantEmailClean != "" && existingTenantEmails[tenantEmailClean] {
			fmt.Printf("[PropertyHandler] SKIP: Tenant %s already exists in property\n", tenantEmail)
			continue
		}

		if !isActive {
			fmt.Printf("[PropertyHandler] SKIP: Tenant %s is not active\n", tenantEmail)
			continue
		}

		// Skip if the tenant is just a "ghost" empty entry with no valid data
		if tenantEmailClean == "" && strings.TrimSpace(tenantFirstName) == "" && strings.TrimSpace(tenantLastName) == "" && strings.TrimSpace(tenantPhone) == "" {
			fmt.Printf("[PropertyHandler] SKIP: Tenant is empty/ghost entry\n")
			continue
		}

		tenantName := strings.TrimSpace(tenantFirstName + " " + tenantLastName)
		if tenantName == "" {
			tenantName = "Tenant"
		}

		// Use tenant-specific values if available, otherwise fallback to property values
		agreementStart := leaseStart
		if agreementStart == "" {
			agreementStart = property.AgreementStartDate
		}
		agreementEnd := leaseEnd
		if agreementEnd == "" {
			agreementEnd = property.AgreementEndDate
		}
		monthlyRent := tenantRent
		if monthlyRent == "" {
			monthlyRent = property.MonthlyRent
		}

		fmt.Printf("[PropertyHandler] TRIGGER: Sending welcome email to %s for property %s\n", tenantEmail, property.Title)

		// Send email notification
		go func(tName, tEmail, tPhone, aStart, aEnd, mRent string) {
			emailData := services.TenantAddedEmailData{
				TenantName:      tName,
				TenantEmail:     tEmail,
				TenantPhone:     tPhone,
				OwnerName:       ownerName,
				OwnerEmail:      ownerEmail,
				OwnerPhone:      ownerPhone,
				PropertyTitle:   property.Title,
				PropertyAddress: property.Location,
				PropertyType:    property.PropertyType,
				MonthlyRent:     mRent,
				AgreementStart:  aStart,
				AgreementEnd:    aEnd,
			}

			if err := h.emailService.SendTenantAddedNotification(emailData); err != nil {
				fmt.Printf("[PropertyHandler] ERROR sending tenant emails for %s: %v\n", tEmail, err)
			} else {
				fmt.Printf("[PropertyHandler] SUCCESS tenant emails triggered for %s\n", tEmail)
			}
		}(tenantName, tenantEmail, tenantPhone, agreementStart, agreementEnd, monthlyRent)
	}
}

// sendPaymentDueDateUpdateEmails sends email notifications when payment due date is updated
func (h *PropertyHandler) sendPaymentDueDateUpdateEmails(ctx context.Context, property *models.Property, existingTenants []models.TenantInfo, newTenantsRaw []interface{}) {
	// Get owner details
	owner, err := h.userService.GetUserByID(ctx, property.OwnerUID)
	if err != nil {
		// Log error but continue with property details
	}

	ownerName := property.OwnerName
	ownerEmail := property.OwnerEmail
	ownerPhone := ""
	if owner != nil {
		ownerName = owner.Name
		if owner.Email != "" {
			ownerEmail = owner.Email
		}
		ownerPhone = owner.PhoneNumber
	}

	// Map existing tenants by ID for easy lookup
	existingTenantsMap := make(map[string]models.TenantInfo)
	for _, tenant := range existingTenants {
		existingTenantsMap[tenant.ID] = tenant
	}

	// Process updated tenants
	for _, tenantInterface := range newTenantsRaw {
		tenantMap, ok := tenantInterface.(map[string]interface{})
		if !ok {
			continue
		}

		id, _ := tenantMap["id"].(string)
		newDueDate, _ := tenantMap["paymentDueDate"].(string)
		email, _ := tenantMap["email"].(string)
		firstName, _ := tenantMap["firstName"].(string)
		lastName, _ := tenantMap["lastName"].(string)

		// Skip if ID missing or new due date missing
		if id == "" || newDueDate == "" {
			continue
		}

		// Check if tenant exists and due date changed
		if existing, exists := existingTenantsMap[id]; exists {
			if existing.PaymentDueDate != newDueDate {
				tenantName := strings.TrimSpace(firstName + " " + lastName)
				if tenantName == "" {
					tenantName = "Tenant"
				}

				// Send email notification
				go func(tName, tEmail, dueDate string) {
					emailData := services.PaymentDueUpdateEmailData{
						TenantName:      tName,
						TenantEmail:     tEmail,
						PropertyTitle:   property.Title,
						PropertyAddress: property.Address + ", " + property.City,
						NewDueDate:      dueDate,
						OwnerName:       ownerName,
						OwnerEmail:      ownerEmail,
						OwnerPhone:      ownerPhone,
					}

					if err := h.emailService.SendPaymentDueUpdateNotification(emailData); err != nil {
						// Log error
					}
				}(tenantName, email, newDueDate)
			}
		}
	}
}
