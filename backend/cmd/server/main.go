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
	"github.com/google/uuid"
	"github.com/gorilla/mux"
	_ "github.com/lib/pq"
)

var schoolRepo *services.SchoolRepository
var schoolService *services.SchoolService
var tournamentRepo *services.TournamentRepository
var tournamentService *services.TournamentService
var calcuttaRepo *services.CalcuttaRepository
var calcuttaService *services.CalcuttaService
var userRepo *services.UserRepository
var userService *services.UserService

func init() {
	log.Printf("Initializing database connection...")

	// Get database connection string from environment
	connString := os.Getenv("DATABASE_URL")
	if connString == "" {
		log.Fatal("DATABASE_URL environment variable is not set")
	}
	log.Printf("Database connection string: %s", connString)

	// Connect to database
	db, err := sql.Open("postgres", connString)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatalf("Failed to ping database: %v", err)
	}
	log.Printf("Successfully connected to database")

	// Initialize repositories
	log.Printf("Initializing repositories...")
	schoolRepo = services.NewSchoolRepository(db)
	tournamentRepo = services.NewTournamentRepository(db)
	calcuttaRepo = services.NewCalcuttaRepository(db)
	userRepo = services.NewUserRepository(db)
	log.Printf("Repositories initialized successfully")

	// Initialize services
	log.Printf("Initializing services...")
	schoolService = services.NewSchoolService(schoolRepo)
	tournamentService = services.NewTournamentService(tournamentRepo)
	calcuttaService = services.NewCalcuttaService(calcuttaRepo)
	userService = services.NewUserService(userRepo)
	log.Printf("Services initialized successfully")
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

		// Log tournament data
		log.Printf("Processing tournament: ID=%s, Name=%s", tournament.ID, tournament.Name)

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

	// Get the updated team with school information
	updatedTeam, err := tournamentRepo.GetTeams(r.Context(), team.TournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Find the updated team in the list
	var responseTeam *models.TournamentTeam
	for _, t := range updatedTeam {
		if t.ID == team.ID {
			responseTeam = t
			break
		}
	}

	if responseTeam == nil {
		http.Error(w, "Failed to retrieve updated team", http.StatusInternalServerError)
		return
	}

	// Get the school information
	school, err := schoolService.GetSchoolByID(r.Context(), responseTeam.SchoolID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Create response with team and school information
	response := map[string]interface{}{
		"id":         responseTeam.ID,
		"schoolId":   responseTeam.SchoolID,
		"seed":       responseTeam.Seed,
		"byes":       responseTeam.Byes,
		"wins":       responseTeam.Wins,
		"eliminated": responseTeam.Eliminated,
		"created":    responseTeam.Created.Format("2006-01-02T15:04:05Z07:00"),
		"updated":    responseTeam.Updated.Format("2006-01-02T15:04:05Z07:00"),
		"school": map[string]interface{}{
			"id":   school.ID,
			"name": school.Name,
		},
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

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

func setupCORS(router *mux.Router) {
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Set CORS headers
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
			w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

			// Handle preflight requests
			if r.Method == "OPTIONS" {
				w.WriteHeader(http.StatusOK)
				return
			}

			next.ServeHTTP(w, r)
		})
	})
}

func setupRoutes(r *mux.Router, calcuttaService *services.CalcuttaService) {
	// Health check
	r.HandleFunc("/health", healthHandler).Methods("GET", "OPTIONS")

	// Auth routes - add OPTIONS method for CORS preflight
	r.HandleFunc("/api/auth/login", loginHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/auth/signup", signupHandler).Methods("POST", "OPTIONS")

	// Schools
	r.HandleFunc("/api/schools", schoolsHandler).Methods("GET", "OPTIONS")

	// ... rest of the routes ...

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

	// Calcutta routes
	r.HandleFunc("/api/calcuttas", calcuttasHandler).Methods("GET", "POST")
	r.HandleFunc("/api/calcuttas/{id}", calcuttaHandler).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}/entries", calcuttaEntriesHandler).Methods("GET")
	r.HandleFunc("/api/calcuttas/{calcuttaId}/entries/{entryId}/teams", calcuttaEntryTeamHandler).Methods("GET")
	r.HandleFunc("/api/entries/{id}/teams", entryTeamsHandler)
	r.HandleFunc("/api/entries/{id}/portfolios", portfoliosHandler)
	r.HandleFunc("/api/portfolios/{id}/teams", portfolioTeamsHandler)
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

func recalculatePortfoliosHandler(w http.ResponseWriter, r *http.Request) {
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

	// Extract tournament ID from URL path
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		http.Error(w, "Tournament ID is required", http.StatusBadRequest)
		return
	}

	// Get all calcuttas for this tournament
	calcuttas, err := calcuttaRepo.GetCalcuttasByTournament(r.Context(), tournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Recalculate portfolios for each calcutta
	for _, calcutta := range calcuttas {
		if err := calcuttaService.RecalculatePortfolio(r.Context(), calcutta.ID); err != nil {
			log.Printf("Error recalculating portfolio for calcutta %s: %v", calcutta.ID, err)
			continue
		}
	}

	w.WriteHeader(http.StatusOK)
}

func createTournamentHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Parse request body
	var request struct {
		Name   string `json:"name"`
		Rounds int    `json:"rounds"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if request.Name == "" {
		http.Error(w, "Name is required", http.StatusBadRequest)
		return
	}
	if request.Rounds <= 0 {
		http.Error(w, "Rounds must be greater than 0", http.StatusBadRequest)
		return
	}

	// Create tournament
	tournament, err := tournamentService.CreateTournament(r.Context(), request.Name, request.Rounds)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created tournament
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(tournament)
}

func createTournamentTeamHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
	w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
	w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, PATCH, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Accept, Authorization")
	w.Header().Set("Access-Control-Allow-Credentials", "true")
	w.Header().Set("Content-Type", "application/json")

	// Handle preflight requests
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	// Extract tournament ID from URL path
	vars := mux.Vars(r)
	tournamentID := vars["id"]
	if tournamentID == "" {
		http.Error(w, "Tournament ID is required", http.StatusBadRequest)
		return
	}

	// Parse request body
	var request struct {
		SchoolID string `json:"schoolId"`
		Seed     int    `json:"seed"`
		Region   string `json:"region"`
	}

	if err := json.NewDecoder(r.Body).Decode(&request); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Validate request
	if request.SchoolID == "" {
		http.Error(w, "School ID is required", http.StatusBadRequest)
		return
	}
	if request.Seed < 1 || request.Seed > 16 {
		http.Error(w, "Seed must be between 1 and 16", http.StatusBadRequest)
		return
	}
	if request.Region == "" {
		request.Region = "Unknown" // Default to "Unknown" if not provided
	}

	// Create the team
	team := &models.TournamentTeam{
		ID:           uuid.New().String(),
		TournamentID: tournamentID,
		SchoolID:     request.SchoolID,
		Seed:         request.Seed,
		Region:       request.Region,
		Byes:         0,
		Wins:         0,
		Eliminated:   false,
	}

	if err := tournamentRepo.CreateTeam(r.Context(), team); err != nil {
		log.Printf("Error creating team: %v", err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the created team
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(team)
}

func tournamentHandler(w http.ResponseWriter, r *http.Request) {
	// Set CORS headers
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

	// Get tournament by ID
	tournament, err := tournamentService.GetTournamentByID(r.Context(), tournamentID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if tournament == nil {
		http.Error(w, "Tournament not found", http.StatusNotFound)
		return
	}

	// Get the winning team for this tournament
	team, err := tournamentService.GetWinningTeam(r.Context(), tournament.ID)
	if err != nil {
		log.Printf("Error getting winning team for tournament %s: %v", tournament.ID, err)
	}

	// Create response with tournament and winning team info
	type TournamentResponse struct {
		ID      string `json:"id"`
		Name    string `json:"name"`
		Rounds  int    `json:"rounds"`
		Winner  string `json:"winner,omitempty"`
		Created string `json:"created"`
	}

	winnerName := ""
	if team != nil {
		// Get the school name
		school, err := schoolService.GetSchoolByID(r.Context(), team.SchoolID)
		if err != nil {
			log.Printf("Error getting school for team %s: %v", team.ID, err)
		} else {
			winnerName = school.Name
		}
	}

	response := TournamentResponse{
		ID:      tournament.ID,
		Name:    tournament.Name,
		Rounds:  tournament.Rounds,
		Winner:  winnerName,
		Created: tournament.Created.Format("2006-01-02T15:04:05Z07:00"),
	}

	// Return the tournament response
	json.NewEncoder(w).Encode(response)
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

func loginHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Email string `json:"email"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := userService.Login(r.Context(), req.Email)
	if err != nil {
		http.Error(w, err.Error(), http.StatusUnauthorized)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func signupHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	var req struct {
		Email     string `json:"email"`
		FirstName string `json:"firstName"`
		LastName  string `json:"lastName"`
	}

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	user, err := userService.Signup(r.Context(), req.Email, req.FirstName, req.LastName)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	json.NewEncoder(w).Encode(user)
}

func main() {
	log.Printf("Starting server initialization...")

	r := mux.NewRouter()
	log.Printf("Router created")

	// Apply CORS middleware to all routes
	r.Use(corsMiddleware)
	log.Printf("CORS middleware applied")

	// Routes
	log.Printf("Setting up routes...")
	r.HandleFunc("/api/health", healthHandler)
	r.HandleFunc("/api/schools", schoolsHandler)
	r.HandleFunc("/api/tournaments", tournamentsHandler).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}", tournamentHandler).Methods("GET")
	r.HandleFunc("/api/tournaments", createTournamentHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/tournaments/{id}/teams", tournamentTeamsHandler).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}/teams", createTournamentTeamHandler).Methods("POST", "OPTIONS")
	r.HandleFunc("/api/teams/{id}", updateTeamHandler).Methods("PATCH", "OPTIONS")
	r.HandleFunc("/api/tournaments/{id}/recalculate-portfolios", recalculatePortfoliosHandler).Methods("POST", "OPTIONS")
	log.Printf("Basic routes configured")

	// Add debug logging for unmatched routes
	r.NotFoundHandler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("No route matched for request: %s %s", r.Method, r.URL.Path)
		http.Error(w, "Not Found", http.StatusNotFound)
	})

	setupRoutes(r, calcuttaService)
	log.Printf("Additional routes configured")

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Using port: %s", port)

	log.Printf("Server starting on port %s", port)
	if err := http.ListenAndServe(":"+port, r); err != nil {
		log.Printf("Server failed to start: %v", err)
		log.Fatalf("Server failed to start: %v", err)
	}
}
