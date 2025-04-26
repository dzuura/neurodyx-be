package config

import (
	"context"
	"log"
	"os"

	firebase "firebase.google.com/go"
	"google.golang.org/api/option"
)

var App *firebase.App

func InitFirebase() error {
	ctx := context.Background()
	credentialsPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
	opt := option.WithCredentialsFile(credentialsPath)
	app, err := firebase.NewApp(ctx, nil, opt)
	if err != nil {
		log.Printf("Error initializing Firebase: %v", err)
		return err
	}
	App = app
	return nil
}