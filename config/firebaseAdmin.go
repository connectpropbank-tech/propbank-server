package config

import (
	"context"
	"fmt"
	"log"
	"os"
	"cloud.google.com/go/firestore"
	"cloud.google.com/go/storage"
	firebase "firebase.google.com/go/v4"
	"firebase.google.com/go/v4/auth"
	"github.com/joho/godotenv"
	"google.golang.org/api/option"
)

var client *firestore.Client
var authClient *auth.Client
var storageClient *storage.Client
var firebaseApp *firebase.App
var storageBucket string

// GetFirestoreClient returns the Firestore client
func GetFirestoreClient() *firestore.Client {
	return client
}


func GetAuthClient() *auth.Client {
	return authClient
}

// returns the Storage client
func GetStorageClient() *storage.Client {
	return storageClient
}

// returns the Firebase app instance
func GetFirebaseApp() *firebase.App {
	return firebaseApp
}

// returns the Storage bucket name
func GetStorageBucket() string {
	return storageBucket
}

func InitFirebase() {
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	ctx := context.Background()
	opt := option.WithCredentialsFile(os.Getenv("FIREBASE_CREDENTIALS"))

	firebaseApp, err = firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	authClient, err = firebaseApp.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	client, err = firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("error getting Firestore client: %v", err)
	}

	storageClient, err = storage.NewClient(ctx, opt)
	if err != nil {
		log.Fatalf("error getting Storage client: %v", err)
	}

	// Get Storage bucket from environment variable or construct from project ID
	storageBucket = os.Getenv("STORAGE_BUCKET")
	if storageBucket == "" {
		// Try to get project ID from Firebase app
		// For Firebase projects, default bucket is usually {project-id}.appspot.com
		// We can also get it from the credentials file
		storageBucket = os.Getenv("FIREBASE_STORAGE_BUCKET")
		if storageBucket == "" {
			// Try to get project ID from credentials file path or use default
			// The bucket name format is typically: {project-id}.appspot.com
			storageBucket = "propbank-a98ed.appspot.com" // Default fallback
			log.Printf(" Using default bucket name. Set STORAGE_BUCKET env var to override.")
		}
	}

	log.Printf("📦 Using Storage bucket: %s", storageBucket)
	
	bucket := storageClient.Bucket(storageBucket)
	if _, err := bucket.Attrs(ctx); err != nil {
		log.Printf(" Warning: Could not verify bucket '%s' exists: %v", storageBucket, err)
		log.Printf("You may need to create the bucket in Firebase Console or update the bucket name")
	}
	fmt.Println("Firebase initialized", client, authClient, storageClient)
}
