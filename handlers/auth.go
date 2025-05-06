package handlers

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"cloud.google.com/go/firestore"
	"github.com/dgrijalva/jwt-go"
	"github.com/dzuura/neurodyx-be/config"
	"github.com/dzuura/neurodyx-be/models"
	"golang.org/x/time/rate"
	"google.golang.org/api/idtoken"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

var (
    authLimiterStore = &LimiterStore{limiters: make(map[string]*rate.Limiter), mu: sync.Mutex{}}
    refreshLimiterStore = &LimiterStore{limiters: make(map[string]*rate.Limiter), mu: sync.Mutex{}}
    tokenCache = sync.Map{}
)

type LimiterStore struct {
    limiters map[string]*rate.Limiter
    mu sync.Mutex
}

// GetLimiter retrieves or creates a rate limiter for a given key
func (ls *LimiterStore) GetLimiter(key string) *rate.Limiter {
    ls.mu.Lock()
    defer ls.mu.Unlock()
    if _, exists := ls.limiters[key]; !exists {
        ls.limiters[key] = rate.NewLimiter(rate.Every(time.Minute), 10)
    }
    return ls.limiters[key]
}

// generateToken creates a JWT token with the specified user ID and expiry
func generateToken(uid string, expiry time.Duration) (string, error) {
    token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
        "uid": uid,
        "exp": time.Now().Add(expiry).Unix(),
    })
    return token.SignedString(config.JWTSecret)
}

// AuthHandler authenticates users and issues access and refresh tokens
func AuthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	clientIP := r.RemoteAddr
	if !authLimiterStore.GetLimiter(clientIP).Allow() {
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(models.AuthResponse{Error: "Rate limit exceeded"})
		return
	}

	var req models.AuthRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.AuthResponse{Error: "Invalid request body: " + err.Error()})
		return
	}

	if req.Token == "" || (req.AuthType != "firebase" && req.AuthType != "google") {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(models.AuthResponse{Error: "Missing or invalid auth type or token"})
		return
	}

	ctx, cancel := context.WithTimeout(r.Context(), 10*time.Second)
	defer cancel()

	var token *idtoken.Payload
	if cached, ok := tokenCache.Load(req.Token); ok {
		token = cached.(*idtoken.Payload)
	} else {
		if req.AuthType == "firebase" {
			client, err := config.App.Auth(ctx)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				json.NewEncoder(w).Encode(models.AuthResponse{Error: "Authentication service unavailable"})
				return
			}
			firebaseToken, err := client.VerifyIDToken(ctx, req.Token)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(models.AuthResponse{Error: "Invalid Firebase token"})
				return
			}
			token = &idtoken.Payload{
				Subject: firebaseToken.UID,
				Claims:  firebaseToken.Claims,
			}
		} else {
			googleClientID := os.Getenv("GOOGLE_CLIENT_ID")
			googleToken, err := idtoken.Validate(ctx, req.Token, googleClientID)
			if err != nil {
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(models.AuthResponse{Error: "Invalid Google ID token"})
				return
			}
			token = googleToken
		}
		tokenCache.Store(req.Token, token)
	}

	uid := token.Subject
	accessToken, err := generateToken(uid, time.Hour*24)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.AuthResponse{Error: "Error generating access token"})
		return
	}

	refreshToken, err := generateToken(uid, time.Hour*24*30)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.AuthResponse{Error: "Error generating refresh token"})
		return
	}

	email := ""
	username := ""
	claims := token.Claims
	if e, ok := claims["email"].(string); ok {
		email = e
	}
	if name, ok := claims["name"].(string); ok {
		username = name
	}

	_, err = config.FirestoreClient.Collection("users").Doc(uid).Get(ctx)
	isNewUser := status.Code(err) == codes.NotFound
	if err != nil && !isNewUser {
		log.Printf("Error checking user existence: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.AuthResponse{Error: "Error checking user existence"})
		return
	}

	userData := map[string]interface{}{
		"refreshToken":         refreshToken,
		"refreshTokenCreatedAt": time.Now(),
		"refreshTokenExpiresAt": time.Now().Add(time.Hour * 24 * 30),
		"email":                email,
		"username":             username,
	}

	if isNewUser {
		userData["createdAt"] = time.Now()
	}

	_, err = config.FirestoreClient.Collection("users").Doc(uid).Set(ctx, userData, firestore.MergeAll)
	if err != nil {
		log.Printf("Error saving user data to Firestore: %v", err)
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(models.AuthResponse{Error: "Error saving user data"})
		return
	}

	json.NewEncoder(w).Encode(models.AuthResponse{
		Token:        accessToken,
		RefreshToken: refreshToken,
	})
}

// RefreshHandler refreshes an access token using a refresh token
func RefreshHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    clientIP := r.RemoteAddr
    if !refreshLimiterStore.GetLimiter(clientIP).Allow() {
        w.WriteHeader(http.StatusTooManyRequests)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Rate limit exceeded"})
        return
    }

    var req models.RefreshRequest
    if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Invalid request body: " + err.Error()})
        return
    }

    if req.RefreshToken == "" {
        w.WriteHeader(http.StatusBadRequest)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Missing refresh token"})
        return
    }

    ctx, cancel := context.WithTimeout(r.Context(), 5*time.Second)
    defer cancel()

    tokenChan := make(chan *jwt.Token, 1)

    go func() {
        token, _ := jwt.Parse(req.RefreshToken, func(token *jwt.Token) (interface{}, error) {
            if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
                return nil, fmt.Errorf("unexpected signing method")
            }
            return config.JWTSecret, nil
        })
        tokenChan <- token
    }()

    var token *jwt.Token
    select {
    case token = <-tokenChan:
        if !token.Valid {
            w.WriteHeader(http.StatusUnauthorized)
            json.NewEncoder(w).Encode(models.AuthResponse{Error: "Invalid or expired refresh token"})
            return
        }
    case <-ctx.Done():
        w.WriteHeader(http.StatusRequestTimeout)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Request timeout"})
        return
    }

    claims, ok := token.Claims.(jwt.MapClaims)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Invalid token claims"})
        return
    }

    uid, ok := claims["uid"].(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Invalid UID in token"})
        return
    }

    doc, err := config.FirestoreClient.Collection("users").Doc(uid).Get(ctx)
    if err != nil {
        log.Printf("Error retrieving user data from Firestore: %v", err)
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Invalid refresh token"})
        return
    }

    var userData map[string]interface{}
    if err := doc.DataTo(&userData); err != nil {
        log.Printf("Error parsing user data: %v", err)
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Invalid refresh token"})
        return
    }

    storedRefreshToken, ok := userData["refreshToken"].(string)
    if !ok || storedRefreshToken != req.RefreshToken {
        log.Printf("Refresh token mismatch for userID %s", uid)
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Invalid refresh token"})
        return
    }

    refreshTokenExpiresAt, ok := userData["refreshTokenExpiresAt"].(time.Time)
    if !ok || refreshTokenExpiresAt.Before(time.Now()) {
        log.Printf("Refresh token expired for userID %s", uid)
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Expired refresh token"})
        return
    }

    newAccessToken, err := generateToken(uid, time.Hour*24)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Error generating new access token"})
        return
    }

    newRefreshToken, err := generateToken(uid, time.Hour*24*30)
    if err != nil {
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Error generating new refresh token"})
        return
    }

    updatedUserData := map[string]interface{}{
        "refreshToken": newRefreshToken,
        "refreshTokenCreatedAt": time.Now(),
        "refreshTokenExpiresAt": time.Now().Add(time.Hour * 24 * 30),
    }
    _, err = config.FirestoreClient.Collection("users").Doc(uid).Set(ctx, updatedUserData, firestore.MergeAll)
    if err != nil {
        log.Printf("Error updating user data in Firestore: %v", err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.AuthResponse{Error: "Error updating user data"})
        return
    }

    json.NewEncoder(w).Encode(models.AuthResponse{
        Token:        newAccessToken,
        RefreshToken: newRefreshToken,
    })
}