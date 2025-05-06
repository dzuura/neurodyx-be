package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "github.com/dzuura/neurodyx-be/config"
    "github.com/dzuura/neurodyx-be/handlers"
    "github.com/dzuura/neurodyx-be/middleware"
)

func main() {
    // Load .env file
    if err := godotenv.Load(); err != nil {
        log.Fatalf("Error loading .env file: %v", err)
    }

    // Initialize Firebase
    if err := config.InitFirebase(); err != nil {
        log.Fatalf("Failed to initialize Firebase: %v", err)
    }

    // Setup router
    r := mux.NewRouter()

    // Public routes
    r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Neurodyx Backend is running"))
    }).Methods("GET")
    r.HandleFunc("/api/auth", handlers.AuthHandler).Methods("POST")
    r.HandleFunc("/api/refresh", handlers.RefreshHandler).Methods("POST")

    // Protected routes for screening
    screeningRouter := r.PathPrefix("/api/screening").Subrouter()
    screeningRouter.HandleFunc("/questions", middleware.AuthMiddleware(handlers.AddScreeningQuestionHandler)).Methods("POST")
    screeningRouter.HandleFunc("/questions", middleware.AuthMiddleware(handlers.GetScreeningQuestionsHandler)).Methods("GET")
    screeningRouter.HandleFunc("/submit", middleware.AuthMiddleware(handlers.SubmitScreeningHandler)).Methods("POST")

    // Protected routes for assessment
    assessmentRouter := r.PathPrefix("/api/assessment").Subrouter()
    assessmentRouter.HandleFunc("/questions", middleware.AuthMiddleware(handlers.AddAssessmentQuestionHandler)).Methods("POST")
    assessmentRouter.HandleFunc("/questions", middleware.AuthMiddleware(handlers.GetAssessmentQuestionsHandler)).Methods("GET")
    assessmentRouter.HandleFunc("/submit", middleware.AuthMiddleware(handlers.SubmitAnswerHandler)).Methods("POST")
    assessmentRouter.HandleFunc("/results", middleware.AuthMiddleware(handlers.GetAssessmentResultsHandler)).Methods("GET")

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Server starting on :%s", port)
    if err := http.ListenAndServe(":"+port, r); err != nil {
        log.Fatalf("Server failed to start: %v", err)
    }
}