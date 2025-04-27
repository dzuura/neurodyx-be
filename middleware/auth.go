package middleware

import (
    "context"
    "net/http"
    "os"
    "strings"

    "github.com/dgrijalva/jwt-go"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

func AuthMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        authHeader := r.Header.Get("Authorization")
        if authHeader == "" {
            http.Error(w, "No token provided", http.StatusUnauthorized)
            return
        }

        tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
        token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
            return jwtSecret, nil
        })
        if err != nil || !token.Valid {
            http.Error(w, "Invalid token", http.StatusUnauthorized)
            return
        }

        claims, ok := token.Claims.(jwt.MapClaims)
        if !ok {
            http.Error(w, "Invalid token claims", http.StatusUnauthorized)
            return
        }

        userID := claims["uid"].(string)
        firebaseToken := claims["firebaseToken"].(string)
        ctx := context.WithValue(r.Context(), "userID", userID)
        ctx = context.WithValue(ctx, "firebaseToken", firebaseToken)
        next.ServeHTTP(w, r.WithContext(ctx))
    }
}