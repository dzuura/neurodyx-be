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
)

var (
    App *firebase.App
    FirestoreClient *firestore.Client
    once sync.Once
    JWTSecret []byte
    AssessmentQuestionCache sync.Map
    ScreeningQuestionCache sync.Map
    TherapyQuestionCache sync.Map
    cacheCleanupInterval = 5 * time.Minute
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

        go startCacheCleanup()
    })
    return initErr
}

// logAndReturnError logs an error message and returns it as an error
func logAndReturnError(msg string, args ...interface{}) error {
    log.Printf(msg, args...)
    return fmt.Errorf(msg, args...)
}

// startCacheCleanup periodically cleans up expired cache entries
func startCacheCleanup() {
    ticker := time.NewTicker(cacheCleanupInterval)
    defer ticker.Stop()

    for range ticker.C {
        cleanupCache(&AssessmentQuestionCache)
        cleanupCache(&ScreeningQuestionCache)
        cleanupCache(&TherapyQuestionCache)
    }
}

// cleanupCache removes expired entries from the cache
func cleanupCache(cache *sync.Map) {
    now := time.Now()
    cache.Range(func(key, value interface{}) bool {
        if entry, ok := value.(cacheEntry); ok {
            if now.Sub(entry.timestamp) > cacheCleanupInterval {
                cache.Delete(key)
                log.Printf("Cleaned up cached entry for key: %v", key)
            }
        }
        return true
    })
}

type cacheEntry struct {
    data     interface{}
    timestamp time.Time
}

// StoreInCache stores data in the cache with a timestamp
func StoreInCache(cache *sync.Map, key, value interface{}) {
    cache.Store(key, cacheEntry{data: value, timestamp: time.Now()})
}

// LoadFromCache retrieves data from the cache if it exists
func LoadFromCache(cache *sync.Map, key interface{}) (interface{}, bool) {
    if val, ok := cache.Load(key); ok {
        if entry, ok := val.(cacheEntry); ok {
            return entry.data, true
        }
    }
    return nil, false
}