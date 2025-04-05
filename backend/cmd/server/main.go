package main

import (
	"encoding/json"
	"log"
	"net/http"

	"calcutta/internal/models"
)

var schools = []models.School{
	{ID: "550e8400-e29b-41d4-a716-446655440000", Name: "Duke"},
	{ID: "550e8400-e29b-41d4-a716-446655440001", Name: "North Carolina"},
	{ID: "550e8400-e29b-41d4-a716-446655440002", Name: "Kansas"},
	{ID: "550e8400-e29b-41d4-a716-446655440003", Name: "Kentucky"},
	{ID: "550e8400-e29b-41d4-a716-446655440004", Name: "Gonzaga"},
	{ID: "550e8400-e29b-41d4-a716-446655440005", Name: "Villanova"},
	{ID: "550e8400-e29b-41d4-a716-446655440006", Name: "Michigan"},
	{ID: "550e8400-e29b-41d4-a716-446655440007", Name: "UCLA"},
	{ID: "550e8400-e29b-41d4-a716-446655440008", Name: "Arizona"},
	{ID: "550e8400-e29b-41d4-a716-446655440009", Name: "Baylor"},
}

func enableCORS(next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next(w, r)
	}
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "healthy",
		"message": "API is running",
	})
}

func schoolsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(schools)
}

func main() {
	http.HandleFunc("/api/health", enableCORS(healthHandler))
	http.HandleFunc("/api/schools", enableCORS(schoolsHandler))

	log.Printf("Starting server on :8080")
	if err := http.ListenAndServe(":8080", nil); err != nil {
		log.Fatal(err)
	}
}
