package middleware

import (
    "context"
    "log"
    "net/http"
    "strings"
    "time"
    "fmt"

    "github.com/golang-jwt/jwt/v5"
    "github.com/dzuura/neurodyx-be/config"
)

// contextKey is a custom type for context keys to avoid collisions
type contextKey string

// UserIDKey is the key used to store user ID in the context
const UserIDKey contextKey = "userID"

// AuthMiddleware verifies JWT tokens and adds userID to the request context.
func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
            return
        }

        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
        if tokenStr == authHeader {
            http.Error(w, "Invalid Authorization header format", http.StatusUnauthorized)
            return
        }

        log.Printf("Received token for verification: %s", tokenStr)

        token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method")
            }
            return config.JWTSecret, nil
        })
        if err != nil {
            log.Printf("Token parsing error for %s %s: %v", r.Method, r.URL.Path, err)
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            log.Printf("Token claims invalid for %s %s", r.Method, r.URL.Path)
            http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            return
        }

        exp, ok := claims["exp"].(float64)
        if !ok || time.Now().Unix() > int64(exp) {
            log.Printf("Token expired for %s %s", r.Method, r.URL.Path)
            http.Error(w, "Token expired", http.StatusUnauthorized)
            return
        }

        uid, ok := claims["uid"].(string)
        if !ok {
            log.Printf("No UID in token for %s %s", r.Method, r.URL.Path)
            http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            return
        }

        ctx := context.WithValue(r.Context(), UserIDKey, uid)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}