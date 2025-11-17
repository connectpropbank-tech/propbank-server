package handlers

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"shoprop-backend/config"
	"shoprop-backend/models"
	"shoprop-backend/services"
	"strings"

	"cloud.google.com/go/firestore"
)

type PropertyHandler struct {
	propertyService *services.PropertyService
	userService     *services.UserService
	imageService    *services.ImageService
}

func NewPropertyHandler(client *firestore.Client) *PropertyHandler {
	// Initialize image service with Firebase Storage
	storageClient := config.GetStorageClient()
	bucketName := "propbank-a98ed.appspot.com" // Replace with your Firebase project bucket name

	return &PropertyHandler{
		propertyService: services.NewPropertyService(client),
		userService:     services.NewUserService(client),
		imageService:    services.NewImageService(storageClient, bucketName),
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
		PlotArea:        req.PlotArea,
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
			ID:           createdProperty.ID,
			Title:        createdProperty.Title,
			Description:  createdProperty.Description,
			Price:        createdProperty.Price,
			Address:      createdProperty.Address,
			City:         createdProperty.City,
			State:        createdProperty.State,
			ZipCode:      createdProperty.ZipCode,
			PropertyType: createdProperty.PropertyType,
			Bedrooms:     createdProperty.Bedrooms,
			Bathrooms:    createdProperty.Bathrooms,
			SquareFeet:   createdProperty.SquareFeet,
			Images:       createdProperty.Images,
			OwnerUID:     createdProperty.OwnerUID,
			OwnerName:    createdProperty.OwnerName,
			OwnerEmail:   createdProperty.OwnerEmail,
			IsActive:     createdProperty.IsActive,
			CreatedAt:    createdProperty.CreatedAt,
			UpdatedAt:    createdProperty.UpdatedAt,
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
			ID:           property.ID,
			Title:        property.Title,
			Description:  property.Description,
			Price:        property.Price,
			Address:      property.Address,
			City:         property.City,
			State:        property.State,
			ZipCode:      property.ZipCode,
			PropertyType: property.PropertyType,
			Bedrooms:     property.Bedrooms,
			Bathrooms:    property.Bathrooms,
			SquareFeet:   property.SquareFeet,
			Images:       property.Images,
			OwnerUID:     property.OwnerUID,
			OwnerName:    property.OwnerName,
			OwnerEmail:   property.OwnerEmail,
			IsActive:     property.IsActive,
			CreatedAt:    property.CreatedAt,
			UpdatedAt:    property.UpdatedAt,
		})
	}

	response := map[string]interface{}{
		"success":    true,
		"properties": propertyResponses,
		"count":      len(propertyResponses),
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

	properties, err := h.propertyService.GetPropertiesByOwner(r.Context(), ownerUID)
	if err != nil {
		log.Printf("❌ Failed to get properties by owner: %v", err)
		http.Error(w, "Failed to get properties", http.StatusInternalServerError)
		return
	}

	// Convert to response format
	var propertyResponses []models.PropertyResponse
	for _, property := range properties {
		propertyResponses = append(propertyResponses, models.PropertyResponse{
			ID:           property.ID,
			Title:        property.Title,
			Description:  property.Description,
			Price:        property.Price,
			Address:      property.Address,
			City:         property.City,
			State:        property.State,
			ZipCode:      property.ZipCode,
			PropertyType: property.PropertyType,
			ListingType:  property.ListingType,
			Bedrooms:     property.Bedrooms,
			Bathrooms:    property.Bathrooms,
			SquareFeet:   property.SquareFeet,
			Images:       property.Images,
			Tenants:      property.Tenants,
			Buyers:       property.Buyers,
			OwnerUID:     property.OwnerUID,
			OwnerName:    property.OwnerName,
			OwnerEmail:   property.OwnerEmail,
			IsActive:     property.IsActive,
			CreatedAt:    property.CreatedAt,
			UpdatedAt:    property.UpdatedAt,
		})
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
		default:
			updateData[key] = value
		}
	}

	// Always update the timestamp
	updateData["updatedAt"] = existingProperty.UpdatedAt

	// Update property in database
	updatedProperty, err := h.propertyService.UpdateProperty(r.Context(), path, updateData)
	if err != nil {
		log.Printf("❌ Failed to update property: %v", err)
		http.Error(w, "Failed to update property", http.StatusInternalServerError)
		return
	}

	log.Printf("✅ Property updated successfully: %s by %s", updatedProperty.Title, updatedProperty.OwnerName)

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
			IsActive:         property.IsActive,
			CreatedAt:        property.CreatedAt,
			UpdatedAt:        property.UpdatedAt,
		})
	}

	log.Printf("✅ Found %d properties matching search criteria", len(propertyResponses))

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

	// Try to parse as float
	if cleanPrice == "" {
		return 0, fmt.Errorf("empty price")
	}

	// Simple conversion - you might want to use strconv.ParseFloat for more robust parsing
	var price float64
	if _, err := fmt.Sscanf(cleanPrice, "%f", &price); err != nil {
		return 0, err
	}

	return price, nil
}
