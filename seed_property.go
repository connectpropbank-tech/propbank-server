package main

import (
	"context"
	"fmt"
	"log"
	"shoprop-backend/config"
	"shoprop-backend/models"
	"shoprop-backend/services"
	"time"

	"github.com/joho/godotenv"
)

func main() {
	// Load env
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found")
	}

	// Initialize Firebase
	config.InitFirebase()
	client := config.GetFirestoreClient()
	propertyService := services.NewPropertyService(client)

	ctx := context.Background()

	// Create a test property
	property := models.Property{
		Title:        "Test Property for Rent Reminder",
		PropertyType: "residential",
		ListingType:  "rent",
		Address:      "123 Test Lane, Reminder City",
		OwnerUID:     "test-owner-uid-123", // Dummy owner
		Status:       "active",
		IsActive:     true,
		Tenants: []models.TenantInfo{
			{
				FirstName:      "Rains",
				LastName:       "Dwivedi",
				Email:          "rains.dwivedi98@gmail.com",
				Phone:          "9719262537",
				MonthlyRent:    "50000",
				PaymentDueDate: "17", // Today's date
				IsActive:       true,
				LeaseStartDate: "2024-01-01",
				LeaseEndDate:   "2025-12-31",
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	created, err := propertyService.CreateProperty(ctx, property)
	if err != nil {
		log.Fatalf("Failed to create property: %v", err)
	}

	fmt.Printf("Successfully created property with ID: %s\n", created.ID)
	fmt.Println("Tenant email: rains.dwivedi98@gmail.com")
	fmt.Println("Payment Due Date: 17")
}
