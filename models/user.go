package models

import "time"

type Credentials struct {
    Email    string `json:"email"`
    Password string `json:"password,omitempty"`
}

type User struct {
    ID                 string       `json:"id,omitempty"`
    Username           string       `json:"username,omitempty"`
    Email              string       `json:"email,omitempty"`
    CreatedAt          time.Time    `json:"createdAt,omitempty"`
    RefreshTokenCreatedAt time.Time `json:"refreshTokenCreatedAt,omitempty"`
    RefreshTokenExpiresAt time.Time `json:"refreshTokenExpiresAt,omitempty"`
}

type AuthRequest struct {
    Token      string `json:"token"`
    AuthType   string `json:"authType"`
}

type RefreshRequest struct {
    RefreshToken string `json:"refreshToken"`
}

type AuthResponse struct {
    Token        string `json:"token,omitempty"`
    RefreshToken string `json:"refreshToken,omitempty"`
    Error        string `json:"error,omitempty"`
}