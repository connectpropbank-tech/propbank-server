package main

import (
	"fmt"
	"log"
	"net/http"
	"shoprop-backend/config"
	"shoprop-backend/handlers"
	"strings"

	"github.com/joho/godotenv"
)

// CORS middleware function
func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-User-ID")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
}

func main() {
	// Load environment variables
	if err := godotenv.Load("config/.env"); err != nil {
		log.Println("⚠️ Warning: No .env file found, using system environment variables")
	} else {
		log.Println("✅ Environment variables loaded from .env file")
	}

	// Initialize Firebase
	config.InitFirebase()

	// Initialize handlers
	userHandler := handlers.NewUserHandler(config.GetFirestoreClient())
	authHandler := handlers.NewAuthHandler(config.GetFirestoreClient())
	propertyHandler := handlers.NewPropertyHandler(config.GetFirestoreClient())
	visitHandler := handlers.NewVisitHandler(config.GetFirestoreClient())
	serviceHandler := handlers.NewServiceHandler(config.GetFirestoreClient())

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "ShoPROP Backend API is running!", "status": "success"}`)
	})

	// Auth routes
	http.HandleFunc("/auth/user", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		authHandler.CreateOrUpdateUser(w, r)
	})

	// User routes
	http.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.URL.Path == "/users" {
			if r.Method == "GET" {
				userHandler.GetAllUsers(w, r)
			} else if r.Method == "POST" {
				userHandler.CreateUser(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		} else if strings.HasPrefix(r.URL.Path, "/users/") {
			userHandler.GetUser(w, r)
		} else {
			http.NotFound(w, r)
		}
	})

	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		userHandler.GetUser(w, r)
	})

	// Property routes
	http.HandleFunc("/properties", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			propertyHandler.CreateProperty(w, r)
		} else if r.Method == "GET" {
			// Check if this is a request for properties by owner
			if r.URL.Query().Get("ownerUID") != "" {
				propertyHandler.GetPropertiesByOwner(w, r)
			} else {
				propertyHandler.GetAllProperties(w, r)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Property search route
	http.HandleFunc("/properties/search", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			propertyHandler.SearchProperties(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/properties/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			propertyHandler.GetProperty(w, r)
		} else if r.Method == "PUT" {
			propertyHandler.UpdateProperty(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Visit routes
	http.HandleFunc("/visits", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			visitHandler.CreateVisit(w, r)
		} else if r.Method == "GET" {
			// Check if this is a request for visits by user
			if r.URL.Query().Get("userId") != "" {
				visitHandler.GetVisitsByUser(w, r)
			} else {
				http.Error(w, "User ID required for visits", http.StatusBadRequest)
			}
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/visits/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			visitHandler.GetVisit(w, r)
		} else if r.Method == "PUT" {
			visitHandler.UpdateVisit(w, r)
		} else if r.Method == "DELETE" {
			visitHandler.DeleteVisit(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Service routes
	http.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			serviceHandler.GetAllServices(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Admin route to populate services
	http.HandleFunc("/admin/populate-services", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			serviceHandler.PopulateServices(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Service request routes
	http.HandleFunc("/service-requests", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			serviceHandler.CreateServiceRequest(w, r)
		} else if r.Method == "GET" {
			serviceHandler.GetServiceRequests(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/service-requests/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "PUT" {
			serviceHandler.UpdateServiceRequestStatus(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Note: Tenant management is handled via property updates
	// Tenants are stored as arrays within property documents

	log.Println("🚀 ShoPROP Backend Server starting on :8002")
	log.Println("📍 API Endpoints:")
	log.Println("   GET  /           - Health check")
	log.Println("   POST /auth/user  - Create/Update user")
	log.Println("   POST /users      - Create user with phone number")
	log.Println("   GET  /users      - Get all users")
	log.Println("   GET  /users/{id} - Get user by ID")
	log.Println("   POST /properties - Create property")
	log.Println("   GET  /properties - Get all properties")
	log.Println("   GET  /properties?ownerUID={uid} - Get properties by owner")
	log.Println("   GET  /properties/search?q={query}&listingType={type}&projectCondition={condition} - Search properties")
	log.Println("   GET  /properties/{id} - Get property by ID")
	log.Println("   PUT  /properties/{id} - Update property")
	log.Println("   POST /visits     - Create visit")
	log.Println("   GET  /visits?userId={uid} - Get visits by user")
	log.Println("   GET  /visits/{id} - Get visit by ID")
	log.Println("   PUT  /visits/{id} - Update visit")
	log.Println("   DELETE /visits/{id} - Delete visit")
	log.Println("   GET  /services - Get all services")
	log.Println("   POST /service-requests - Create service request")
	log.Println("   GET  /service-requests?userUID={uid} - Get service requests by user")
	log.Println("   GET  /service-requests?admin=true - Get all service requests (admin)")
	log.Println("   PUT  /service-requests/{id} - Update service request status (admin)")
	log.Println("   📝 Note: Tenants managed via property updates (PUT /properties/{id})")

	if err := http.ListenAndServe(":8002", nil); err != nil {
		log.Fatalf("❌ Could not start server: %s\n", err.Error())
	}
}
