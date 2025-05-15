package handlers

import (
    "encoding/json"
    "log"
    "net/http"
    "strconv"
	"time"

    "github.com/dzuura/neurodyx-be/middleware"
    "github.com/dzuura/neurodyx-be/models"
    "github.com/dzuura/neurodyx-be/services"
)

// GetWeeklyProgressHandler retrieves the user's therapy progress for the last 7 days.
func GetWeeklyProgressHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    progress, err := services.GetWeeklyProgress(r.Context(), userID)
    if err != nil {
        log.Printf("Error retrieving weekly progress for userID %s: %v", userID, err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve weekly progress: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(progress)
}

// GetMonthlyProgressHandler retrieves the user's therapy progress for a specific month.
func GetMonthlyProgressHandler(w http.ResponseWriter, r *http.Request) {
    w.Header().Set("Content-Type", "application/json")

    userID, ok := r.Context().Value(middleware.UserIDKey).(string)
    if !ok {
        w.WriteHeader(http.StatusUnauthorized)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "User ID missing"})
        return
    }

    now := time.Now().UTC()
    yearStr := r.URL.Query().Get("year")
    monthStr := r.URL.Query().Get("month")

    year, err := strconv.Atoi(yearStr)
    if err != nil || year < 2000 || year > now.Year() {
        year = now.Year()
    }

    month, err := strconv.Atoi(monthStr)
    if err != nil || month < 1 || month > 12 {
        month = int(now.Month())
    }

    progress, err := services.GetMonthlyProgress(r.Context(), userID, year, month)
    if err != nil {
        log.Printf("Error retrieving monthly progress for userID %s, year %d, month %d: %v", userID, year, month, err)
        w.WriteHeader(http.StatusInternalServerError)
        json.NewEncoder(w).Encode(models.ErrorResponse{Error: "Failed to retrieve monthly progress: " + err.Error()})
        return
    }

    w.WriteHeader(http.StatusOK)
    json.NewEncoder(w).Encode(progress)
}