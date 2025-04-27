package handlers

import (
    "context"
    "encoding/json"
    "net/http"
    "os"
    "time"

    "github.com/dgrijalva/jwt-go"
    "github.com/dzuura/neurodyx-be/config"
    "google.golang.org/api/idtoken"
)

var jwtSecret = []byte(os.Getenv("JWT_SECRET"))

// Handles user login by verifying the Firebase token and issuing an internal JWT
func LoginHandler(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        FirebaseToken string `json:"firebaseToken"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Verify Firebase Token
    ctx := r.Context()
    client, err := config.App.Auth(ctx)
    if err != nil {
        http.Error(w, "Authentication service unavailable", http.StatusInternalServerError)
        return
    }

    token, err := client.VerifyIDToken(ctx, payload.FirebaseToken)
    if err != nil {
        http.Error(w, "Invalid Firebase token", http.StatusUnauthorized)
        return
    }

    // Generate internal JWT
    jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "uid":          token.UID,
        "firebaseToken": payload.FirebaseToken,
        "exp":          time.Now().Add(time.Hour * 24).Unix(),
    })
    tokenString, err := jwtToken.SignedString(jwtSecret)
    if err != nil {
        http.Error(w, "Error generating internal token", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// Handles user registration by verifying the Firebase token and issuing an internal JWT
func RegisterHandler(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        FirebaseToken string `json:"firebaseToken"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Verify Firebase Token
    ctx := r.Context()
    client, err := config.App.Auth(ctx)
    if err != nil {
        http.Error(w, "Authentication service unavailable", http.StatusInternalServerError)
        return
    }

    token, err := client.VerifyIDToken(ctx, payload.FirebaseToken)
    if err != nil {
        http.Error(w, "Invalid Firebase token", http.StatusUnauthorized)
        return
    }

    // Generate internal JWT
    jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "uid":          token.UID,
        "firebaseToken": payload.FirebaseToken,
        "exp":          time.Now().Add(time.Hour * 24).Unix(),
    })
    tokenString, err := jwtToken.SignedString(jwtSecret)
    if err != nil {
        http.Error(w, "Error generating internal token", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}

// Handles Google Sign-In by verifying the Google ID token and issuing an internal JWT
func GoogleLoginHandler(w http.ResponseWriter, r *http.Request) {
    var payload struct {
        IDToken string `json:"idToken"`
    }
    if err := json.NewDecoder(r.Body).Decode(&payload); err != nil {
        http.Error(w, "Invalid request body", http.StatusBadRequest)
        return
    }

    // Validate Google ID Token
    googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
    token, err := idtoken.Validate(context.Background(), payload.IDToken, googleClientID)
    if err != nil {
        http.Error(w, "Invalid Google ID token", http.StatusUnauthorized)
        return
    }

    // Generate internal JWT
    jwtToken := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "uid":          token.Subject,
        "firebaseToken": payload.IDToken, // Store Google ID token as firebaseToken
        "exp":          time.Now().Add(time.Hour * 24).Unix(),
    })
    tokenString, err := jwtToken.SignedString(jwtSecret)
    if err != nil {
        http.Error(w, "Error generating internal token", http.StatusInternalServerError)
        return
    }

    json.NewEncoder(w).Encode(map[string]string{"token": tokenString})
}