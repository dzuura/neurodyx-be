package config

import (
	"context"
	"log"
	"os"
	"sync"
	"time"
	"fmt"

	firebase "firebase.google.com/go"
	"cloud.google.com/go/firestore"
	"google.golang.org/api/option"
	"github.com/patrickmn/go-cache"
)

var (
	App *firebase.App
	FirestoreClient *firestore.Client
	once sync.Once
	JWTSecret []byte
	AssessmentQuestionCache *cache.Cache
	ScreeningQuestionCache *cache.Cache
	TherapyQuestionCache *cache.Cache
	cacheExpiration = 20 * time.Minute
)

// InitFirebase initializes Firebase and Firestore clients.
func InitFirebase() error {
	var initErr error
	once.Do(func() {
		secret := os.Getenv("JWT_SECRET")
		if secret == "" || len(secret) < 32 {
			initErr = logAndReturnError("JWT_SECRET is not set or too short, minimum 32 characters required")
			return
		}
		JWTSecret = []byte(secret)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		credentialsPath := os.Getenv("FIREBASE_CREDENTIALS_PATH")
		if credentialsPath == "" {
			initErr = logAndReturnError("FIREBASE_CREDENTIALS_PATH is not set")
			return
		}

		opt := option.WithCredentialsFile(credentialsPath)
		app, err := firebase.NewApp(ctx, nil, opt)
		if err != nil {
			initErr = logAndReturnError("Error initializing Firebase: %v", err)
			return
		}
		App = app

		FirestoreClient, err = app.Firestore(ctx)
		if err != nil {
			initErr = logAndReturnError("Error initializing Firestore client: %v", err)
			return
		}

        // Initialize caches with a default expiration time
		AssessmentQuestionCache = cache.New(cacheExpiration, 10*time.Minute)
		ScreeningQuestionCache = cache.New(cacheExpiration, 10*time.Minute)
		TherapyQuestionCache = cache.New(cacheExpiration, 10*time.Minute)
	})
	return initErr
}

// logAndReturnError logs an error message and returns it as an error
func logAndReturnError(msg string, args ...interface{}) error {
	log.Printf(msg, args...)
	return fmt.Errorf(msg, args...)
}

// StoreInCache stores data in the cache with a specified expiration
func StoreInCache(cache *cache.Cache, key, value interface{}) {
	cache.Set(fmt.Sprintf("%v", key), value, cacheExpiration)
    log.Printf("Stored in cache with key: %v, expires in: %v", key, cacheExpiration)
}

// LoadFromCache retrieves data from the cache if it exists
func LoadFromCache(cache *cache.Cache, key interface{}) (interface{}, bool) {
	val, found := cache.Get(fmt.Sprintf("%v", key))
    if found {
        log.Printf("Cache hit for key: %v", key)
    } else {
        log.Printf("Cache miss for key: %v", key)
    }
	return val, found
}