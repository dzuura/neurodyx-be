package main

import (
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/joho/godotenv"
)

func main() {
	// Load .env file
	if err := godotenv.Load(); err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	// Setup router
	r := mux.NewRouter()

	// Placeholder route
	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Neurodyx Backend is running"))
	}).Methods("GET")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Server starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}