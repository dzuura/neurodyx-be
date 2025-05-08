package middleware

import (
    "net/http"
    "sync"
    "time"
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
        ls.limiters[key] = rate.NewLimiter(rate.Every(time.Minute), 10)
    }
    return ls.limiters[key]
}

// NewLimiterStore creates a new LimiterStore.
func NewLimiterStore() *LimiterStore {
    return &LimiterStore{limiters: make(map[string]*rate.Limiter), mu: sync.Mutex{}}
}

// RateLimitMiddleware enforces rate limiting based on client IP.
func RateLimitMiddleware(ls *LimiterStore, next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        clientIP := r.RemoteAddr
        if !ls.GetLimiter(clientIP).Allow() {
            w.Header().Set("Content-Type", "application/json")
            w.WriteHeader(http.StatusTooManyRequests)
            json.NewEncoder(w).Encode(models.AuthResponse{Error: "Rate limit exceeded"})
            return
        }
        next.ServeHTTP(w, r)
    }
}