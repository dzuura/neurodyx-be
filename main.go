package main

import (
    "log"
    "net/http"
    "os"

    "github.com/joho/godotenv"
    "github.com/gorilla/mux"
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

    // Setup rate limiters
    authLimiterStore := middleware.NewLimiterStore()
    refreshLimiterStore := middleware.NewLimiterStore()

    // Setup router
    r := mux.NewRouter()

    // Public routes
    r.HandleFunc("/api/health", middleware.PanicRecoveryMiddleware(func(w http.ResponseWriter, r *http.Request) {
        w.Write([]byte("Neurodyx Backend is running"))
    })).Methods("GET")
    r.HandleFunc("/api/auth", middleware.PanicRecoveryMiddleware(middleware.RateLimitMiddleware(authLimiterStore, handlers.AuthHandler))).Methods("POST")
    r.HandleFunc("/api/refresh", middleware.PanicRecoveryMiddleware(middleware.RateLimitMiddleware(refreshLimiterStore, handlers.RefreshHandler))).Methods("POST")

    // Protected routes for screening
    screeningRouter := r.PathPrefix("/api/screening").Subrouter()
    screeningRouter.HandleFunc("/questions", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.AddScreeningQuestionHandler))).Methods("POST")
    screeningRouter.HandleFunc("/questions", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.GetScreeningQuestionsHandler))).Methods("GET")
    screeningRouter.HandleFunc("/submit", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.SubmitScreeningHandler))).Methods("POST")

    // Protected routes for assessment
    assessmentRouter := r.PathPrefix("/api/assessment").Subrouter()
    assessmentRouter.HandleFunc("/questions", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.AddAssessmentQuestionHandler))).Methods("POST")
    assessmentRouter.HandleFunc("/questions", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.GetAssessmentQuestionsHandler))).Methods("GET")
    assessmentRouter.HandleFunc("/submit", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.SubmitAnswerHandler))).Methods("POST")
    assessmentRouter.HandleFunc("/results", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.GetAssessmentResultsHandler))).Methods("GET")

    // Protected routes for therapy
    therapyRouter := r.PathPrefix("/api/therapy").Subrouter()
    therapyRouter.HandleFunc("/questions", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.AddTherapyQuestionHandler))).Methods("POST")
    therapyRouter.HandleFunc("/categories", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.GetTherapyCategoriesHandler))).Methods("GET")
    therapyRouter.HandleFunc("/questions", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.GetTherapyQuestionsHandler))).Methods("GET")
    therapyRouter.HandleFunc("/submit", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.SubmitTherapyAnswerHandler))).Methods("POST")
    therapyRouter.HandleFunc("/results", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.GetTherapyResultsHandler))).Methods("GET")

    // Protected routes for progress tracking
    progressRouter := r.PathPrefix("/api/progress").Subrouter()
    progressRouter.HandleFunc("/weekly", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.GetWeeklyProgressHandler))).Methods("GET")
    progressRouter.HandleFunc("/monthly", middleware.PanicRecoveryMiddleware(middleware.AuthMiddleware(handlers.GetMonthlyProgressHandler))).Methods("GET")
    
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