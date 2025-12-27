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
	}

	ctx := context.Background()
	creds := os.Getenv("FIREBASE_CREDENTIALS")
	var opt option.ClientOption

	// Check if creds is JSON content (starts with {) or a file path
	if len(creds) > 0 && creds[0] == '{' {
		fmt.Println("Loading Firebase credentials from JSON content")
		opt = option.WithCredentialsJSON([]byte(creds))
	} else {
		if creds == "" {
			fmt.Println("Warning: FIREBASE_CREDENTIALS is empty. Application Default Credentials will be used.")
		} else {
			fmt.Println("Loading Firebase credentials from file path")
		}
		opt = option.WithCredentialsFile(creds)
	}

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
		}
	}

	bucket := storageClient.Bucket(storageBucket)
	if _, err := bucket.Attrs(ctx); err != nil {
	}
	fmt.Println("Firebase initialized", client, authClient, storageClient)
}
