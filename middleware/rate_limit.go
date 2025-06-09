package middleware

import (
    "net/http"
    "sync"
    "encoding/json"

    "github.com/dzuura/neurodyx-be/models"
    "golang.org/x/time/rate"
)

// LimiterStore manages rate limiters for different keys.
type LimiterStore struct {
    limiters map[string]*rate.Limiter
    mu       sync.Mutex
}

// GetLimiter retrieves or creates a rate limiter for a given key.
func (ls *LimiterStore) GetLimiter(key string) *rate.Limiter {
    ls.mu.Lock()
    defer ls.mu.Unlock()
    if _, exists := ls.limiters[key]; !exists {
        ls.limiters[key] = rate.NewLimiter(rate.Limit(25)/60, 25)
    }
    return ls.limiters[key]
}

// NewLimiterStore creates a new LimiterStore.
func NewLimiterStore() *LimiterStore {
    return &LimiterStore{limiters: make(map[string]*rate.Limiter), mu: sync.Mutex{}}
}

// RateLimitMiddleware enforces rate limiting based on userID if available, otherwise falls back to client IP.
func RateLimitMiddleware(ls *LimiterStore, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        key := r.RemoteAddr

        if userID, ok := r.Context().Value(UserIDKey).(string); ok {
            key = userID
        }

        if !ls.GetLimiter(key).Allow() {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusTooManyRequests)
            json.NewEncoder(w).Encode(models.AuthResponse{Error: "Rate limit exceeded"})
            return
        }
        next.ServeHTTP(w, r)
    }
}