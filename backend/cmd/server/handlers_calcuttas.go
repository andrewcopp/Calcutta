package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/gorilla/mux"
)

// Calcutta API handlers
func calcuttasHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	if r.Method == "POST" {
		log.Printf("Handling POST request to /api/calcuttas")

		// Parse the request body
		var calcutta models.Calcutta
		if err := json.NewDecoder(r.Body).Decode(&calcutta); err != nil {
			log.Printf("Error decoding request body: %v", err)
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		log.Printf("Received request to create calcutta: %+v", calcutta)

		// Create the calcutta
		if err := calcuttaRepo.Create(r.Context(), &calcutta); err != nil {
			log.Printf("Error creating calcutta: %v", err)
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		log.Printf("Successfully created calcutta with ID: %s", calcutta.ID)

		// Create the rounds for the Calcutta
		rounds := []struct {
			round  int
			points int
		}{
			{1, 50},  // Round of 64
			{2, 100}, // Round of 32
			{3, 150}, // Sweet 16
			{4, 200}, // Elite 8
			{5, 250}, // Final 4
			{6, 300}, // Championship
		}

		log.Printf("Creating %d rounds for calcutta %s", len(rounds), calcutta.ID)
		for _, round := range rounds {
			calcuttaRound := &models.CalcuttaRound{
				CalcuttaID: calcutta.ID,
				Round:      round.round,
				Points:     round.points,
			}
			if err := calcuttaRepo.CreateRound(r.Context(), calcuttaRound); err != nil {
				log.Printf("Error creating round %d for calcutta %s: %v", round.round, calcutta.ID, err)
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}
			log.Printf("Successfully created round %d for calcutta %s", round.round, calcutta.ID)
		}

		// Return the created calcutta
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(calcutta)
		log.Printf("Successfully completed POST request for calcutta %s", calcutta.ID)
		return
	}

	// Handle GET request
	log.Printf("Handling GET request to /api/calcuttas")
	calcuttas, err := calcuttaRepo.GetAll(r.Context())
	if err != nil {
		log.Printf("Error getting all calcuttas: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	log.Printf("Successfully retrieved %d calcuttas", len(calcuttas))

	// Ensure we return an empty array [] instead of null when there are no calcuttas
	if calcuttas == nil {
		calcuttas = []*models.Calcutta{}
	}

	json.NewEncoder(w).Encode(calcuttas)
}

func calcuttaEntriesHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract calcutta ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	calcuttaID := pathParts[3]

	entries, err := calcuttaRepo.GetEntries(r.Context(), calcuttaID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if entries == nil {
		entries = []*models.CalcuttaEntry{}
	}
	json.NewEncoder(w).Encode(entries)
}

func entryTeamsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract entry ID from URL path
	entryID := r.URL.Path[len("/api/entries/"):]
	if entryID == "" {
		http.Error(w, "Entry ID is required", http.StatusBadRequest)
		return
	}

	teams, err := calcuttaRepo.GetEntryTeams(r.Context(), entryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(teams)
}

func calcuttaEntryTeamHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract calcutta ID and entry ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 6 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	entryID := pathParts[5]

	teams, err := calcuttaRepo.GetEntryTeams(r.Context(), entryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(teams)
}

func portfoliosHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract entry ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	entryID := pathParts[3]

	portfolios, err := calcuttaRepo.GetPortfoliosByEntry(r.Context(), entryID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(portfolios)
}

func portfolioTeamsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	// Extract portfolio ID from URL path
	pathParts := strings.Split(r.URL.Path, "/")
	if len(pathParts) < 4 {
		http.Error(w, "Invalid URL path", http.StatusBadRequest)
		return
	}
	portfolioID := pathParts[3]

	teams, err := calcuttaRepo.GetPortfolioTeams(r.Context(), portfolioID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(teams)
}

func calcuttaHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	log.Printf("Handling GET request to /api/calcuttas/{id}")
	log.Printf("Request URL: %s", r.URL.String())
	log.Printf("Request Method: %s", r.Method)

	// Extract calcutta ID from URL path
	vars := mux.Vars(r)
	calcuttaID := vars["id"]
	log.Printf("Extracted calcutta ID from path: %s", calcuttaID)

	if calcuttaID == "" {
		log.Printf("Error: Calcutta ID is empty")
		http.Error(w, "Calcutta ID is required", http.StatusBadRequest)
		return
	}

	log.Printf("Attempting to fetch calcutta with ID: %s", calcuttaID)
	calcutta, err := calcuttaRepo.GetByID(r.Context(), calcuttaID)
	if err != nil {
		if err.Error() == "calcutta not found" {
			log.Printf("Calcutta not found with ID: %s", calcuttaID)
			http.Error(w, err.Error(), http.StatusNotFound)
			return
		}
		log.Printf("Error fetching calcutta: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Printf("Successfully retrieved calcutta: %+v", calcutta)
	json.NewEncoder(w).Encode(calcutta)
}
