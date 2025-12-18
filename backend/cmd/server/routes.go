package main

import "github.com/gorilla/mux"

// RegisterRoutes centralizes HTTP route registration
func RegisterRoutes(r *mux.Router) {
	// Health & basic
	r.HandleFunc("/api/health", healthHandler).Methods("GET")
	r.HandleFunc("/api/schools", schoolsHandler).Methods("GET")

	// Auth
	r.HandleFunc("/api/auth/login", loginHandler).Methods("POST")
	r.HandleFunc("/api/auth/signup", signupHandler).Methods("POST")

	// Tournaments
	r.HandleFunc("/api/tournaments", tournamentsHandler).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}", tournamentHandler).Methods("GET")
	r.HandleFunc("/api/tournaments", createTournamentHandler).Methods("POST")
	r.HandleFunc("/api/tournaments/{id}/teams", tournamentTeamsHandler).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}/teams", createTournamentTeamHandler).Methods("POST")
	r.HandleFunc("/api/teams/{id}", updateTeamHandler).Methods("PATCH")
	r.HandleFunc("/api/tournaments/{id}/recalculate-portfolios", recalculatePortfoliosHandler).Methods("POST")

	// Portfolio scoring
	r.HandleFunc("/api/portfolios/{id}/calculate-scores", calculatePortfolioScoresHandler).Methods("POST")
	r.HandleFunc("/api/portfolios/{id}/teams/{teamId}/scores", updatePortfolioTeamScoresHandler).Methods("PUT")
	r.HandleFunc("/api/portfolios/{id}/maximum-score", updatePortfolioMaximumScoreHandler).Methods("PUT")

	// Calcutta
	r.HandleFunc("/api/calcuttas", calcuttasHandler).Methods("GET", "POST")
	r.HandleFunc("/api/calcuttas/{id}", calcuttaHandler).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}/entries", calcuttaEntriesHandler).Methods("GET")
	r.HandleFunc("/api/calcuttas/{calcuttaId}/entries/{entryId}/teams", calcuttaEntryTeamHandler).Methods("GET")
	r.HandleFunc("/api/entries/{id}/teams", entryTeamsHandler)
	r.HandleFunc("/api/entries/{id}/portfolios", portfoliosHandler)
	r.HandleFunc("/api/portfolios/{id}/teams", portfolioTeamsHandler)
}
