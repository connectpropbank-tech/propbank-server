package services

import (
	"context"
	"fmt"
	"log"
	"math/rand"
	"shoprop-backend/models"
	"strings"
	"time"

	"cloud.google.com/go/firestore"
	"google.golang.org/api/iterator"
)

type PropertyService struct {
	client *firestore.Client
}

func NewPropertyService(client *firestore.Client) *PropertyService {
	return &PropertyService{
		client: client,
	}
}

// GenerateID generates a new document ID for property in format: 4 capitals + 4 numbers (e.g., ABCD1234)
func (s *PropertyService) GenerateID() string {
	const letters = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	const numbers = "0123456789"

	// Create a new random source
	source := rand.NewSource(time.Now().UnixNano())
	rng := rand.New(source)

	// Generate unique ID (retry if duplicate exists)
	for attempts := 0; attempts < 10; attempts++ {
		// Generate 4 random capital letters
		var result strings.Builder
		for i := 0; i < 4; i++ {
			result.WriteByte(letters[rng.Intn(len(letters))])
		}

		// Generate 4 random numbers
		for i := 0; i < 4; i++ {
			result.WriteByte(numbers[rng.Intn(len(numbers))])
		}

		propertyID := result.String()

		// Check if this ID already exists
		ctx := context.Background()
		_, err := s.client.Collection("properties").Doc(propertyID).Get(ctx)
		if err != nil {
			// Document doesn't exist, so this ID is unique
			return propertyID
		}
		// If we reach here, the ID exists, so try again
	}

	// Fallback to Firestore auto-generated ID if we can't generate unique custom ID
	return s.client.Collection("properties").NewDoc().ID
}

func (s *PropertyService) CreateProperty(ctx context.Context, property models.Property) (*models.Property, error) {
	var docRef *firestore.DocumentRef

	// Use existing ID if provided, otherwise generate new one
	if property.ID != "" {
		docRef = s.client.Collection("properties").Doc(property.ID)
	} else {
		docRef = s.client.Collection("properties").NewDoc()
		property.ID = docRef.ID
	}

	property.CreatedAt = time.Now()
	property.UpdatedAt = time.Now()
	
	// Set default status to "active" if not provided
	if property.Status == "" {
		property.Status = "active"
	}
	
	// Set default isActive to true if not explicitly set (keep independent from status)
	// Note: status and isActive are independent fields
	if !property.IsActive {
		property.IsActive = true // Default to true, but keep independent from status
	}

	// Set the property data
	_, err := docRef.Set(ctx, property)
	if err != nil {
		return nil, fmt.Errorf("failed to create property: %v", err)
	}

	return &property, nil
}

func (s *PropertyService) GetPropertyByID(ctx context.Context, id string) (*models.Property, error) {
	doc, err := s.client.Collection("properties").Doc(id).Get(ctx)
	if err != nil {
		return nil, fmt.Errorf("property not found: %v", err)
	}

	var property models.Property
	if err := doc.DataTo(&property); err != nil {
		return nil, fmt.Errorf("failed to parse property data: %v", err)
	}

	// Set ID from document reference
	property.ID = doc.Ref.ID

	// Set default status if not set (for backward compatibility)
	if property.Status == "" {
		if property.IsActive {
			property.Status = "active"
		} else {
			property.Status = "inactive"
		}
	}

	return &property, nil
}

func (s *PropertyService) GetAllProperties(ctx context.Context) ([]models.Property, error) {
	var properties []models.Property

	iter := s.client.Collection("properties").Where("isActive", "==", true).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate properties: %v", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			continue // Skip invalid documents
		}

		// Set ID from document reference
		property.ID = doc.Ref.ID

		// Set default status if not set (for backward compatibility)
		if property.Status == "" {
			if property.IsActive {
				property.Status = "active"
			} else {
				property.Status = "inactive"
			}
		}

		properties = append(properties, property)
	}

	return properties, nil
}

func (s *PropertyService) GetPropertiesByOwner(ctx context.Context, ownerUID string) ([]models.Property, error) {
	var properties []models.Property

	// Query ONLY by ownerUID - return ALL properties for this owner (regardless of status)
	// Client-side will filter based on status field
	iter := s.client.Collection("properties").Where("ownerUID", "==", ownerUID).Documents(ctx)
	defer iter.Stop()

	log.Printf("🔍 Querying Firestore for ALL properties with ownerUID: %s", ownerUID)
	
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("❌ Error iterating properties: %v", err)
			return nil, fmt.Errorf("failed to iterate properties: %v", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			log.Printf("⚠️  Skipping property document %s due to conversion error: %v", doc.Ref.ID, err)
			continue // Skip invalid documents
		}
		
		// Set ID from document reference
		property.ID = doc.Ref.ID
		
		// Set default status to "active" if not set (for backward compatibility)
		if property.Status == "" {
			if property.IsActive {
				property.Status = "active"
			} else {
				property.Status = "inactive"
			}
		}
		
		// Include ALL properties regardless of status
		properties = append(properties, property)
		log.Printf("✅ Added property: ID=%s, Title=%s, Status=%s", property.ID, property.Title, property.Status)
	}
	
	log.Printf("📊 Found %d total properties for ownerUID: %s", len(properties), ownerUID)
	
	return properties, nil
}

// GetPropertiesByTenantUID gets all properties where the user is a tenant (via tenants array)
func (s *PropertyService) GetPropertiesByTenantUID(ctx context.Context, tenantUID string) ([]models.Property, error) {
	var properties []models.Property

	// Since Firestore doesn't support querying nested array fields directly,
	// we need to fetch all properties and filter in memory
	// For better performance, we could maintain a separate index, but for now this will work
	// Fetch ALL properties (similar to GetPropertiesByOwner) - client will filter by status
	iter := s.client.Collection("properties").Documents(ctx)
	defer iter.Stop()

	log.Printf("🔍 Querying Firestore for properties where user is a tenant (tenantUID: %s)", tenantUID)

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			log.Printf("❌ Error iterating properties: %v", err)
			return nil, fmt.Errorf("failed to iterate properties: %v", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			log.Printf("⚠️  Skipping property document %s due to conversion error: %v", doc.Ref.ID, err)
			continue
		}

		// Set ID from document reference
		property.ID = doc.Ref.ID

		// Check if any tenant has a matching userUID
		isTenant := false
		for _, tenant := range property.Tenants {
			if tenant.UserUID == tenantUID && tenant.IsActive {
				isTenant = true
				break
			}
		}

		// Only include if user is a tenant of this property
		if isTenant {
			// Set default status if not set (for backward compatibility)
			if property.Status == "" {
				if property.IsActive {
					property.Status = "active"
				} else {
					property.Status = "inactive"
				}
			}
			properties = append(properties, property)
			log.Printf("✅ Added tenant property: ID=%s, Title=%s, Status=%s", property.ID, property.Title, property.Status)
		}
	}

	log.Printf("📊 Found %d properties where user is a tenant (tenantUID: %s)", len(properties), tenantUID)

	return properties, nil
}

func (s *PropertyService) GetArchivedPropertiesByOwner(ctx context.Context, ownerUID string) ([]models.Property, error) {
	var properties []models.Property

	// Query for all properties by ownerUID, then filter for archived ones (isActive == false)
	iter := s.client.Collection("properties").Where("ownerUID", "==", ownerUID).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate archived properties: %v", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			continue // Skip invalid documents
		}
		
		// Filter for archived properties (isActive == false)
		// Properties with isActive explicitly set to false should appear in archived
		// Properties without isActive field or with isActive=true should not appear
		if property.IsActive == false {
			properties = append(properties, property)
		}
	}

	log.Printf("🔍 GetArchivedPropertiesByOwner: Found %d archived properties for ownerUID: %s", len(properties), ownerUID)
	return properties, nil
}

func (s *PropertyService) GetPropertiesByListingType(ctx context.Context, listingType string) ([]models.Property, error) {
	var properties []models.Property

	iter := s.client.Collection("properties").Where("listingType", "==", listingType).Where("isActive", "==", true).Documents(ctx)
	defer iter.Stop()

	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate properties: %v", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			continue // Skip invalid documents
		}

		// Set ID from document reference
		property.ID = doc.Ref.ID

		// Set default status if not set (for backward compatibility)
		if property.Status == "" {
			if property.IsActive {
				property.Status = "active"
			} else {
				property.Status = "inactive"
			}
		}

		properties = append(properties, property)
	}

	return properties, nil
}

func (s *PropertyService) UpdateProperty(ctx context.Context, id string, updates map[string]interface{}) (*models.Property, error) {
	// Handle special processing for tenants and buyers
	if tenants, hasTenants := updates["tenants"]; hasTenants {
		if tenantsSlice, ok := tenants.([]interface{}); ok {
			for i, tenant := range tenantsSlice {
				if tenantMap, ok := tenant.(map[string]interface{}); ok {
					// Set timestamps for new tenant entries that don't have them
					if _, hasCreated := tenantMap["createdAt"]; !hasCreated {
						tenantMap["createdAt"] = time.Now()
					}
					if _, hasUpdated := tenantMap["updatedAt"]; !hasUpdated {
						tenantMap["updatedAt"] = time.Now()
					}
					tenantsSlice[i] = tenantMap
				}
			}
			updates["tenants"] = tenantsSlice
		}
	}

	if buyers, hasBuyers := updates["buyers"]; hasBuyers {
		if buyersSlice, ok := buyers.([]interface{}); ok {
			for i, buyer := range buyersSlice {
				if buyerMap, ok := buyer.(map[string]interface{}); ok {
					// Set timestamps for new buyer entries that don't have them
					if _, hasCreated := buyerMap["createdAt"]; !hasCreated {
						buyerMap["createdAt"] = time.Now()
					}
					if _, hasUpdated := buyerMap["updatedAt"]; !hasUpdated {
						buyerMap["updatedAt"] = time.Now()
					}
					buyersSlice[i] = buyerMap
				}
			}
			updates["buyers"] = buyersSlice
		}
	}

	updates["updatedAt"] = time.Now()

	// Convert updates map to Firestore updates
	var firestoreUpdates []firestore.Update
	for path, value := range updates {
		firestoreUpdates = append(firestoreUpdates, firestore.Update{
			Path:  path,
			Value: value,
		})
	}

	_, err := s.client.Collection("properties").Doc(id).Update(ctx, firestoreUpdates)
	if err != nil {
		return nil, fmt.Errorf("failed to update property: %v", err)
	}

	// Get the updated property
	return s.GetPropertyByID(ctx, id)
}

func (s *PropertyService) DeleteProperty(ctx context.Context, id string, ownerUID string) error {
	// Verify ownership before deletion
	property, err := s.GetPropertyByID(ctx, id)
	if err != nil {
		return err
	}

	if property.OwnerUID != ownerUID {
		return fmt.Errorf("unauthorized: you can only delete your own properties")
	}

	// Soft delete by setting isActive to false
	_, err = s.client.Collection("properties").Doc(id).Update(ctx, []firestore.Update{
		{Path: "isActive", Value: false},
		{Path: "updatedAt", Value: time.Now()},
	})
	if err != nil {
		return fmt.Errorf("failed to delete property: %v", err)
	}

	return nil
}

// SearchProperties searches properties based on query string, listing type, project condition, and filters
func (s *PropertyService) SearchProperties(ctx context.Context, query string, listingType string, projectCondition string) ([]models.Property, error) {
	var properties []models.Property

	// Start with base query for active properties
	collection := s.client.Collection("properties")
	queryRef := collection.Where("isActive", "==", true)

	// Add listing type filter if provided
	if listingType != "" {
		queryRef = queryRef.Where("listingType", "==", listingType)
	}

	// Add project condition filter if provided
	if projectCondition != "" {
		queryRef = queryRef.Where("projectCondition", "==", projectCondition)
	}

	// Get all matching properties first
	iter := queryRef.Documents(ctx)
	defer iter.Stop()

	// If no search query provided, return all properties with listing type filter
	if query == "" {
		for {
			doc, err := iter.Next()
			if err == iterator.Done {
				break
			}
			if err != nil {
				return nil, fmt.Errorf("failed to iterate properties: %v", err)
			}

			var property models.Property
			if err := doc.DataTo(&property); err != nil {
				continue // Skip invalid documents
			}
			properties = append(properties, property)
		}
		return properties, nil
	}

	// Convert query to lowercase for case-insensitive search
	searchQuery := strings.ToLower(strings.TrimSpace(query))

	// Search through properties and filter based on query
	for {
		doc, err := iter.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			return nil, fmt.Errorf("failed to iterate properties: %v", err)
		}

		var property models.Property
		if err := doc.DataTo(&property); err != nil {
			continue // Skip invalid documents
		}

		// Search in multiple fields (case-insensitive)
		matches := false

		// Search in location fields
		if (property.Location != "" && strings.Contains(strings.ToLower(property.Location), searchQuery)) ||
			(property.Address != "" && strings.Contains(strings.ToLower(property.Address), searchQuery)) ||
			(property.City != "" && strings.Contains(strings.ToLower(property.City), searchQuery)) ||
			(property.State != "" && strings.Contains(strings.ToLower(property.State), searchQuery)) {
			matches = true
		}

		// Search in property details
		if (property.Title != "" && strings.Contains(strings.ToLower(property.Title), searchQuery)) ||
			(property.PropertyType != "" && strings.Contains(strings.ToLower(property.PropertyType), searchQuery)) ||
			(property.Configuration != "" && strings.Contains(strings.ToLower(property.Configuration), searchQuery)) {
			matches = true
		}

		// Search in specific features like "2bhk", "3bhk", "1rk", etc.
		configLower := ""
		if property.Configuration != "" {
			configLower = strings.ToLower(property.Configuration)
		}
		if (configLower != "" && strings.Contains(configLower, searchQuery)) ||
			(strings.Contains(searchQuery, "bhk") && strings.Contains(configLower, "bhk")) ||
			(strings.Contains(searchQuery, "rk") && strings.Contains(configLower, "rk")) {
			matches = true
		}

		// Search by unit details
		if (property.UnitNumber != "" && strings.Contains(strings.ToLower(property.UnitNumber), searchQuery)) ||
			(property.Floor != "" && strings.Contains(strings.ToLower(property.Floor), searchQuery)) {
			matches = true
		}

		if matches {
			properties = append(properties, property)
		}
	}

	return properties, nil
}
