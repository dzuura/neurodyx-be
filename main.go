package main

import (
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

func main() {
	// Setup router
	r := mux.NewRouter()

	// Placeholder route
	r.HandleFunc("/api/health", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Neurodyx Backend is running"))
	}).Methods("GET")

	// Start server
	log.Println("Server starting on :8080")
	log.Fatal(http.ListenAndServe(":8080", r))
}