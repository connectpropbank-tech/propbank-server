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

// GetFirestoreClient returns the Firestore client
func GetFirestoreClient() *firestore.Client {
	return client
}

// GetAuthClient returns the Auth client
func GetAuthClient() *auth.Client {
	return authClient
}

// GetStorageClient returns the Storage client
func GetStorageClient() *storage.Client {
	return storageClient
}

// GetFirebaseApp returns the Firebase app instance
func GetFirebaseApp() *firebase.App {
	return firebaseApp
}

func InitFirebase() {
	// Load environment variables from .env file
	err := godotenv.Load("config/.env")
	if err != nil {
		log.Printf("Warning: Error loading .env file: %v", err)
	}

	ctx := context.Background()
	opt := option.WithCredentialsFile(os.Getenv("FIREBASE_CREDENTIALS"))

	// Initialize Firebase app
	firebaseApp, err = firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Fatalf("error initializing app: %v", err)
	}

	// Initialize Auth client
	authClient, err = firebaseApp.Auth(ctx)
	if err != nil {
		log.Fatalf("error getting Auth client: %v", err)
	}

	// Initialize Firestore client
	client, err = firebaseApp.Firestore(ctx)
	if err != nil {
		log.Fatalf("error getting Firestore client: %v", err)
	}

	// Initialize Storage client
	storageClient, err = storage.NewClient(ctx, opt)
	if err != nil {
		log.Fatalf("error getting Storage client: %v", err)
	}

	fmt.Println("🔥 Firebase initialized", client, authClient, storageClient)
}
