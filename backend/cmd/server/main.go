package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/andrewcopp/Calcutta/backend/pkg/services"
	_ "github.com/lib/pq"
)

var schoolRepo *services.SchoolRepository
var schoolService *services.SchoolService
var tournamentRepo *services.TournamentRepository
var tournamentService *services.TournamentService
var calcuttaRepo *services.CalcuttaRepository
var calcuttaService *services.CalcuttaService

func init() {
	// Get database connection string from environment
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}

	// Connect to database
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Initialize repositories
	schoolRepo = services.NewSchoolRepository(db)
	tournamentRepo = services.NewTournamentRepository(db)
	calcuttaRepo = services.NewCalcuttaRepository(db)

	// Initialize services
	schoolService = services.NewSchoolService(schoolRepo)
	tournamentService = services.NewTournamentService(tournamentRepo)
	calcuttaService = services.NewCalcuttaService()
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

	schools, err := schoolService.GetAllSchools(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	json.NewEncoder(w).Encode(schools)
}

func tournamentsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	tournaments, err := tournamentService.GetAllTournaments(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create a response that includes tournament winners
	type TournamentResponse struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Rounds  int    `json:"rounds"`
		Winner  string `json:"winner,omitempty"`
		Created string `json:"created"`
	}

	var response []TournamentResponse
	for _, tournament := range tournaments {
		// Get the winning team for this tournament
		team, err := tournamentService.GetWinningTeam(r.Context(), tournament.ID)
		if err != nil {
			log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
			continue
		}

		winnerName := ""
		if team != nil {
			// Get the school name
			school, err := schoolService.GetSchoolByID(r.Context(), team.SchoolID)
			if err != nil {
				log.Printf("Error getting school for team %s: %v", team.ID, err)
				continue
			}
			winnerName = school.Name
		}

		response = append(response, TournamentResponse{
			ID:      tournament.ID,
			Name:    tournament.Name,
			Rounds:  tournament.Rounds,
			Winner:  winnerName,
			Created: tournament.Created.Format("2006-01-02T15:04:05Z07:00"),
		})
	}

	json.NewEncoder(w).Encode(response)
}

// Calcutta API handlers

func calcuttasHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	calcuttas, err := calcuttaRepo.GetAll(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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

func main() {
	// Define routes
	http.HandleFunc("/api/health", enableCORS(healthHandler))
	http.HandleFunc("/api/schools", enableCORS(schoolsHandler))
	http.HandleFunc("/api/tournaments", enableCORS(tournamentsHandler))
	http.HandleFunc("/api/calcuttas", enableCORS(calcuttasHandler))
	http.HandleFunc("/api/calcuttas/", func(w http.ResponseWriter, r *http.Request) {
		pathParts := strings.Split(r.URL.Path, "/")
		if len(pathParts) >= 6 && pathParts[4] == "entries" && pathParts[6] == "teams" {
			enableCORS(calcuttaEntryTeamHandler)(w, r)
		} else if len(pathParts) >= 4 {
			enableCORS(calcuttaEntriesHandler)(w, r)
		} else {
			http.NotFound(w, r)
		}
	})
	http.HandleFunc("/api/entries/", enableCORS(entryTeamsHandler))

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
