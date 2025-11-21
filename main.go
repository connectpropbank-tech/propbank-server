package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"shoprop-backend/config"
	"shoprop-backend/handlers"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

// CORS middleware function
func enableCORS(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, PATCH, DELETE, OPTIONS")
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
	adminNotificationHandler := handlers.NewAdminNotificationHandler(config.GetFirestoreClient())
	inspectionReportHandler := handlers.NewInspectionReportHandler(config.GetFirestoreClient())
	reviewHandler := handlers.NewReviewHandler(config.GetFirestoreClient())
	documentHandler := handlers.NewDocumentHandler(config.GetFirestoreClient())
	agreementHandler := handlers.NewAgreementHandler(config.GetFirestoreClient())

	// Create a new mux to have better control over routing
	mux := http.NewServeMux()

	// Auth routes
	mux.HandleFunc("/auth/user", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		authHandler.CreateOrUpdateUser(w, r)
	})

	// User routes
	mux.HandleFunc("/users/search", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			userHandler.SearchUserByPhone(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/users", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		userHandler.GetUser(w, r)
	})

	// Property routes
	mux.HandleFunc("/properties", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/properties/search", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/properties/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			propertyHandler.GetProperty(w, r)
		} else if r.Method == "PUT" || r.Method == "PATCH" {
			propertyHandler.UpdateProperty(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Visit routes
	mux.HandleFunc("/visits", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/visits/", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/services", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/admin/populate-services", func(w http.ResponseWriter, r *http.Request) {
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
	mux.HandleFunc("/service-requests", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/service-requests/", func(w http.ResponseWriter, r *http.Request) {
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

	// Admin notification routes
	mux.HandleFunc("/admin/notifications", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			adminNotificationHandler.CreateAdminNotification(w, r)
		} else if r.Method == "GET" {
			adminNotificationHandler.GetAdminNotifications(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/admin/notifications/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "PUT" && strings.HasSuffix(r.URL.Path, "/read") {
			adminNotificationHandler.MarkNotificationAsRead(w, r)
		} else if r.Method == "DELETE" {
			adminNotificationHandler.DeleteAdminNotification(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Inspection report routes
	mux.HandleFunc("/inspection-reports", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			inspectionReportHandler.CreateInspectionReport(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/inspection-reports/property/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			inspectionReportHandler.GetInspectionReportsByProperty(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/inspection-reports/user/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			inspectionReportHandler.GetInspectionReportsByUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/inspection-reports/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			inspectionReportHandler.GetInspectionReportByID(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Review routes
	mux.HandleFunc("/reviews", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			reviewHandler.CreateReview(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/reviews/property/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			reviewHandler.GetReviewsByProperty(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/reviews/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			reviewHandler.GetReviewByID(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Document routes
	mux.HandleFunc("/documents", func(w http.ResponseWriter, r *http.Request) {
		// Add panic recovery
		defer func() {
			if err := recover(); err != nil {
				log.Printf("❌ PANIC in /documents handler: %v", err)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(map[string]interface{}{
					"success": false,
					"message": "Internal server error: " + fmt.Sprintf("%v", err),
				})
			}
		}()

		log.Printf("📄 /documents route hit: Method=%s, Path=%s", r.Method, r.URL.Path)
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			documentHandler.CreateDocument(w, r)
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusMethodNotAllowed)
			json.NewEncoder(w).Encode(map[string]interface{}{
				"success": false,
				"message": "Method not allowed",
			})
		}
	})

	mux.HandleFunc("/documents/property/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			documentHandler.GetDocumentsByProperty(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/documents/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			documentHandler.GetDocumentByID(w, r)
		} else if r.Method == "DELETE" {
			documentHandler.DeleteDocument(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Agreement routes - Register BEFORE root route to ensure proper matching
	mux.HandleFunc("/agreements/renew", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔵 /agreements/renew route hit: Method=%s", r.Method)
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			agreementHandler.RenewAgreement(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	mux.HandleFunc("/agreements/terminate", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🔴 /agreements/terminate route MATCHED: Method=%s, Path=%s, URL=%s", r.Method, r.URL.Path, r.URL.String())
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			log.Printf("🔴 OPTIONS request handled")
			return
		}
		if r.Method == "POST" {
			log.Printf("🔴 Calling TerminateAgreement handler")
			agreementHandler.TerminateAgreement(w, r)
		} else {
			log.Printf("🔴 Method not allowed: %s", r.Method)
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Root route - Register LAST and use exact match only
	// Use a custom handler that only matches exact "/" path
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🟢 Root route handler called: Method=%s, Path=%s", r.Method, r.URL.Path)

		// CRITICAL: Only handle exact root path, return 404 for everything else
		// This prevents the root route from catching other paths
		if r.URL.Path != "/" {
			log.Printf("🟡 Root route rejecting path: %s (not exact match)", r.URL.Path)
			http.NotFound(w, r)
			return
		}

		log.Printf("🟢 Root route processing exact match")
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "ShoPROP Backend API is running!", "status": "success"}`)
	})

	// Note: Tenant management is handled via property updates
	// Tenants are stored as arrays within property documents

	log.Println("🚀 PropBank Backend Server starting on :8002")
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
	log.Println("   PUT/PATCH  /properties/{id} - Update property")
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
	log.Println("   POST /admin/notifications - Create admin notification")
	log.Println("   GET  /admin/notifications - Get all admin notifications")
	log.Println("   GET  /admin/notifications?unread=true - Get unread admin notifications")
	log.Println("   PUT  /admin/notifications/{id}/read - Mark notification as read")
	log.Println("   DELETE /admin/notifications/{id} - Delete admin notification")
	log.Println("   POST /agreements/renew - Renew agreement")
	log.Println("   POST /agreements/terminate - Terminate agreement")
	log.Println("   POST /documents - Upload document")
	log.Println("   GET  /documents/property/{id} - Get documents by property")
	log.Println("   POST /inspection-reports - Create inspection report")
	log.Println("   POST /reviews - Create review")
	log.Println("   📝 Note: Tenants managed via property updates (PUT /properties/{id})")

	// Wrap mux with CORS middleware and logging
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("🌐 Incoming request: Method=%s, Path=%s, URL=%s, Content-Length=%s",
			r.Method, r.URL.Path, r.URL.String(), r.Header.Get("Content-Length"))
		enableCORS(w, r)
		mux.ServeHTTP(w, r)
	})

	// Create server with longer timeouts for large file uploads (base64 encoded files can be large)
	server := &http.Server{
		Addr:           ":8002",
		Handler:        handler,
		ReadTimeout:    60 * time.Second, // 60 seconds to read request (for large uploads)
		WriteTimeout:   60 * time.Second, // 60 seconds to write response
		MaxHeaderBytes: 1 << 20,          // 1MB header limit
	}

	log.Println("✅ Server configured with 60s timeouts for large uploads")
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("❌ Could not start server: %s\n", err.Error())
	}
}
