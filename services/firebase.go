package services

import (
    "context"
    "cloud.google.com/go/firestore"
    "github.com/dzuura/neurodyx-be/config"
)

// GetFirestoreClient returns a Firestore client for the given context.
func GetFirestoreClient(ctx context.Context) (*firestore.Client, error) {
    return config.App.Firestore(ctx)
}