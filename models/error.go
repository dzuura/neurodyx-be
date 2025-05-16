package models

// ErrorResponse represents a standard error response structure.
type ErrorResponse struct {
    Error string `json:"error,omitempty"`
}