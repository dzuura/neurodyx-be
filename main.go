package main

import (
    "log"
    "net/http"
    "os"

    "github.com/gorilla/mux"
    "github.com/joho/godotenv"
    "github.com/dzuura/neurodyx-be/config"
    "github.com/dzuura/neurodyx-be/handlers"
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
    r.HandleFunc("/api/auth/login", handlers.LoginHandler).Methods("POST")
    r.HandleFunc("/api/auth/register", handlers.RegisterHandler).Methods("POST")
    r.HandleFunc("/api/auth/google", handlers.GoogleLoginHandler).Methods("POST")

    // Start server
    port := os.Getenv("PORT")
    if port == "" {
        port = "8080"
    }
    log.Printf("Server starting on :%s", port)
    log.Fatal(http.ListenAndServe(":"+port, r))
}