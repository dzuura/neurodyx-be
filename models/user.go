package models

import "time"

// Credentials represents the credentials for user authentication.
type Credentials struct {
    Email    string `json:"email"`
    Password string `json:"password,omitempty"`
}

// User represents a user in the system with unique fields.
type User struct {
    ID                 string       `json:"id,omitempty"`
    Username           string       `json:"username,omitempty"`
    Email              string       `json:"email,omitempty"`
    CreatedAt          time.Time    `json:"createdAt,omitempty"`
    RefreshTokenCreatedAt time.Time `json:"refreshTokenCreatedAt,omitempty"`
    RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt,omitempty"`
}

// AuthRequest represents a request for authentication with unique fields.
type AuthRequest struct {
    Token      string `json:"token"`
    AuthType   string `json:"authType"`
}

// RegisterRequest represents a request for user registration with unique fields.
type RefreshRequest struct {
    RefreshToken string `json:"refreshToken"`
}

// AuthResponse represents a response for authentication with unique fields.
type AuthResponse struct {
    Token        string `json:"token,omitempty"`
    RefreshToken string `json:"refreshToken,omitempty"`
    Error        string `json:"error,omitempty"`
}