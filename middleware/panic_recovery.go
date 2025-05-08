package middleware

import (
    "log"
    "net/http"
	"encoding/json"

    "github.com/dzuura/neurodyx-be/models"
)

// PanicRecoveryMiddleware catches panics and returns a 500 error response.
func PanicRecoveryMiddleware(next http.HandlerFunc) http.HandlerFunc {
    return func(w http.ResponseWriter, r *http.Request) {
        defer func() {
            if r := recover(); r != nil {
                log.Printf("Recovered from panic: %v", r)
                w.Header().Set("Content-Type", "application/json")
                w.WriteHeader(http.StatusInternalServerError)
                json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Internal server error"})
            }
        }()
        next.ServeHTTP(w, r)
    }
}