package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
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
}

func NewPropertyHandler(client *firestore.Client) *PropertyHandler {
	// Initialize image service with Firebase Storage
	storageClient := config.GetStorageClient()
	bucketName := config.GetStorageBucket()

	return &PropertyHandler{
		propertyService:          services.NewPropertyService(client),
		userService:              services.NewUserService(client),
		imageService:             services.NewImageService(storageClient, bucketName),
		adminNotificationService: services.NewAdminNotificationService(client),
		siteSettingsService:      services.NewSiteSettingsService(client),
	}
}

func (h *PropertyHandler) CreateProperty(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var req models.CreatePropertyRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
		return
	}

	// Validate required fields
	if req.PropertyTitle == "" || req.PropertyType == "" || req.Location == "" || req.OwnerUID == "" {
		http.Error(w, "Missing required fields: propertyTitle, propertyType, location, ownerUID", http.StatusBadRequest)
		return
	}

	// Get owner information
	owner, err := h.userService.GetUserByID(r.Context(), req.OwnerUID)
	if err != nil {
		log.Printf("❌ Owner not found: %v", err)
		http.Error(w, "Owner not found", http.StatusBadRequest)
		return
	}

	// Generate unique property ID for image storage
	propertyID := h.propertyService.GenerateID()

	// Upload images to Firebase Storage
	imageURLs := []string{}
	if len(req.Images) > 0 {
		log.Printf("📸 Uploading %d images to Firebase Storage...", len(req.Images))
		uploadedURLs, err := h.imageService.UploadPropertyImages(r.Context(), req.Images, propertyID)
		if err != nil {
			log.Printf("⚠️  Warning: Failed to upload some images: %v", err)
		}
		imageURLs = uploadedURLs
		log.Printf("✅ Successfully uploaded %d images", len(imageURLs))
	}

	// Create property object with comprehensive fields
	property := models.Property{
		ID: propertyID, // Set the generated ID
		// Basic Property Details
		Title:         req.PropertyTitle,
		PropertyType:  req.PropertyType,
		Configuration: req.Configuration,
		ListingType:   req.ListingType,

		// Unit Details
		UnitNumber: req.UnitNumber,
		Floor:      req.Floor,
		Location:   req.Location,
		Address:    req.Location, // Using location as address for compatibility

		// Area Details
		CarpetArea:      req.CarpetArea,
		ConstructedArea: req.ConstructedArea,

		// Tenant Information
		TenantName:   req.TenantName,
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
		RentalStatus:          req.RentalStatus,
		FurnishedChecklist:    req.FurnishedChecklist,

		// Images & Comments
		Images:           imageURLs,
		SpecificComments: req.SpecificComments,

		// Owner Info
		OwnerUID:   req.OwnerUID,
		OwnerName:  owner.Name,
		OwnerEmail: owner.Email,

		// Status - set default to "active" if not provided
		Status: "active", // Default status for new properties
	}

	// Create property in database
	createdProperty, err := h.propertyService.CreateProperty(r.Context(), property)
	if err != nil {
		log.Printf("❌ Failed to create property: %v", err)
		http.Error(w, "Failed to create property", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Property created successfully: %s by %s", createdProperty.Title, createdProperty.OwnerName)

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

	var properties []models.Property
	var err error

	if listingType != "" {
		properties, err = h.propertyService.GetPropertiesByListingType(r.Context(), listingType)
		log.Printf("📍 Filtering properties by listing type: %s, found %d properties", listingType, len(properties))
	} else {
		properties, err = h.propertyService.GetAllProperties(r.Context())
		log.Printf("📍 Getting all properties, found %d properties", len(properties))
	}

	if err != nil {
		log.Printf("❌ Failed to get properties: %v", err)
		http.Error(w, "Failed to get properties", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var propertyResponses []models.PropertyResponse
	for _, property := range properties {
		propertyResponses = append(propertyResponses, models.PropertyResponse{
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
		})
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
		log.Printf("❌ Property not found: %v", err)
		http.Error(w, "Property not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":  true,
		"property": property, // Return full property object with all fields
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
		log.Printf("❌ Failed to get properties by owner: %v", err)
		http.Error(w, "Failed to get properties", http.StatusInternalServerError)
		return
	}

	// Also get properties where the user is a tenant
	tenantProperties, err := h.propertyService.GetPropertiesByTenantUID(r.Context(), ownerUID)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to get properties by tenant: %v (continuing with owned properties only)", err)
		tenantProperties = []models.Property{} // Continue with empty tenant properties if error
	}

	log.Printf("📦 Found %d owned properties and %d tenant properties for userUID: %s", len(ownedProperties), len(tenantProperties), ownerUID)

	// Create a map to track property IDs to avoid duplicates (in case a user is both owner and tenant)
	propertyMap := make(map[string]*models.PropertyResponse)

	// Convert owned properties to response format with "owner" role
	for _, property := range ownedProperties {
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
			UserRole:              "owner", // Mark as owner
		}
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

	log.Printf("✅ Sent %d properties (%d owned, %d tenant) to client for userUID: %s", len(propertyResponses), len(ownedProperties), len(tenantProperties), ownerUID)
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
		log.Printf("❌ Failed to get archived properties by owner: %v", err)
		http.Error(w, "Failed to get archived properties", http.StatusInternalServerError)
		return
	}

	log.Printf("🔍 Found %d archived properties for ownerUID: %s", len(properties), ownerUID)

	// Convert to response format
	var propertyResponses []models.PropertyResponse
	for _, property := range properties {
		propertyResponses = append(propertyResponses, models.PropertyResponse{
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
		})
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":    true,
		"properties": propertyResponses,
		"count":      len(propertyResponses),
		"ownerUID":   ownerUID,
	})

	log.Printf("📍 Retrieved %d archived properties for ownerUID: %s", len(propertyResponses), ownerUID)
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
		log.Printf("❌ Property not found: %v", err)
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
			log.Printf("📸 Uploading %d new images to Firebase Storage...", len(imageStrings))
			uploadedURLs, err := h.imageService.UploadPropertyImages(r.Context(), imageStrings, path)
			if err != nil {
				log.Printf("⚠️  Warning: Failed to upload some images: %v", err)
			} else {
				imageURLs = uploadedURLs
				log.Printf("✅ Successfully uploaded %d images", len(imageURLs))
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
					log.Printf("📝 Setting status to: '%s' for property %s (isActive unchanged)", statusStr, path)
				} else {
					log.Printf("⚠️  Invalid status value: '%s' (must be 'active' or 'inactive'), skipping", statusStr)
				}
			} else {
				updateData[key] = value
				log.Printf("📝 Using status value as-is: %v (type: %T) for property %s", value, value, path)
			}
		case "isActive":
			// Ensure isActive is properly converted to boolean
			// Keep status and isActive independent - do NOT sync
			if boolVal, ok := value.(bool); ok {
				updateData[key] = boolVal
				log.Printf("📝 Setting isActive to: %v for property %s (status unchanged)", boolVal, path)
			} else if strVal, ok := value.(string); ok {
				// Handle string "true"/"false" from JSON
				boolVal := (strVal == "true")
				updateData[key] = boolVal
				log.Printf("📝 Converting isActive string '%s' to bool: %v for property %s (status unchanged)", strVal, boolVal, path)
			} else {
				updateData[key] = value
				log.Printf("📝 Using isActive value as-is: %v (type: %T) for property %s", value, value, path)
			}
		default:
			updateData[key] = value
		}
	}

	// Always update the timestamp
	updateData["updatedAt"] = time.Now()

	// Check if wantToSell is being toggled ON before updating (we need existing property for notification)
	var wantToSellBeingSetToTrue bool = false
	if wantToSell, hasWantToSell := updateData["wantToSell"]; hasWantToSell {
		if wantToSellBool, ok := wantToSell.(bool); ok && wantToSellBool {
			// Check if it's actually changing from false to true
			if !existingProperty.WantToSell {
				wantToSellBeingSetToTrue = true
			}
		}
	}

	// Update property in database
	updatedProperty, err := h.propertyService.UpdateProperty(r.Context(), path, updateData)
	if err != nil {
		log.Printf("❌ Failed to update property: %v", err)
		http.Error(w, "Failed to update property", http.StatusInternalServerError)
		return
	}

	// If tenants were updated, map property information to users with userUID
	if tenants, hasTenants := updateData["tenants"]; hasTenants {
		if tenantsArray, ok := tenants.([]interface{}); ok {
			h.updateUsersWithRentedProperty(r.Context(), path, updatedProperty.OwnerUID, updatedProperty.OwnerName, tenantsArray)
		}
	}

	// Handle "Want to Sell" toggle - create notification if toggled ON
	if wantToSellBeingSetToTrue {
		h.createWantToSellNotification(r.Context(), updatedProperty)
	}

	// Log the isActive status after update
	log.Printf("✅ Property updated successfully: %s by %s (isActive: %v)", updatedProperty.Title, updatedProperty.OwnerName, updatedProperty.IsActive)

	// Return success response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success": true,
		"message": "Property updated successfully",
		"property": models.PropertyResponse{
			ID:           updatedProperty.ID,
			Title:        updatedProperty.Title,
			Description:  updatedProperty.Description,
			Price:        updatedProperty.Price,
			Address:      updatedProperty.Address,
			City:         updatedProperty.City,
			State:        updatedProperty.State,
			ZipCode:      updatedProperty.ZipCode,
			PropertyType: updatedProperty.PropertyType,
			ListingType:  updatedProperty.ListingType,
			Bedrooms:     updatedProperty.Bedrooms,
			Bathrooms:    updatedProperty.Bathrooms,
			SquareFeet:   updatedProperty.SquareFeet,
			Images:       updatedProperty.Images,
			Tenants:      updatedProperty.Tenants,
			Buyers:       updatedProperty.Buyers,
			OwnerUID:     updatedProperty.OwnerUID,
			OwnerName:    updatedProperty.OwnerName,
			OwnerEmail:   updatedProperty.OwnerEmail,
			IsActive:     updatedProperty.IsActive,
			CreatedAt:    updatedProperty.CreatedAt,
			UpdatedAt:    updatedProperty.UpdatedAt,
		},
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

	log.Printf("🔍 Searching properties with query: '%s', listingType: '%s', projectCondition: '%s'", query, listingType, projectCondition)

	// Search properties
	properties, err := h.propertyService.SearchProperties(r.Context(), query, listingType, projectCondition)
	if err != nil {
		log.Printf("❌ Failed to search properties: %v", err)
		http.Error(w, "Failed to search properties", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var propertyResponses []models.PropertyResponse
	for _, property := range properties {
		// Calculate price based on listing type for display
		var displayPrice float64
		if property.ListingType == "rent" && property.MonthlyRent != "" {
			// Try to parse monthly rent as float
			if parsedRent, err := parsePrice(property.MonthlyRent); err == nil {
				displayPrice = parsedRent
			}
		} else if property.ListingType == "sell" && property.SellingPrice != "" {
			// Try to parse selling price as float
			if parsedPrice, err := parsePrice(property.SellingPrice); err == nil {
				displayPrice = parsedPrice
			}
		}

		propertyResponses = append(propertyResponses, models.PropertyResponse{
			ID:               property.ID,
			Title:            property.Title,
			Description:      property.Description,
			Price:            displayPrice,
			Address:          property.Address,
			City:             property.City,
			State:            property.State,
			ZipCode:          property.ZipCode,
			PropertyType:     property.PropertyType,
			ListingType:      property.ListingType,
			ProjectCondition: property.ProjectCondition,
			Bedrooms:         property.Bedrooms,
			Bathrooms:        property.Bathrooms,
			SquareFeet:       property.SquareFeet,
			Images:           property.Images,
			Tenants:          property.Tenants,
			Buyers:           property.Buyers,
			OwnerUID:         property.OwnerUID,
			OwnerName:        property.OwnerName,
			OwnerEmail:       property.OwnerEmail,
			Status:           property.Status,
			IsActive:         property.IsActive,
			CreatedAt:        property.CreatedAt,
			UpdatedAt:        property.UpdatedAt,
			RentalStatus:     property.RentalStatus,
		})
	}

	log.Printf("Found %d properties matching search criteria", len(propertyResponses))

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
			log.Printf("⚠️  Warning: Failed to get user %s for property mapping: %v", userUID, err)
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
			log.Printf("⚠️  Warning: Failed to update user %s with property info: %v", userUID, err)
		} else {
			log.Printf("✅ User %s updated with rented property: %s (Owner: %s)", userUID, propertyID, ownerName)
		}
	}
}

// createWantToSellNotification creates an admin notification when owner toggles "Want to Sell" ON
func (h *PropertyHandler) createWantToSellNotification(ctx context.Context, property *models.Property) {
	// Get owner details to include phone number
	owner, err := h.userService.GetUserByID(ctx, property.OwnerUID)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to get owner details for notification: %v", err)
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
		Timestamp:  time.Now().Format(time.RFC3339),
		IsRead:     false,
		Priority:   "high",
	}

	// Create notification
	_, err = h.adminNotificationService.CreateNotification(ctx, notificationReq)
	if err != nil {
		log.Printf("⚠️  Warning: Failed to create admin notification for want to sell: %v", err)
	} else {
		log.Printf("✅ Admin notification created for property %s (Owner wants to sell)", property.ID)
	}
}
