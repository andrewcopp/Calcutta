package main

import "github.com/gorilla/mux"

// RegisterRoutes centralizes HTTP route registration
func (s *Server) RegisterRoutes(r *mux.Router) {
	s.registerBasicRoutes(r)
	s.registerAuthRoutes(r)
	s.registerTournamentRoutes(r)
	s.registerPortfolioRoutes(r)
	s.registerCalcuttaRoutes(r)
}

func (s *Server) registerBasicRoutes(r *mux.Router) {
	// Health & basic
	r.HandleFunc("/api/health", s.healthHandler).Methods("GET")
	r.HandleFunc("/api/schools", s.schoolsHandler).Methods("GET")
}

func (s *Server) registerAuthRoutes(r *mux.Router) {
	// Auth
	r.HandleFunc("/api/auth/login", s.loginHandler).Methods("POST")
	r.HandleFunc("/api/auth/signup", s.signupHandler).Methods("POST")
}

func (s *Server) registerTournamentRoutes(r *mux.Router) {
	// Tournaments
	r.HandleFunc("/api/tournaments", s.tournamentsHandler).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}", s.tournamentHandler).Methods("GET")
	r.HandleFunc("/api/tournaments", s.createTournamentHandler).Methods("POST")
	r.HandleFunc("/api/tournaments/{id}/teams", s.tournamentTeamsHandler).Methods("GET")
	r.HandleFunc("/api/tournaments/{id}/teams", s.createTournamentTeamHandler).Methods("POST")
	r.HandleFunc("/api/teams/{id}", s.updateTeamHandler).Methods("PATCH")
	r.HandleFunc("/api/tournaments/{id}/recalculate-portfolios", s.recalculatePortfoliosHandler).Methods("POST")
}

func (s *Server) registerPortfolioRoutes(r *mux.Router) {
	// Portfolio scoring
	r.HandleFunc("/api/portfolios/{id}/calculate-scores", s.calculatePortfolioScoresHandler).Methods("POST")
	r.HandleFunc("/api/portfolios/{id}/teams/{teamId}/scores", s.updatePortfolioTeamScoresHandler).Methods("PUT")
	r.HandleFunc("/api/portfolios/{id}/maximum-score", s.updatePortfolioMaximumScoreHandler).Methods("PUT")
	r.HandleFunc("/api/portfolios/{id}/teams", s.portfolioTeamsHandler).Methods("GET")
}

func (s *Server) registerCalcuttaRoutes(r *mux.Router) {
	// Calcutta
	r.HandleFunc("/api/calcuttas", s.calcuttasHandler).Methods("GET", "POST")
	r.HandleFunc("/api/calcuttas/{id}", s.calcuttaHandler).Methods("GET")
	r.HandleFunc("/api/calcuttas/{id}/entries", s.calcuttaEntriesHandler).Methods("GET")
	r.HandleFunc("/api/calcuttas/{calcuttaId}/entries/{entryId}/teams", s.calcuttaEntryTeamHandler).Methods("GET")
	r.HandleFunc("/api/entries/{id}/teams", s.entryTeamsHandler).Methods("GET")
	r.HandleFunc("/api/entries/{id}/portfolios", s.portfoliosHandler).Methods("GET")
}
