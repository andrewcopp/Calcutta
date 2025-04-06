package main

import (
	"database/sql"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/andrewcopp/Calcutta/backend/pkg/models"
	"github.com/andrewcopp/Calcutta/backend/pkg/services"
	"github.com/gorilla/mux"
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
	calcuttaService = services.NewCalcuttaService(calcuttaRepo)
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

func tournamentTeamsHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers explicitly for this handler
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	// Extract tournament ID from URL path
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		http.Error(w, "Tournament ID is required", http.StatusBadRequest)
		return
	}

	// Get teams for the tournament
	teams, err := tournamentRepo.GetTeams(r.Context(), tournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Get school information for each team
	var response []map[string]interface{}
	for _, team := range teams {
		// Get the school
		school, err := schoolService.GetSchoolByID(r.Context(), team.SchoolID)
		if err != nil {
			log.Printf("Error getting school for team %s: %v", team.ID, err)
			continue
		}

		// Create a response object with team and school information
		teamResponse := map[string]interface{}{
			"id":         team.ID,
			"schoolId":   team.SchoolID,
			"seed":       team.Seed,
			"byes":       team.Byes,
			"wins":       team.Wins,
			"eliminated": team.Eliminated,
			"created":    team.Created.Format("2006-01-02T15:04:05Z07:00"),
			"updated":    team.Updated.Format("2006-01-02T15:04:05Z07:00"),
			"school": map[string]interface{}{
				"id":   school.ID,
				"name": school.Name,
			},
		}

		response = append(response, teamResponse)
	}

	json.NewEncoder(w).Encode(response)
}

func updateTeamHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers explicitly for this handler
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Access-Control-Max-Age", "3600")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract team ID from URL path
	vars := mux.Vars(r)
	teamID := vars["id"]
	if teamID == "" {
		http.Error(w, "Team ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var request struct {
		Wins       *int  `json:"wins,omitempty"`
		Byes       *int  `json:"byes,omitempty"`
		Eliminated *bool `json:"eliminated,omitempty"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Get the team
	team, err := calcuttaRepo.GetTournamentTeam(r.Context(), teamID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Update the team fields
	if request.Wins != nil {
		team.Wins = *request.Wins
	}
	if request.Byes != nil {
		team.Byes = *request.Byes
	}
	if request.Eliminated != nil {
		team.Eliminated = *request.Eliminated
	}

	// Validate the team
	if err := team.ValidateDefault(); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Update the team in the database
	err = tournamentRepo.UpdateTournamentTeam(r.Context(), team)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Team updated successfully",
	})
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

func setupRoutes(r *mux.Router, calcuttaService *services.CalcuttaService) {
	// Portfolio routes
	r.HandleFunc("/api/portfolios/{id}/calculate-scores", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portfolioID := vars["id"]

		err := calcuttaService.CalculatePortfolioScores(r.Context(), portfolioID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}).Methods("POST")

	r.HandleFunc("/api/portfolios/{id}/teams/{teamId}/scores", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portfolioID := vars["id"]
		teamID := vars["teamId"]

		var request struct {
			ExpectedPoints  float64 `json:"expectedPoints"`
			PredictedPoints float64 `json:"predictedPoints"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		teams, err := calcuttaService.GetPortfolioTeams(r.Context(), portfolioID)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var portfolioTeam *models.CalcuttaPortfolioTeam
		for _, team := range teams {
			if team.TeamID == teamID {
				portfolioTeam = team
				break
			}
		}

		if portfolioTeam == nil {
			http.Error(w, "Portfolio team not found", http.StatusNotFound)
			return
		}

		portfolioTeam.ExpectedPoints = request.ExpectedPoints
		portfolioTeam.PredictedPoints = request.PredictedPoints
		portfolioTeam.Updated = time.Now()

		err = calcuttaService.UpdatePortfolioTeam(r.Context(), portfolioTeam)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}).Methods("PUT")

	r.HandleFunc("/api/portfolios/{id}/maximum-score", func(w http.ResponseWriter, r *http.Request) {
		vars := mux.Vars(r)
		portfolioID := vars["id"]

		var request struct {
			MaximumPoints float64 `json:"maximumPoints"`
		}

		if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		err := calcuttaService.UpdatePortfolioScores(r.Context(), portfolioID, request.MaximumPoints)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}).Methods("PUT")
}

func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Set CORS headers
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		w.Header().Set("Access-Control-Max-Age", "3600")

		// Handle preflight requests
		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}

func main() {
	r := mux.NewRouter()

	// Apply CORS middleware to all routes
	r.Use(corsMiddleware)

	// Routes
	r.HandleFunc("/api/health", healthHandler)
	r.HandleFunc("/api/schools", schoolsHandler)
	r.HandleFunc("/api/tournaments", tournamentsHandler)
	r.HandleFunc("/api/tournaments/{id}/teams", tournamentTeamsHandler)
	r.HandleFunc("/api/teams/{id}", updateTeamHandler).Methods("PATCH", "OPTIONS")
	r.HandleFunc("/api/calcuttas", calcuttasHandler)
	r.HandleFunc("/api/calcuttas/{id}/entries", calcuttaEntriesHandler)
	r.HandleFunc("/api/calcuttas/{id}/entries/{entryId}/teams", calcuttaEntryTeamHandler)
	r.HandleFunc("/api/entries/{id}/teams", entryTeamsHandler)
	r.HandleFunc("/api/entries/{id}/portfolios", portfoliosHandler)
	r.HandleFunc("/api/portfolios/{id}/teams", portfolioTeamsHandler)

	setupRoutes(r, calcuttaService)

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}
