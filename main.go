package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"shoprop-backend/config"
	"shoprop-backend/handlers"
	"shoprop-backend/services"
	"strings"

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
	} else {
	}

	// Initialize Firebase
	config.InitFirebase()

	// Initialize upload handler (Cloudflare R2)
	uploadHandler, err := handlers.NewUploadHandler()
	if err != nil {
		log.Fatalf("Failed to initialize upload handler: %v", err)
	}

	// Initialize handlers
	userHandler := handlers.NewUserHandler(config.GetFirestoreClient())
	authHandler := handlers.NewAuthHandler(config.GetFirestoreClient())
	propertyHandler := handlers.NewPropertyHandler(config.GetFirestoreClient())
	visitHandler := handlers.NewVisitHandler(config.GetFirestoreClient())
	serviceHandler := handlers.NewServiceHandler(config.GetFirestoreClient(), uploadHandler.GetR2Service())
	// Initialize email service and reminder scheduler
	emailService := services.NewEmailService()

	adminNotificationHandler := handlers.NewAdminNotificationHandler(config.GetFirestoreClient(), emailService)
	siteSettingsHandler := handlers.NewSiteSettingsHandler(config.GetFirestoreClient())
	agreementHandler := handlers.NewAgreementHandler(config.GetFirestoreClient())
	documentHandler := handlers.NewDocumentHandler(config.GetFirestoreClient(), uploadHandler.GetR2Service())
	if err != nil {
	} else {
	}

	// Initialize email service and reminder scheduler
	visitService := services.NewVisitService(config.GetFirestoreClient())
	userService := services.NewUserService(config.GetFirestoreClient())
	propertyService := services.NewPropertyService(config.GetFirestoreClient())
	reminderScheduler := services.NewReminderScheduler(emailService, visitService, userService, propertyService)

	// Start the reminder scheduler (checks every minute)
	reminderScheduler.Start()

	// Admin route to force trigger reminders (for testing)
	http.HandleFunc("/admin/trigger-reminders", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}

		// Run checks in background to not block response
		go reminderScheduler.ForceCheck()

		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"success": true, "message": "Reminder checks triggered successfully"}`)
	})

	// Routes
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"message": "ShoPROP Backend API is running! (v2)", "status": "success"}`)
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

	// User search route - must be before /users/ to match correctly
	http.HandleFunc("/users/search", func(w http.ResponseWriter, r *http.Request) {
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

	// Property search route - use trailing slash to ensure proper matching
	http.HandleFunc("/properties/search/", func(w http.ResponseWriter, r *http.Request) {
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

	// Get properties where user is a tenant - use trailing slash to ensure proper matching
	http.HandleFunc("/properties/tenant/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			propertyHandler.GetPropertiesByTenant(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/properties/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}

		// Skip if this is the /properties/tenant or /properties/search route
		// These are handled by their own handlers (ServeMux matches longest pattern)
		// But keeping this check for safety if needed, or allowing fallthrough
		path := r.URL.Path
		if strings.HasPrefix(path, "/properties/tenant") || strings.HasPrefix(path, "/properties/search") {
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

	// Document routes
	http.HandleFunc("/documents", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			documentHandler.CreateDocument(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Document routes with ID or Property ID
	http.HandleFunc("/documents/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}

		path := r.URL.Path
		if strings.HasPrefix(path, "/documents/property/") {
			if r.Method == "GET" {
				documentHandler.GetDocumentsByProperty(w, r)
				return
			}
		}

		if r.Method == "GET" {
			documentHandler.GetDocumentByID(w, r)
		} else if r.Method == "DELETE" {
			documentHandler.DeleteDocument(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// User By ID Routes (GET, PUT, PATCH)
	http.HandleFunc("/users/", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}

		if r.Method == "GET" {
			userHandler.GetUser(w, r)
		} else if r.Method == "PUT" || r.Method == "PATCH" {
			userHandler.UpdateUser(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Agreement routes - must be registered before other routes that might match
	http.HandleFunc("/agreements/terminate", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			agreementHandler.TerminateAgreement(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/agreements/request-termination", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			agreementHandler.RequestTermination(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/agreements/renew", func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/agreements/request-renewal", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "POST" {
			agreementHandler.RequestRenewal(w, r)
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

	// Admin notification routes
	http.HandleFunc("/admin/notifications", func(w http.ResponseWriter, r *http.Request) {
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

	http.HandleFunc("/admin/notifications/", func(w http.ResponseWriter, r *http.Request) {
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

	// Site Settings routes (for admin to update dynamic content)
	http.HandleFunc("/admin/site-settings", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "GET" {
			siteSettingsHandler.GetSiteSettings(w, r)
		} else if r.Method == "PUT" || r.Method == "PATCH" {
			siteSettingsHandler.UpdateSiteSettings(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	http.HandleFunc("/admin/site-settings/quote", func(w http.ResponseWriter, r *http.Request) {
		enableCORS(w, r)
		if r.Method == "OPTIONS" {
			return
		}
		if r.Method == "PUT" || r.Method == "PATCH" {
			siteSettingsHandler.UpdateQuote(w, r)
		} else {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Tenants are stored as arrays within property documents

	// Upload routes (Cloudflare R2)
	if uploadHandler != nil {
		http.HandleFunc("/upload", func(w http.ResponseWriter, r *http.Request) {
			enableCORS(w, r)
			if r.Method == "OPTIONS" {
				return
			}
			if r.Method == "POST" {
				uploadHandler.UploadImage(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		http.HandleFunc("/upload/image", func(w http.ResponseWriter, r *http.Request) {
			enableCORS(w, r)
			if r.Method == "OPTIONS" {
				return
			}
			if r.Method == "POST" {
				uploadHandler.UploadBase64Image(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		http.HandleFunc("/upload/document", func(w http.ResponseWriter, r *http.Request) {
			enableCORS(w, r)
			if r.Method == "OPTIONS" {
				return
			}
			if r.Method == "POST" {
				uploadHandler.UploadDocument(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		http.HandleFunc("/upload/property-images", func(w http.ResponseWriter, r *http.Request) {
			enableCORS(w, r)
			if r.Method == "OPTIONS" {
				return
			}
			if r.Method == "POST" {
				uploadHandler.UploadPropertyImages(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		http.HandleFunc("/upload/service-request-image", func(w http.ResponseWriter, r *http.Request) {
			enableCORS(w, r)
			if r.Method == "OPTIONS" {
				return
			}
			if r.Method == "POST" {
				uploadHandler.UploadServiceRequestImage(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})

		http.HandleFunc("/files/", func(w http.ResponseWriter, r *http.Request) {
			enableCORS(w, r)
			if r.Method == "OPTIONS" {
				return
			}
			if r.Method == "DELETE" {
				uploadHandler.DeleteFile(w, r)
			} else {
				http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			}
		})
	}

	// Get port from environment variable or default to 8002
	port := os.Getenv("PORT")
	if port == "" {
		port = "8002"
	}

	fmt.Printf("Server starting on port %s...\n", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("❌ Could not start server: %s\n", err.Error())
	}
}
