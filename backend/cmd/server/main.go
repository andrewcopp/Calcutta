package main

import (
	"encoding/json"
	"log"
	"net/http"
)

type School struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

var schools = []School{
	{"550e8400-e29b-41d4-a716-446655440000", "Duke"},
	{"550e8400-e29b-41d4-a716-446655440001", "North Carolina"},
	{"550e8400-e29b-41d4-a716-446655440002", "Kansas"},
	{"550e8400-e29b-41d4-a716-446655440003", "Kentucky"},
	{"550e8400-e29b-41d4-a716-446655440004", "Gonzaga"},
	{"550e8400-e29b-41d4-a716-446655440005", "Villanova"},
	{"550e8400-e29b-41d4-a716-446655440006", "Michigan"},
	{"550e8400-e29b-41d4-a716-446655440007", "UCLA"},
	{"550e8400-e29b-41d4-a716-446655440008", "Arizona"},
	{"550e8400-e29b-41d4-a716-446655440009", "Baylor"},
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
